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

	"github.com/pingcap/errors"
)

// MetaData is used in a pattern for attaching meta data to codes and inheriting it from a parent.
// See MetaDataFromAncestors.
// This is used to attach an HTTP code to a Code as meta data.
type MetaData map[CodeStr]interface{}

// MetaDataFromAncestors looks for meta data starting at the current code.
// If not found, it traverses up the hierarchy
// by looking for the first ancestor with the given metadata key.
// This is used in the HTTPCode implementation to inherit the HTTP Code from ancestors.
func (code Code) MetaDataFromAncestors(metaData MetaData) interface{} {
	if existing, ok := metaData[code.CodeStr()]; ok {
		return existing
	}
	if code.Parent == nil {
		return nil
	}
	return (*code.Parent).MetaDataFromAncestors(metaData)
}

type existingCodeError struct {
	existingMetaData interface{}
	code             Code
}

func (e existingCodeError) Error() string {
	return fmt.Sprintf("for code %v metadata exists: %v", e.code, e.existingMetaData)
}

// SetMetaData is used to implement meta data setters such as SetHTTPCode.
// Return an error if the metadata is already set.
func (code Code) SetMetaData(metaData MetaData, item interface{}) error {
	if existingCode, ok := metaData[code.CodeStr()]; ok {
		return existingCodeError{
			existingMetaData: existingCode,
			code:             code,
		}
	}
	metaData[code.CodeStr()] = item
	return nil
}

var httpMetaData = make(MetaData)

// SetHTTP adds an HTTP code to the meta data.
// The code can be retrieved with HTTPCode.
// Panic if the metadata is already set for the code.
// Returns itself.
func (code Code) SetHTTP(httpCode int) Code {
	if err := code.SetMetaData(httpMetaData, httpCode); err != nil {
		panic(errors.Annotate(err, "SetHTTP"))
	}
	return code
}

// HTTPCode retrieves the HTTP code for a code or its first ancestor with an HTTP code.
// If none are specified, it defaults to 400 BadRequest
func (code Code) HTTPCode() int {
	httpCode := code.MetaDataFromAncestors(httpMetaData)
	if httpCode == nil {
		return http.StatusBadRequest
	}
	return httpCode.(int)
}
