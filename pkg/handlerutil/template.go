package handlerutil

/*
 * File: pkg/handlerutil/template.go
 *
 * Purpose: declaration of a higher level function that dictates the generic flow of a HTTP handlerfunc
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
)

// Type `WorkspaceInit` is a function pointer that returns a pointer to an initialized HandlerWorkspace
type WorkspaceInit func() *HandlerWorkspace

// Type `WorkflowStarter` is a function pointer that sets up a handler workflow pipeline a returns the control surfaces to it
type WorkflowStarter func() (context.Context, context.CancelCauseFunc, chan<- *HandlerWorkspace, <-chan *HandlerWorkspace)

// Type `WorkflowWaiter` is a function pointer that awaits the pipeline exit signals and processes the result
type WorkflowWaiter func(context.Context, *gin.Context, <-chan *HandlerWorkspace, int, *handlerFailureFormatter)

// Type `TransitionFn` represents a processing step within a pipeline that works with a given `HandlerWorkspace`.
// It should work with the workspace in-place and return an error to indicate failure
type TransitionFn func(*HandlerWorkspace) error

// Predefined errors that can emerge when requesting keys from a handler workspace
var (
	ErrKeyNotExists   = errors.New("the requested key does not exist")
	ErrNotAddressable = errors.New("value must be addressable")
	ErrNotAssignable  = errors.New("value cannot be assigned")
)

// Type `HandlerWorkspace` contains a synced key/value store for handlers to use as a scratchpad when processing requests
//
// Members:
//   - mu: the synchronization primitive for data access
//   - data: the key/value store
type HandlerWorkspace struct {
	mu   sync.RWMutex
	data map[string]any
}

// Function `DefaultWorkspace` creates a sane zero value workspace for immediate use
//
// Returns:
//   - `HandlerWorkspace`: the zero value workspace
func DefaultWorkspace() HandlerWorkspace {
	return HandlerWorkspace{
		mu:   sync.RWMutex{},
		data: make(map[string]any),
	}
}

// Function `(*HandlerWorkspace).Get` looks up the given key and copies the existing value to the given addressable value
// This function reads the value from the internal data map as-is and creates a copy of it for external mutation.
// Mutated versions must be updated with the (*HandlerWorkspace).Set() method for changes to persist with later accesses
//
// Parameters:
//   - key: the key to lookup
//   - out: the addressable location to copy the existing value associated with key
//
// Returns:
//   - `error`: issue that occurred during lookup or copy (nil if no issue occurred)
func (s *HandlerWorkspace) Get(key string, out any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return ErrKeyNotExists
	}

	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Pointer || outVal.IsNil() {
		return ErrNotAddressable
	}

	valVal := reflect.ValueOf(val)
	if !valVal.Type().AssignableTo(outVal.Type().Elem()) {
		return ErrNotAssignable
	}

	outVal.Elem().Set(valVal)
	return nil
}

// Function `(*HandlerWorkspace).Set` set the given key/value pair in the workspace store
//
// Parameters:
//   - key: the key that will be set
//   - value: the value that will be associated with key
func (s *HandlerWorkspace) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Function `HandlerTemplate` describes a generalized structure for HTTP handler function utilizing pipelined execution
//
// Parameters:
//   - stateInitializer: a callable that initializes the state for the handler
//   - pipeline: a callable that starts the pipeline goroutines and returns control surfaces
//   - successCode: the HTTP status code that should be included with successful responses
//
// Returns:
//   - `gin.HandlerFunc`: the templated HTTP handler function
func HandlerTemplate(
	stateInitializer func() *HandlerWorkspace,
	pipeline func() (context.Context, context.CancelCauseFunc, chan<- *HandlerWorkspace, <-chan *HandlerWorkspace),
	await func(context.Context, *gin.Context, <-chan *HandlerWorkspace, int),
	successCode int,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel, in, out := pipeline()
		defer cancel(nil)
		defer close(in)

		in <- stateInitializer()
		await(ctx, c, out, successCode)
	}
}

// Function `stage` represents a step of work done by a `TransitionFn`. Internal goroutine immediately starts and blocks until an input value is ready to process
//
// Parameters:
//   - ctx: the context managing the lifetime of this pipeline
//   - cancelFunc: the callable function to report the error that occurs (if any) to the managing context (cancels context and forces all steps to exit)
//   - t: the processing step to call within the internal goroutine to "do work"
//   - in: the input channel (read only) to "do work" on
//
// Returns:
//   - `<-chan *HandlerWorkspace`: the output channel (read only) to pass on as an input channel to the following stage or to be read as the final result
func Stage(ctx context.Context, cancelFunc context.CancelCauseFunc, t TransitionFn, in <-chan *HandlerWorkspace) <-chan *HandlerWorkspace {
	out := make(chan *HandlerWorkspace)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-in:
				if !ok {
					// cancelFunc(errors.New("broken pipe"))
					return
				}
				if err := t(item); err != nil {
					cancelFunc(err)
					return
				} else {
					out <- item
				}
			}
		}
	}()

	return out
}
