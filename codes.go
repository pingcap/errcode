// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package errcode

import (
	"fmt"
	"net/http"
)

var (
	// InternalCode is equivalent to HTTP 500 Internal Server Error.
	InternalCode = NewCode("internal").SetHTTP(http.StatusInternalServerError)

	// NotFoundCode is equivalent to HTTP 404 Not Found.
	NotFoundCode = NewCode("missing").SetHTTP(http.StatusNotFound)

	// UnimplementedCode is mapped to HTTP 501.
	UnimplementedCode = InternalCode.Child("internal.unimplemented").SetHTTP(http.StatusNotImplemented)

	// StateCode is an error that is invalid due to the current system state.
	// This operatiom could become valid if the system state changes
	// This is mapped to HTTP 400.
	StateCode = NewCode("state").SetHTTP(http.StatusBadRequest)

	// AlreadyExistsCode indicates an attempt to create an entity failed because it already exists.
	// This is mapped to HTTP 409.
	AlreadyExistsCode = StateCode.Child("state.exists").SetHTTP(http.StatusConflict)

	// OutOfRangeCode indicates an operation was attempted past a valid range.
	// This is mapped to HTTP 400.
	OutOfRangeCode = StateCode.Child("state.range")

	// InvalidInputCode is equivalent to HTTP 400 Bad Request.
	InvalidInputCode = NewCode("input").SetHTTP(http.StatusBadRequest)

	// AuthCode represents an authentication or authorization issue.
	AuthCode = NewCode("auth")

	// NotAuthenticatedCode indicates the user is not authenticated.
	// This is mapped to HTTP 401.
	// Note that HTTP 401 is poorly named "Unauthorized".
	NotAuthenticatedCode = AuthCode.Child("auth.unauthenticated").SetHTTP(http.StatusUnauthorized)

	// ForbiddenCode indicates the user is not authorized.
	// This is mapped to HTTP 403.
	ForbiddenCode = AuthCode.Child("auth.forbidden").SetHTTP(http.StatusForbidden)
)

// invalidInput gives the code InvalidInputCode.
type invalidInputErr struct{ CodedError }

// NewInvalidInputErr creates an invalidInput from an err.
// If the error is already an ErrorCode it will use that code.
// Otherwise it will use InvalidInputCode which gives HTTP 400.
func NewInvalidInputErr(err error) ErrorCode {
	return invalidInputErr{NewCodedError(err, InvalidInputCode)}
}

var _ ErrorCode = (*invalidInputErr)(nil)     // assert implements interface
var _ HasClientData = (*invalidInputErr)(nil) // assert implements interface
var _ Causer = (*invalidInputErr)(nil)        // assert implements interface

// internalError gives the code InternalCode
type internalErr struct{ StackCode }

var internalStackCode = makeInternalStackCode(InternalCode)

// NewInternalErr creates an internalError from an err.
// If the given err is an ErrorCode that is a descendant of InternalCode,
// its code will be used.
// This ensures the intention of sending an HTTP 50x.
// This function also records a stack trace.
func NewInternalErr(err error) ErrorCode {
	return internalErr{internalStackCode(err)}
}

var _ ErrorCode = (*internalErr)(nil)     // assert implements interface
var _ HasClientData = (*internalErr)(nil) // assert implements interface
var _ Causer = (*internalErr)(nil)        // assert implements interface

// makeInternalStackCode builds a function for making an an internal error with a stack trace.
func makeInternalStackCode(defaultCode Code) func(error) StackCode {
	if !defaultCode.IsAncestor(InternalCode) {
		panic(fmt.Errorf("code is not an internal code: %v", defaultCode))
	}
	return func(err error) StackCode {
		if err == nil {
			panic("makeInternalStackCode error is nil")
		}
		code := defaultCode
		if errcode, ok := err.(ErrorCode); ok {
			errCode := errcode.Code()
			if errCode.IsAncestor(InternalCode) {
				code = errCode
			}
		}
		return NewStackCode(CodedError{GetCode: code, Err: err}, 3)
	}
}

type unimplementedErr struct{ StackCode }

