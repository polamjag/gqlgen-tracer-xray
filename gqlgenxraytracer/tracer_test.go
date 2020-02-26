package gqlgenxraytracer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/aereal/gqlgen-tracer-xray/testdata"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func TestTracer(t *testing.T) {
	xray.Configure(xray.Config{
		ContextMissingStrategy: nopMissingStrategy(0),
	})
	specs := []struct {
		Name             string
		Tracer           graphql.Tracer
		ExpectedSegments []*xray.Segment
		Body             string
	}{
		{
			Name:   "ok",
			Tracer: New(),
			ExpectedSegments: []*xray.Segment{
				&xray.Segment{
					Name: "gql op (unnamed)",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.complexity": 0,
							"gql.variables":  map[string]interface{}{},
						},
					}},
				&xray.Segment{
					Name: "gql field visitor",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.field":  "visitor",
							"gql.object": "Query",
						},
					}},
				&xray.Segment{
					Name: "gql field name",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.field":  "name",
							"gql.object": "User",
						},
					}},
			},
			Body: `{"query":"{ visitor { name } }"}`,
		},
		{
			Name:   "with name",
			Tracer: New(),
			ExpectedSegments: []*xray.Segment{
				&xray.Segment{
					Name: "gql op GetVisitorName",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.complexity": 0,
							"gql.variables":  map[string]interface{}{},
						},
					}},
				&xray.Segment{
					Name: "gql field visitor",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.field":  "visitor",
							"gql.object": "Query",
						},
					}},
				&xray.Segment{
					Name: "gql field name",
					Metadata: map[string]map[string]interface{}{
						"default": map[string]interface{}{
							"gql.field":  "name",
							"gql.object": "User",
						},
					}},
			},
			Body: `{"query":"query GetVisitorName { visitor { name } }","operationName":"GetVisitorName"}`,
		},
	}
	for _, spec := range specs {
		t.Run(spec.Name, func(t *testing.T) {
			h := handler.GraphQL(testdata.NewExecutableSchema(testdata.Config{
				Resolvers: &testdata.Resolver{},
			}), handler.Tracer(spec.Tracer))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx, seg := xray.BeginSegment(ctx, "test")
			defer seg.Close(nil)

			resp := doRequest(ctx, h, http.MethodPost, "/graphql", spec.Body)
			seg.Close(nil)
			if resp.Code != 200 {
				t.Error("request failed")
				return
			}

			gotSeg := xray.GetSegment(ctx)
			subSegs := drainSegments(gotSeg)
			if len(subSegs) != len(spec.ExpectedSegments) {
				t.Errorf("expected %d sub-segments but got %d", len(spec.ExpectedSegments), len(subSegs))
				return
			}
			for i, expectedSubSeg := range spec.ExpectedSegments {
				gotSubSeg := subSegs[i]
				normalizeSegment(gotSubSeg)
				normalizeSegment(expectedSubSeg)
				got := encodeJSON(t, gotSubSeg)
				expected := encodeJSON(t, expectedSubSeg)
				if string(got) != string(expected) {
					t.Errorf("[#%02d] expected segment:%s but got:%s", i, string(expected), string(got))
					return
				}
			}
		})
	}
}

func normalizeSegment(seg *xray.Segment) {
	seg.ID = ""
	seg.StartTime = 0
	seg.EndTime = 0
	seg.Subsegments = []json.RawMessage{}
}

func encodeJSON(t *testing.T, seg *xray.Segment) []byte {
	bytes, err := json.Marshal(seg)
	if err != nil {
		t.Errorf("failed to encode segment to JSON: %v", err)
		return nil
	}
	return bytes
}

func drainSegments(root *xray.Segment) []*xray.Segment {
	segs := []*xray.Segment{}
	if root == nil {
		return segs
	}
	for _, raw := range root.Subsegments {
		var s *xray.Segment
		err := json.Unmarshal(raw, &s)
		if err != nil {
			panic(err)
		}
		segs = append(segs, s)
		if len(s.Subsegments) > 0 {
			xs := drainSegments(s)
			segs = append(segs, xs...)
		}
	}
	return segs
}

func doRequest(ctx context.Context, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r = r.WithContext(ctx)
	r.Header.Set("content-type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

type nopMissingStrategy int

func (_ nopMissingStrategy) ContextMissing(v interface{}) {}
