//go:build go1.24

package decoder

// embeddedFieldNames is whether the Field of a type error before Go 1.27 has the names of the embedded fields
// which a field is promoted through, as encoding/json of Go 1.24 and later reports them.
const embeddedFieldNames = true
