# errcode

This package extends go errors via interfaces to have error codes.

There are other packages that add capabilities including error codes to errors.
There is a fundamental difference in architectural decision that this package makes.
Other packages use a single error struct. A code or other annotation is a field on that struct.
errcode follows the model of pkg/errors to use wrapping and interfaces.
You are encouraged to make your own error structures and then fulfill the ErrorCode interface by adding a function.
Additional features (for example annotating the operation) are done via wrapping.

This package integrates and relies on pkg/errors (actually the pingcap/errors enhancement of it).

See the go documentation for extensive documentation.


## Error codes

Error codes are useful to transmit error types across network boundaries.
A program can reliably respond to an error if it can be sure to understand its type.
