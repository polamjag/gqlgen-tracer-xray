package gqlgenxraytracer

// heavily inspired from gqlgen-contrib/gqlopentracing
// https://github.com/99designs/gqlgen-contrib/tree/master/gqlopentracing

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aereal/gqlgen-tracer-xray/log"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func New() tracer {
	return tracer(1)
}

type tracer int

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.FieldInterceptor
	graphql.ResponseInterceptor
} = tracer(1)

func (t tracer) ExtensionName() string {
	return "XrayTracer"
}

func (t tracer) Validate(es graphql.ExecutableSchema) error {
	return nil
}

func (t tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	log.Logger.Printf("start build response")
	startedAt := time.Now()
	res := next(ctx)
	finishedAt := time.Now()
	log.Logger.Printf("finish build response; elapsed:%s", finishedAt.Sub(startedAt))
	return res
}

func (t tracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)
	opName := oc.OperationName
	if opName == "" {
		opName = "(unnamed)"
	}
	subCtx, seg := xray.BeginSubsegment(ctx, fmt.Sprintf("gql op %s", opName))
	if seg == nil {
		log.Logger.Printf("InterceptOperation: no parent segment")
		return next(ctx)
	}
	finish := func() {
		if seg.InProgress {
			log.Logger.Println("finished at defer")
			seg.Close(nil)
			log.Close(seg)
		}
	}
	defer finish()
	seg.AddMetadata("gql.variables", oc.Variables)
	log.Start(seg)
	// TODO: complexity

	h := next(subCtx)
	finish()
	return h
}

func (t tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	parentSubSeg := xray.GetSegment(ctx)
	log.Logger.Printf("InterceptField: parent subsegment name=%q %#v", parentSubSeg.Name, parentSubSeg)

	fc := graphql.GetFieldContext(ctx)

	subCtx, seg := xray.BeginSubsegment(ctx, fmt.Sprintf("gql field %s", fc.Field.Name))
	if seg == nil {
		log.Logger.Printf("InterceptField: no parent segment")
		return next(ctx)
	}
	finish := func() {
		if seg.InProgress {
			log.Logger.Println("finished at defer")
			seg.Close(nil)
			log.Close(seg)
		}
	}
	defer finish()
	seg.AddMetadata("gql.object", fc.Object)
	seg.AddMetadata("gql.field", fc.Field.Name)
	log.Start(seg)

	errs := graphql.GetFieldErrors(subCtx, fc)
	for _, err := range errs {
		seg.AddError(err)
	}

	res, err := next(subCtx)
	finish()
	return res, err
}
