## About Error codes

Error codes are particularly useful to reliably communicate the error type across program (network) boundaries.
A program can reliably respond to an error if it can be sure to understand its type.
Error monitoring systems can reliably understand what errors are occurring.


# errcode overview

This package extends go errors via interfaces to have error codes.
The two main goals are

  * The clients can reliably understand errors by checking against error codes.
  * Structure information is sent to the client

See the [go docs](https://godoc.org/github.com/pingcap/errcode) for extensive API documentation.

There are other packages that add error code capabilities to errors.
However, they almost universally use a single underlying error struct.
A code or other annotation is a field on that struct.
In most cases a structured error response is only possible to create dynamically via tags and labels.

errcode instead follows the model of pkg/errors to use wrapping and interfaces.
You are encouraged to make your own error structures and then fulfill the ErrorCode interface by adding a function.
Additional features (for example annotating the operation) are done via wrapping.

This design makes errcode highly extensible, inter-operable, and structure preserving (of the original error).
It is easy to gradually introduce it into a project that is using pkg/errors (or the Cause interface).


## Features

* structured error representation
* Follows the pkg/errors model where error enhancements are annotations via the Causer interface.
* Internal errors show a stack trace but others don't.
* Operation annotation. This concept is [explained here](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html).
* Works for multiple errors when the Errors() interface is used. See the `Combine` function for constructing multiple error codes.
* Extensible metadata. See how SetHTTPCode is implemented.
* Integration with existing error codes
  * HTTP
  * GRPC (provided by separate grpc package)


## Example


``` go
// First define a normal error type
type PathBlocked struct {
	start     uint64 `json:"start"`
	end       uint64 `json:"end"`
	obstacle  uint64 `json:"end"`
}

func (e PathBlocked) Error() string {
	return fmt.Sprintf("The path %d -> %d has obstacle %d", e.start, e.end, e.obstacle)
}

// Define a code. These often get re-used between different errors.
// Note that codes use a hierarchy to share metadata.
// This code is a child of the "state" code.
var PathBlockedCode = errcode.StateCode.Child("state.blocked")

// Now attach the code to your custom error type.
func (e PathBlocked) Code() Code {
	return PathBlockedCode
}

var _ ErrorCode = (*PathBlocked)(nil)  // assert implements the ErrorCode interface
```

Now lets see how you can send the error code to a client in a way that works with your exising code.

``` go
// Given just a type of error, give an error code to a client if it is present
if errCode := errcode.CodeChain(err); errCode != nil {
	// If it helps, you can set the code a header
	w.Header().Set("X-Error-Code", errCode.Code().CodeStr().String())

	// But the code will also be in the JSON body
	// Codes have HTTP codes attached as meta data.
	// You can also attach other kinds of meta data to codes.
	// Our error code inherits StatusBadRequest from its parent code "state"
	rd.JSON(w, errCode.Code().HTTPCode(), errcode.NewJSONFormat(errCode))
}
```

Let see a usage site. This example will include an annotation concept of "operation".

``` go
func moveX(start, end, obstacle) error {}
	op := errcode.Op("path.move.x")

	if start < obstable && obstacle < end  {
		return op.AddTo(PathBlocked{start, end, obstacle})
	}
}
```


## Development

``` shell
./scripts build
./scripts test
./scripts check
```
