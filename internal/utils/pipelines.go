package utils

/*
 * File: internal/utils/pipelines.go
 *
 * Purpose: utilities for constructing and running goroutine based pipeline operations
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"context"
	"errors"
)

// Type `StepFn` represents a processing step within a pipeline. It is connectable by an incoming step whose output is type `I` and an outgoing step whose input is type `O`
//
// Type Parameters
//   - I: the input type, typically coming from another `StepFn` whose output is of type `I`
//   - O: the outpt type, typically matching another `StepFn` whose input is of type `O`
//
// Parameters:
//   - I: the incoming value to process
//
// Returns:
//   - O: the value to pass off to another step
//   - error: issue that occurred specifically in this step
type StepFn[I any, O any] func(I) (O, error)

// Function `Stage` defines a processing step in a pipeline. Internal goroutine immediately starts and blocks until an input value is ready to process
//
// Type Parameters:
//   - I: the type of value the input channel yields
//   - O: the type of value to push to the output channel
//
// Parameters:
//   - ctx: the context managing the lifetime of this pipeline
//   - cancelFunc: the callable function to report the error that occurs (if any) to the managing context (cancels context and forces all steps to exit)
//   - f: the processing step to call within the internal goroutine to "do work"
//   - in: the input channel (read only) to "do work" on
//
// Returns:
//   - `<-chan O`: the output channel (read only) to pass on as an input channel to the following stage
func Stage[I any, O any](ctx context.Context, cancelFunc context.CancelCauseFunc, f StepFn[I, O], in <-chan I) <-chan O {
	out := make(chan O)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-in:
				if !ok {
					cancelFunc(errors.New("broken pipe"))
					return
				}
				if next, err := f(item); err != nil {
					cancelFunc(err)
					return
				} else {
					out <- next
				}
			}
		}
	}()

	return out
}

// Function `Feed` creates a channel from the given values to act as an input to the first stage of a pipeline. Effectively a start button
//
// Type Parameters:
//   - T: the type of values to feed to the first stage of the pipeline
//
// Parameters:
//   - ctx: the context managing the lifetime of this pipeline
//   - values...: the sequence of values to process
//
// Returns:
//   - `<-chan T`: an output channel to pass to the first stage of a pipeline and kick-off work
func Feed[T any](ctx context.Context, values ...T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			default:
				out <- v
			}
		}
	}()

	return out
}
