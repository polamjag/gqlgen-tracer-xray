//go:generate go run github.com/99designs/gqlgen

package testdata

import (
	"context"
)

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Visitor(ctx context.Context) (*User, error) {
	user := &User{Name: "testuser"}
	return user, nil
}

func (r *queryResolver) User(ctx context.Context, name string) (*User, error) {
	user := &User{Name: name}
	return user, nil
}
