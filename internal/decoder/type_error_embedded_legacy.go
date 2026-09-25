//go:build !go1.24

package decoder

// embeddedFieldNames is whether the Field of a type error before Go 1.27 has the names of the embedded fields
// which a field is promoted through: encoding/json before Go 1.24 reports the name of the field only.
const embeddedFieldNames = false
