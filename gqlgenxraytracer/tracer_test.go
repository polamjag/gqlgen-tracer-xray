package gqlgenxraytracer

import (
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
		Name   string
		Tracer graphql.Tracer
	}{
		{
			Name:   "ok",
			Tracer: New(),
		},
	}
	for _, spec := range specs {
		t.Run(spec.Name, func(t *testing.T) {
			h := handler.GraphQL(testdata.NewExecutableSchema(testdata.Config{
				Resolvers: &testdata.Resolver{},
			}), handler.Tracer(spec.Tracer))

			resp := doRequest(h, http.MethodPost, "/graphql", `{"query":"{ visitor { name } }"}`)
			if resp.Code != 200 {
				t.Error("request failed")
			}
			t.Logf("response body: %q", resp.Body)
		})
	}
}

func doRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("content-type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

type nopMissingStrategy int

func (_ nopMissingStrategy) ContextMissing(v interface{}) {}
