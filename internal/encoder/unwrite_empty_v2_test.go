package encoder_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/encoder/run"
	ierrors "github.com/goccy/go-json/internal/errors"
)

// encodeV2 encodes v by the opcodes of the v2 semantics, with the options of flags in addition to the default ones of
// encoding/json/v2, and returns the output without the comma after the value, or the error.
func encodeV2(v any, flags encoder.OptionFlag) (string, error) {
	ctx := encoder.TakeRuntimeContext()
	defer encoder.ReleaseRuntimeContext(ctx)
	ctx.Option.Flag = encoder.V2Option | encoder.TextEscapeOption | encoder.RejectInvalidUTF8Option | flags
	b, err := run.Encode(ctx, v)
	if err != nil {
		return string(b), err
	}
	return strings.TrimSuffix(string(b), ","), nil
}

// skipText is a text whose MarshalText writes "", an empty value of omitempty.
type skipText string

func (skipText) MarshalText() ([]byte, error) { return nil, nil }

// skipOmit has fields of omitempty of each kind of pointer, each after a member whose value is "": an omitted field
// skips the opcode which unwrites an empty member of it, and doesn't unwrite the member before it, which ends with
// an empty value as an empty member of the field would.
type skipOmit struct {
	E0  string
	Int **int `json:",omitempty"`
	E1  string
	Str **string `json:",omitempty"`
	E2  string
	Byt *[]byte `json:",omitempty"`
	E3  string
	Txt *skipText `json:",omitempty"`
	E4  string
	Any any `json:",omitempty"`
	E5  string
}

func TestUnwriteEmptyMember(t *testing.T) {
	var nilInt *int
	empty := ""
	emptyPointer := &empty
	members := `"E0":"","E1":"","E2":"","E3":"","E4":"","E5":""`
	for _, tt := range []struct {
		v    skipOmit
		want string
	}{
		// omitted by the check of nil before the member is written.
		{skipOmit{}, `{` + members + `}`},
		// written as null or "", then unwritten.
		{skipOmit{Int: &nilInt, Str: &emptyPointer, Byt: &[]byte{}, Txt: new(skipText), Any: ""}, `{` + members + `}`},
		// "\"" ends with the bytes of "", but isn't empty.
		{skipOmit{Any: `"`}, `{"E0":"","E1":"","E2":"","E3":"","E4":"","Any":"\"","E5":""}`},
	} {
		got, err := encodeV2(tt.v, encoder.UnorderedMapOption)
		if err != nil || got != tt.want {
			t.Errorf("%+v: got %s, %v; want %s", tt.v, got, err, tt.want)
		}
	}
}

// dupKey is a key of a map whose name is the one of other keys: the keys of the same parity have the same name.
type dupKey int

func (k dupKey) MarshalText() ([]byte, error) { return []byte([]string{"even", "odd"}[k&1]), nil }

// The entries of a sorted map are checked for the same names as they are written in the order of their names: the
// first name which the entry before it has is reported, with the output before it.
func TestSortedMapDuplicateNames(t *testing.T) {
	// the entries of the same name are in the order of the map, which is random: the output before the duplicate
	// name has the one which comes first.
	for _, tt := range []struct {
		v       map[dupKey]int
		flags   encoder.OptionFlag
		want    []string
		wantOut []string
	}{
		{v: map[dupKey]int{1: 1, 2: 2}, want: []string{`{"even":2,"odd":1}`}},
		{v: map[dupKey]int{0: 0, 1: 1, 2: 2, 3: 3}, wantOut: []string{`{"even":0,`, `{"even":2,`}},
		{v: map[dupKey]int{1: 1, 2: 2, 3: 3}, wantOut: []string{`{"even":2,"odd":1,`, `{"even":2,"odd":3,`}},
		{v: map[dupKey]int{1: 1, 3: 3}, flags: encoder.AllowDuplicateNamesOption, want: []string{`{"odd":1,"odd":3}`, `{"odd":3,"odd":1}`}},
	} {
		got, err := encodeV2(tt.v, tt.flags)
		var terr *ierrors.TextError
		switch {
		case tt.wantOut == nil:
			if err != nil || !slices.Contains(tt.want, got) {
				t.Errorf("%v: got %s, %v; want one of %q", tt.v, got, err, tt.want)
			}
		case !errors.As(err, &terr) || terr.Err != ierrors.ErrDuplicateName || !slices.Contains(tt.wantOut, string(terr.Out)):
			t.Errorf("%v: got %s, %v; want the error of a duplicate name after one of %q", tt.v, got, err, tt.wantOut)
		}
	}
}
