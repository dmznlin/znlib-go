package restruct

import (
	"github.com/dmznlin/znlib-go/znlib/restruct/expr"
)

var (
	expressionsEnabled = false
	stdLibResolver     = expr.NewMapResolver(exprStdLib)
)

// EnableExprBeta enables you to use restruct expr while it is still in beta.
// Use at your own risk. Functionality may change in unforeseen, incompatible
// ways at any time.
func EnableExprBeta() {
	expressionsEnabled = true
}
