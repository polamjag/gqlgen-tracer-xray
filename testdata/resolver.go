//go:generate go run github.com/99designs/gqlgen

package testdata

import (
	"context"

	"github.com/aereal/gqlgen-tracer-xray/log"
)

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	log.Logger.Printf("resolver Query")
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Visitor(ctx context.Context) (*User, error) {
	log.Logger.Printf("resolver Visitor")
	user := &User{Name: "testuser"}
	return user, nil
}

func (r *queryResolver) User(ctx context.Context, name string) (*User, error) {
	log.Logger.Printf("resolver User")
	user := &User{Name: name}
	return user, nil
}
