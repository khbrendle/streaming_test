package main

import "context"

type RequestContext struct {
	ID string
}

const requestContextKey string = "requestContext"

func NewRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, rc)
}

func FromRequestContext(ctx context.Context) (rc *RequestContext, ok bool) {
	rc, ok = ctx.Value(requestContextKey).(*RequestContext)
	return
}
