package gqlgenxraytracer

// heavily inspired from gqlgen-contrib/gqlopentracing
// https://github.com/99designs/gqlgen-contrib/tree/master/gqlopentracing

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func New() graphql.Tracer {
	return tracer(1)
}

type tracer int

func (t tracer) StartOperationExecution(ctx context.Context) context.Context {
	reqCtx := graphql.GetRequestContext(ctx)
	opName := reqCtx.OperationName
	if opName == "" {
		opName = "(unnamed)"
	}
	subCtx, seg := xray.BeginSubsegment(ctx, "gql op "+opName)
	if seg == nil {
		return ctx
	}
	seg.AddMetadata("gql.variables", reqCtx.Variables)
	seg.AddMetadata("gql.complexity", reqCtx.OperationComplexity)
	return subCtx
}

func (t tracer) StartFieldExecution(ctx context.Context, field graphql.CollectedField) context.Context {
	subCtx, _ := xray.BeginSubsegment(ctx, "gql field "+field.Name)
	if subCtx == nil {
		return ctx
	}
	return subCtx
}

func (t tracer) StartFieldResolverExecution(ctx context.Context, rc *graphql.ResolverContext) context.Context {
	seg := xray.GetSegment(ctx)
	if seg == nil {
		return ctx
	}
	seg.AddMetadata("gql.object", rc.Object)
	seg.AddMetadata("gql.field", rc.Field.Name)
	return ctx
}

func (t tracer) EndFieldExecution(ctx context.Context) {
	seg := xray.GetSegment(ctx)
	if seg == nil {
		return
	}
	defer seg.Close(nil)

	rc := graphql.GetResolverContext(ctx)
	reqCtx := graphql.GetRequestContext(ctx)

	errs := reqCtx.GetErrors(rc)
	for _, err := range errs {
		seg.AddError(err)
	}
}

func (t tracer) EndOperationExecution(ctx context.Context) {
	seg := xray.GetSegment(ctx)
	if seg == nil {
		return
	}
	defer seg.Close(nil)
}

func (t tracer) StartOperationParsing(ctx context.Context) context.Context {
	return ctx
}
func (t tracer) EndOperationParsing(ctx context.Context) {}
func (t tracer) StartOperationValidation(ctx context.Context) context.Context {
	return ctx
}
func (t tracer) EndOperationValidation(ctx context.Context) {}
func (t tracer) StartFieldChildExecution(ctx context.Context) context.Context {
	return ctx
}
