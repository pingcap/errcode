// Package grpc attaches GRPC codes to the standard error codes.
// It also provides helpers for integrating with GRPC.
//
// Note that not all GRPC codes are mapped right now: you are welcome to contribute more.
// Available mappings are documented here: https://cloud.google.com/apis/design/errors
//
// The init functiom performs the mapping and is reproduced here:
//
//	SetCode(errcode.InternalCode, codes.Internal)
//	SetCode(errcode.InvalidInputCode, codes.InvalidArgument)
//	SetCode(errcode.NotFoundCode, codes.NotFound)
//	SetCode(errcode.StateCode, codes.FailedPrecondition)
//	SetCode(errcode.ForbiddenCode, codes.PermissionDenied)
//	SetCode(errcode.NotAuthenticatedCode, codes.Unauthenticated)
//	SetCode(errcode.AlreadyExistsCode, codes.AlreadyExists)
//	SetCode(errcode.OutOfRangeCode, codes.OutOfRange)
//	SetCode(errcode.UnimplementedCode, codes.Unimplemented)
package grpc

import (
	"github.com/pingcap/errcode"
	"github.com/pingcap/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status creates a GRPC Status object from an ErrorCode.
// TODO: add more information in the details fields.
func Status(code errcode.ErrorCode) *status.Status {
	return status.New(GetCode(code.Code()), code.Error())
}

var grpcMetaData = make(errcode.MetaData)

// SetCode adds a GRPC code to the meta data of a code.
// The code can be retrieved with GRPCCode.
// Panic if the metadata is already set for the code.
// Returns itself.
func SetCode(code errcode.Code, grpcCode codes.Code) errcode.Code {
	if err := code.SetMetaData(grpcMetaData, grpcCode); err != nil {
		panic(errors.Annotate(err, "SetGRPC"))
	}
	return code
}

// GetCode retrieves the GRPC code for a code or its first ancestor with a GRPC code.
// If none are specified, it defaults to Unknown (Code 2).
// The return of this is a GRPC codes package Code, not an errcode.Code
func GetCode(code errcode.Code) codes.Code {
	grpcCode := code.MetaDataFromAncestors(grpcMetaData)
	if grpcCode == nil {
		return codes.Unknown
	}
	return grpcCode.(codes.Code)
}

func init() {
	SetCode(errcode.InternalCode, codes.Internal)
	SetCode(errcode.InvalidInputCode, codes.InvalidArgument)
	SetCode(errcode.NotFoundCode, codes.NotFound)
	SetCode(errcode.StateCode, codes.FailedPrecondition)
	SetCode(errcode.ForbiddenCode, codes.PermissionDenied)
	SetCode(errcode.NotAuthenticatedCode, codes.Unauthenticated)
	SetCode(errcode.AlreadyExistsCode, codes.AlreadyExists)
	SetCode(errcode.OutOfRangeCode, codes.OutOfRange)
	SetCode(errcode.UnimplementedCode, codes.Unimplemented)
}
