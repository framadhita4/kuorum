package mediator

import (
	"context"
	"fmt"
	"reflect"
)

// Command marks a struct as a CQRS Command.
type Command[TResult any] interface {
	IsCommand()
}

// Query marks a struct as a CQRS Query.
type Query[TResult any] interface {
	IsQuery()
}

// CommandHandler handles a specific Command.
type CommandHandler[TCommand Command[TResult], TResult any] interface {
	Handle(ctx context.Context, cmd TCommand) (TResult, error)
}

// QueryHandler handles a specific Query.
type QueryHandler[TQuery Query[TResult], TResult any] interface {
	Handle(ctx context.Context, q TQuery) (TResult, error)
}

// PipelineBehavior defines middleware intercepts around mediator dispatches.
type PipelineBehavior func(ctx context.Context, req any, next func(ctx context.Context, req any) (any, error)) (any, error)

// Mediator dispatches commands and queries.
type Mediator interface {
	Send(ctx context.Context, cmd any) (any, error)
	Query(ctx context.Context, query any) (any, error)
	RegisterCommandHandler(cmdType any, handler any)
	RegisterQueryHandler(queryType any, handler any)
	AddBehavior(behavior PipelineBehavior)
}

type InMemoryMediator struct {
	commandHandlers map[reflect.Type]func(ctx context.Context, cmd any) (any, error)
	queryHandlers   map[reflect.Type]func(ctx context.Context, query any) (any, error)
	behaviors       []PipelineBehavior
}

// NewInMemoryMediator creates a new in-memory Mediator instance.
func NewInMemoryMediator() *InMemoryMediator {
	return &InMemoryMediator{
		commandHandlers: make(map[reflect.Type]func(ctx context.Context, cmd any) (any, error)),
		queryHandlers:   make(map[reflect.Type]func(ctx context.Context, query any) (any, error)),
		behaviors:       make([]PipelineBehavior, 0),
	}
}

// AddBehavior appends a pipeline behavior to the mediator.
func (m *InMemoryMediator) AddBehavior(behavior PipelineBehavior) {
	m.behaviors = append(m.behaviors, behavior)
}

// RegisterCommandHandler registers a handler function or struct for a command type.
func (m *InMemoryMediator) RegisterCommandHandler(cmdType any, handler any) {
	t := reflect.TypeOf(cmdType)
	handlerVal := reflect.ValueOf(handler)
	method := handlerVal.MethodByName("Handle")
	if !method.IsValid() {
		panic(fmt.Sprintf("handler %T does not implement Handle method", handler))
	}

	m.commandHandlers[t] = func(ctx context.Context, cmd any) (any, error) {
		results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(cmd)})
		res := results[0].Interface()
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return res, err
	}
}

// RegisterQueryHandler registers a handler function or struct for a query type.
func (m *InMemoryMediator) RegisterQueryHandler(queryType any, handler any) {
	t := reflect.TypeOf(queryType)
	handlerVal := reflect.ValueOf(handler)
	method := handlerVal.MethodByName("Handle")
	if !method.IsValid() {
		panic(fmt.Sprintf("handler %T does not implement Handle method", handler))
	}

	m.queryHandlers[t] = func(ctx context.Context, query any) (any, error) {
		results := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(query)})
		res := results[0].Interface()
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		return res, err
	}
}

func (m *InMemoryMediator) executeWithPipeline(ctx context.Context, req any, coreHandler func(ctx context.Context, req any) (any, error)) (any, error) {
	handler := coreHandler
	for i := len(m.behaviors) - 1; i >= 0; i-- {
		b := m.behaviors[i]
		prev := handler
		handler = func(c context.Context, r any) (any, error) {
			return b(c, r, prev)
		}
	}
	return handler(ctx, req)
}

// Send dispatches a Command through the pipeline.
func (m *InMemoryMediator) Send(ctx context.Context, cmd any) (any, error) {
	t := reflect.TypeOf(cmd)
	h, ok := m.commandHandlers[t]
	if !ok {
		return nil, fmt.Errorf("no command handler registered for type %s", t.String())
	}
	return m.executeWithPipeline(ctx, cmd, h)
}

// Query dispatches a Query through the pipeline.
func (m *InMemoryMediator) Query(ctx context.Context, query any) (any, error) {
	t := reflect.TypeOf(query)
	h, ok := m.queryHandlers[t]
	if !ok {
		return nil, fmt.Errorf("no query handler registered for type %s", t.String())
	}
	return m.executeWithPipeline(ctx, query, h)
}
