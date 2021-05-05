package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"plugin"
)

type Resolver struct {
	plugins    []*plugin.Plugin
	middleware []MiddlewareFunc
	resolvers  map[string]HandlerFunc
}

func NewResolver() *Resolver {
	return &Resolver{resolvers: make(map[string]HandlerFunc)}
}

// Use adds middleware to the chain which is run after router.
func (r *Resolver) ResolveFunc(resolver string, handlerFunc HandlerFunc) {
	fmt.Printf("Loaded Resolver for %s\n", resolver)
	r.resolvers[resolver] = handlerFunc
}

// Use adds middleware to the chain which is run after router.
func (r *Resolver) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Resolver) Resolve(ctx context.Context, dbody *DBody) ([]byte, error) {
	args, err := json.Marshal(dbody.Args)
	if err != nil {
		fmt.Println("Could not marshal arguments")
	}

	if r.resolvers[dbody.Resolver] == nil {
		return nil, errors.New(fmt.Sprintf("Could not resolve %s", dbody.Resolver))
	}

	h := r.resolvers[dbody.Resolver]

	h = applyMiddleware(h, r.middleware...)

	res, err := h(ctx, args, dbody.AuthHeader)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
