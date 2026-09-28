package runtime

import "reflect"

// stdMarshalerPackages are the packages of the standard library which have types with marshalers or unmarshalers,
// by their paths, which a package of another module can't have: the go command resolves the path of a package of
// the standard library to it, and reports the import as ambiguous when a module has the same path, and GOPATH
// mode finds the package of GOROOT first. A path is here only from the release of Go which has the package ( see
// std_package_go127.go ): a program built with an earlier release may have a package of its own at that path.
// The packages of encoding/json are not here: json.RawMessage and jsontext.Value hold the JSON of the user.
var stdMarshalerPackages = map[string]bool{
	"crypto/x509": true,
	"log/slog":    true,
	"math/big":    true,
	"net":         true,
	"net/netip":   true,
	"regexp":      true,
	"time":        true,
}

// IsStdMarshalerType is whether the type, or the type it points to, is a named type of a package of the standard
// library whose marshalers are trusted: what MarshalJSON returns is valid and compact, AppendText appends what
// MarshalText returns, and UnmarshalJSON and UnmarshalText keep nothing of the bytes they are given.
func IsStdMarshalerType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer && typ.Name() == "" {
		typ = typ.Elem()
	}
	return typ.Name() != "" && stdMarshalerPackages[typ.PkgPath()]
}