var unimplementedStackCode = makeInternalStackCode(UnimplementedCode)

// NewUnimplementedErr creates an internalError from an err.
// If the given err is an ErrorCode that is a descendant of InternalCode,
// its code will be used.
// This ensures the intention of sending an HTTP 50x.
// This function also records a stack trace.
func NewUnimplementedErr(err error) ErrorCode {
	return unimplementedErr{unimplementedStackCode(err)}
}

// notFound gives the code NotFoundCode.
type notFoundErr struct{ CodedError }

// NewNotFoundErr creates a notFound from an err.
// If the error is already an ErrorCode it will use that code.
// Otherwise it will use NotFoundCode which gives HTTP 404.
func NewNotFoundErr(err error) ErrorCode {
	return notFoundErr{NewCodedError(err, NotFoundCode)}
}

var _ ErrorCode = (*notFoundErr)(nil)     // assert implements interface
var _ HasClientData = (*notFoundErr)(nil) // assert implements interface
var _ Causer = (*notFoundErr)(nil)        // assert implements interface

// notAuthenticatedErr gives the code NotAuthenticatedCode.
type notAuthenticatedErr struct{ CodedError }

// NewNotAuthenticatedErr creates a notAuthenticatedErr from an err.
// If the error is already an ErrorCode it will use that code.
// Otherwise it will use NotAuthenticatedCode which gives HTTP 401.
func NewNotAuthenticatedErr(err error) ErrorCode {
	return notAuthenticatedErr{NewCodedError(err, NotAuthenticatedCode)}
}

var _ ErrorCode = (*notAuthenticatedErr)(nil)     // assert implements interface
var _ HasClientData = (*notAuthenticatedErr)(nil) // assert implements interface
var _ Causer = (*notAuthenticatedErr)(nil)        // assert implements interface

// forbiddenErr gives the code ForbiddenCode.
type forbiddenErr struct{ CodedError }

// NewForbiddenErr creates a forbiddenErr from an err.
// If the error is already an ErrorCode it will use that code.
// Otherwise it will use ForbiddenCode which gives HTTP 401.
func NewForbiddenErr(err error) ErrorCode {
	return forbiddenErr{NewCodedError(err, ForbiddenCode)}
}

var _ ErrorCode = (*forbiddenErr)(nil)     // assert implements interface
var _ HasClientData = (*forbiddenErr)(nil) // assert implements interface
var _ Causer = (*forbiddenErr)(nil)        // assert implements interface

// CodedError is a convenience to attach a code to an error and already satisfy the ErrorCode interface.
// If the error is a struct, that struct will get preseneted as data to the client.
//
// To override the http code or the data representation or just for clearer documentation,
// you are encouraged to wrap CodeError with your own struct that inherits it.
// Look at the implementation of invalidInput, internalError, and notFound.
type CodedError struct {
	GetCode Code
	Err     error
}

// NewCodedError is for constructing broad error kinds (e.g. those representing HTTP codes)
// Which could have many different underlying go errors.
// Eventually you may want to give your go errors more specific codes.
// The second argument is the broad code.
//
// If the error given is already an ErrorCode,
// that will be used as the code instead of the second argument.
func NewCodedError(err error, code Code) CodedError {
	if err == nil {
		panic("NewCodedError error is nil")
	}
	if errcode, ok := err.(ErrorCode); ok {
		code = errcode.Code()
	}
	return CodedError{GetCode: code, Err: err}
}

var _ ErrorCode = (*CodedError)(nil)     // assert implements interface
var _ HasClientData = (*CodedError)(nil) // assert implements interface
var _ Causer = (*CodedError)(nil)        // assert implements interface

func (e CodedError) Error() string {
	return e.Err.Error()
}

// Cause satisfies the Causer interface.
func (e CodedError) Cause() error {
	return e.Err
}

// Code returns the GetCode field
func (e CodedError) Code() Code {
	return e.GetCode
}

// GetClientData returns the underlying Err field.
func (e CodedError) GetClientData() interface{} {
	if errCode, ok := e.Err.(ErrorCode); ok {
		return ClientData(errCode)
	}
	return e.Err
}
