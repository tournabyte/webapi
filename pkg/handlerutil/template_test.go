package handlerutil_test

/*
 * File: pkg/handlerutil/template_test.go
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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tournabyte/webapi/pkg/handlerutil"
)

func TestWorkSpaceReadWrite(t *testing.T) {
	space := handlerutil.DefaultWorkspace()
	key := "SpecialNumber"
	space.Set(key, 5)

	var v int

	t.Run("OK", func(t *testing.T) {
		err := space.Get(key, &v)

		assert.NoError(t, err)
		assert.Equal(t, 5, v)
	})

	t.Run("MissingKey", func(t *testing.T) {
		err := space.Get(strings.Repeat(key, 3), &v)
		assert.ErrorIs(t, err, handlerutil.ErrKeyNotExists)
	})

	t.Run("NonAddressable", func(t *testing.T) {
		err := space.Get(key, v)
		assert.ErrorIs(t, err, handlerutil.ErrNotAddressable)
	})

	t.Run("NotAssignable", func(t *testing.T) {
		var s string
		err := space.Get(key, &s)
		assert.ErrorIs(t, err, handlerutil.ErrNotAssignable)
	})

	t.Run("MutationMustBeSet", func(t *testing.T) {
		var mut int
		var immut int
		err := space.Get(key, &mut)
		require.NoError(t, err)

		mut *= 2
		assert.Equal(t, 10, mut)

		err = space.Get(key, &immut)
		require.NoError(t, err)

		assert.NotEqual(t, mut, immut)
		space.Set(key, mut)
		err = space.Get(key, &immut)

		require.NoError(t, err)
		assert.Equal(t, mut, immut)
	})

}

func TestStepwiseProcessing(t *testing.T) {
	ErrNegativeInput := errors.New("Expected a positive input")
	doubleIt := func(ctx context.Context, ws *handlerutil.HandlerWorkspace) error {
		var i int
		if err := ws.Get("X", &i); err != nil {
			return err
		}
		i *= 2
		ws.Set("X", i)
		return nil
	}
	ensurePositive := func(ctx context.Context, ws *handlerutil.HandlerWorkspace) error {
		var i int
		if err := ws.Get("X", &i); err != nil {
			return err
		}
		if i <= 0 {
			return fmt.Errorf("%w: got %d", ErrNegativeInput, i)
		}
		return nil
	}

	workflow := func(ctx context.Context) (context.Context, context.CancelCauseFunc, chan<- *handlerutil.HandlerWorkspace, <-chan *handlerutil.HandlerWorkspace) {
		ctx, cancel := context.WithCancelCause(ctx)
		in := make(chan *handlerutil.HandlerWorkspace)

		out1 := handlerutil.Stage(ctx, cancel, doubleIt, in)
		out2 := handlerutil.Stage(ctx, cancel, ensurePositive, out1)

		return ctx, cancel, in, out2
	}

	t.Run("Completed", func(t *testing.T) {
		space := handlerutil.DefaultWorkspace()
		space.Set("X", 50)
		ctx, cancel, in, out := workflow(context.Background())
		defer cancel(nil)
		defer close(in)

		in <- &space

		select {
		case <-ctx.Done():
			require.NoError(t, context.Cause(ctx))
		case res, ok := <-out:
			if !ok {
				t.Fatal("Should have been able to read a result")
			} else {
				var x int
				res.Get("X", &x)
				require.Equal(t, 100, x)
			}
		}
	})

	t.Run("Interrupted", func(t *testing.T) {
		space := handlerutil.DefaultWorkspace()
		space.Set("X", -50)
		ctx, cancel, in, out := workflow(context.Background())
		defer cancel(nil)
		defer close(in)

		in <- &space

		<-ctx.Done()
		require.Error(t, context.Cause(ctx))

		select {
		case _, ok := <-out:
			require.False(t, ok, "Context should have been cancelled before reaching this point")
		default:
			break
		}
	})
}
