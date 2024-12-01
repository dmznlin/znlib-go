//go:build !go1.9
// +build !go1.9

package restruct

import (
	"github.com/dmznlin/znlib-go/znlib/restruct/expr"
)

var exprStdLib = map[string]expr.Value{}
