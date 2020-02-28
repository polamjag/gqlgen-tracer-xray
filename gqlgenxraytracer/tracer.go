package gqlgenxraytracer

// heavily inspired from gqlgen-contrib/gqlopentracing
// https://github.com/99designs/gqlgen-contrib/tree/master/gqlopentracing

import (
	"context"
	"fmt"

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

func beginOperationSegment(ctx context.Context) (context.Context, *xray.Segment) {
	oc := graphql.GetOperationContext(ctx)
	opName := oc.OperationName
	if opName == "" {
		opName = "(unnamed)"
	}
	subCtx, seg := xray.BeginSubsegment(ctx, fmt.Sprintf("gql op %s", opName))
	return subCtx, seg
}

func (t tracer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	subCtx, seg := beginOperationSegment(ctx)
	if seg == nil {
		return next(ctx)
	}
	finish := func(atDefer bool) {
		if seg.InProgress {
			seg.Close(nil)
			log.Close(seg, atDefer)
		}
	}
	defer finish(true)
	oc := graphql.GetOperationContext(subCtx)
	seg.AddMetadata("gql.variables", oc.Variables)
	log.Start(seg)

	res := next(subCtx)
	finish(false)

	return res
}

func (t tracer) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// subCtx, seg := beginOperationSegment(ctx)
	// if seg == nil {
	// 	log.Logger.Printf("InterceptOperation: no parent segment")
	// 	return next(ctx)
	// }
	// finish := func() {
	// 	if seg.InProgress {
	// 		log.Logger.Println("finished at defer")
	// 		seg.Close(nil)
	// 		log.Close(seg)
	// 	}
	// }
	// defer finish()
	// oc := graphql.GetOperationContext(ctx)
	// seg.AddMetadata("gql.variables", oc.Variables)
	// log.Start(seg)
	// TODO: complexity

	return next(ctx)
	// h := next(subCtx)
	// finish()
	// return h
}

func (t tracer) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	// parentSubSeg := xray.GetSegment(ctx)
	// log.Logger.Printf("InterceptField: parent subsegment name=%q", parentSubSeg.Name)

	fc := graphql.GetFieldContext(ctx)

	subCtx, seg := xray.BeginSubsegment(ctx, fmt.Sprintf("gql field %s", fc.Field.Name))
	if seg == nil {
		return next(ctx)
	}
	finish := func(atDefer bool) {
		if seg.InProgress {
			seg.Close(nil)
			log.Close(seg, atDefer)
		}
	}
	defer finish(true)
	seg.AddMetadata("gql.object", fc.Object)
	seg.AddMetadata("gql.field", fc.Field.Name)
	log.Start(seg)

	errs := graphql.GetFieldErrors(subCtx, fc)
	for _, err := range errs {
		seg.AddError(err)
	}

	res, err := next(subCtx)
	finish(false)
	return res, err
}
