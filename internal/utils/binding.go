package utils

/*
 * File: internal/utils/binding.go
 *
 * Purpose: helpers to bind request context attributes to addressable values (headers, parameters, forms, body)
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"errors"

	"github.com/carlmjohnson/truthy"
	"github.com/gin-gonic/gin"
)

// Type `binder` represents a function that binds data to the given addressable value and reports the outcome with an error value
type binder func(any) error

// Type `Bindings` represents the possible bindings of an HTTP context to value
//
// Fields:
//   - Headers: binder to bind header values
//   - Body: binder to bind request body
//   - URI: binder to bind URI values
//   - Query: binder to bind query parameters
type Bindings struct {
	Headers binder
	Body    binder
	URI     binder
	Query   binder
}

// errors representing issues that may occur during value binding
var (
	errNilBinder = errors.New("attempting to bind without a binder function")
)

// Constants are bitwise fields for which binders should exist (nil binders should immediately error out)
const (
	ShouldHaveHeaders uint8 = 0b0001 << iota
	ShouldHaveJSONBody
	ShouldHaveURIValues
	ShouldHaveQueryParameters
)

// Function `BindingsFromRequestContext` creates a `Bindings` instance from the given request context
//
// Parameters:
//   - ctx: the request context to obtain binder functions from
//   - flags: the bitwise flags to determine which binder functions to keep
//
// Returns:
//   - `Bindings`: instance with value binding capabilities as specified by the flags
func BindingsFromRequestContext(ctx *gin.Context, flags uint8) Bindings {
	return Bindings{
		Headers: truthy.Cond(flags&ShouldHaveHeaders > 0, ctx.ShouldBindHeader, nil),
		Body:    truthy.Cond(flags&ShouldHaveJSONBody > 0, ctx.ShouldBindJSON, nil),
		URI:     truthy.Cond(flags&ShouldHaveURIValues > 0, ctx.ShouldBindUri, nil),
		Query:   truthy.Cond(flags&ShouldHaveQueryParameters > 0, ctx.ShouldBindQuery, nil),
	}
}

// Function `(*Bindings).BindHeaders` attempts to bind the headers to the given addressable value
//
// Parameters:
//   - dst: the addressable to bind the headers to
//
// Returns:
//   - `error`: issue that occurred with either binding capabilities or the binding function
func (b *Bindings) BindHeaders(dst any) error {
	if b.Headers == nil {
		return errNilBinder
	}
	return b.Headers(dst)
}

// Function `(*Bindings).BindBodyAsJSON` attempts to bind the body (interpreted as JSON) to the given addressable value
//
// Parameters:
//   - dst: the addressable to bind the body to
//
// Returns:
//   - `error`: issue that occurred with either binding capabilities or the binding function
func (b *Bindings) BindBodyAsJSON(dst any) error {
	if b.Body == nil {
		return errNilBinder
	}
	return b.Body(dst)
}

// Function `(*Bindings).BindURI` attempts to bind the URI values to the given addressable value
//
// Parameters:
//   - dst: the addressable to bind the URI values to
//
// Returns:
//   - `error`: issue that occurred with either binding capabilities or the binding function
func (b *Bindings) BindURI(dst any) error {
	if b.URI == nil {
		return errNilBinder
	}
	return b.URI(dst)
}

// Function `(*Bindings).BindQueryParameters` attempts to bind the query parameters to the given addressable value
//
// Parameters:
//   - dst: the addressable to bind the query parameters to
//
// Returns:
//   - `error`: issue that occurred with either binding capabilities or the binding function
func (b *Bindings) BindQueryParameters(dst any) error {
	if b.Query == nil {
		return errNilBinder
	}
	return b.Query(dst)
}
