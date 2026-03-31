package handlerutil_test

/*
 * File: pkg/handlerutil/errors_test.go
 *
 * Purpose: unit tests for the error formatting logic
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tournabyte/webapi/pkg/handlerutil"
)

func TestErrorDetailCanConvertToMap(t *testing.T) {
	f := handlerutil.ErrNotAuthorized(handlerutil.NewDetail("credentials", "invalid email or password"))
	d := f.DetailMapping()

	assert.Contains(t, d, "credentials")
}

func TestErrorDetailsCanConverToMap(t *testing.T) {
	f := handlerutil.ErrUnprocessibleEntity(
		handlerutil.NewDetail("port", "must be between 1 and 65,535"),
		handlerutil.NewDetail("displayName", "must not contain spaces"),
	)
	d := f.DetailMapping()

	assert.Contains(t, d, "port")
	assert.Contains(t, d, "displayName")
}

func TestNoErrorDetailsConvertsToNil(t *testing.T) {
	f := handlerutil.ErrInternalServerError()
	d := f.DetailMapping()

	assert.Nil(t, d)
}

func TestFailureFormatting(t *testing.T) {
	var (
		SomePlannedError    = errors.New("some planned error")
		AnotherPlannedError = errors.New("another planned error")
		UnexpectedError     = errors.New("unexpected error")
	)

	var (
		Rule1 = func(e error) (error, bool) {
			if errors.Is(e, SomePlannedError) {
				return handlerutil.ErrBadRequest(), true
			}
			return nil, false
		}
		Rule2 = func(e error) (error, bool) {
			if errors.Is(e, AnotherPlannedError) {
				return handlerutil.ErrNoAccess(), true
			}
			return nil, false
		}
	)

	ffmt := handlerutil.FailureFormatter(Rule1, Rule2)

	t.Run("MatchRule1", func(t *testing.T) {
		e := ffmt.Format(SomePlannedError)

		assert.NotEqual(t, e.Error(), SomePlannedError.Error())
	})

	t.Run("MatchRule2", func(t *testing.T) {
		e := ffmt.Format(AnotherPlannedError)

		assert.NotEqual(t, e.Error(), AnotherPlannedError.Error())
	})

	t.Run("NoMatch", func(t *testing.T) {
		e := ffmt.Format(UnexpectedError)

		assert.NotEqual(t, e.Error(), UnexpectedError.Error())
	})
}
