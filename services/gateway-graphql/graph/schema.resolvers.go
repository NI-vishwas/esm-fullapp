package graph

import (
	"context"
	"ems-platform/services/gateway-graphql/graph/model"
	"fmt"
)

// CreateEvent is the resolver for the createEvent field.
func (r *mutationResolver) CreateEvent(ctx context.Context, input model.NewEvent) (*model.Event, error) {
	panic(fmt.Errorf("not implemented: CreateEvent - createEvent"))
}

// UpdateEvent is the resolver for the updateEvent field.
func (r *mutationResolver) UpdateEvent(ctx context.Context, id string, input model.NewEvent) (*model.Event, error) {
	panic(fmt.Errorf("not implemented: UpdateEvent - updateEvent"))
}

// DeleteEvent is the resolver for the deleteEvent field.
func (r *mutationResolver) DeleteEvent(ctx context.Context, id string) (bool, error) {
	panic(fmt.Errorf("not implemented: DeleteEvent - deleteEvent"))
}

// Events is the resolver for the events field.
func (r *queryResolver) Events(ctx context.Context) ([]*model.Event, error) {
	// panic(fmt.Errorf("not implemented: Events - events"))
	return r.Resolver.Events(ctx)
}

// Event is the resolver for the event field.
func (r *queryResolver) Event(ctx context.Context, id string) (*model.Event, error) {
	panic(fmt.Errorf("not implemented: Event - event"))
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
