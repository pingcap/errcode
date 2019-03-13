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

package grpc_test

import (
	"fmt"
	"testing"

	"github.com/pingcap/errcode"
	"github.com/pingcap/errcode/grpc"
	"google.golang.org/grpc/codes"
)

// Test setting the HTTP code
type GRPCError struct{}

func (e GRPCError) Error() string { return "error" }

const grpcCodeStr = "input.grpc"

var codeAborted = grpc.SetCode(errcode.InvalidInputCode.Child(grpcCodeStr), codes.Aborted)

func (e GRPCError) Code() errcode.Code {
	return codeAborted
}

func TestGrpcErrorCode(t *testing.T) {
	err := GRPCError{}
	AssertGRPCCode(t, err, codes.Aborted)
}

func TestWrapAsGrpc(t *testing.T) {
	err := grpc.WrapAsGRPC(errcode.NewInternalErr(fmt.Errorf("wrap me up")))
	AssertGRPCCode(t, err, codes.Internal)
}

func AssertGRPCCode(t *testing.T, code errcode.ErrorCode, grpcCode codes.Code) {
	t.Helper()
	expected := grpc.GetCode(code.Code())
	if expected != grpcCode {
		t.Errorf("excpected HTTP Code %v but got %v", grpcCode, expected)
	}
}
