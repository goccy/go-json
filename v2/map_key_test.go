package json

import (
	"fmt"
	"testing"
)

type textStringerKey int

func (k textStringerKey) String() string { return fmt.Sprint(int(k)) }

func (k textStringerKey) MarshalText() ([]byte, error) { return []byte("k" + k.String()), nil }

type plainStringerKey int

func (k plainStringerKey) String() string { return "p" }

// The keys of a map of an interface type with methods are written by their dynamic values, as encoding/json/v2
// writes them.
func TestMarshalMapOfMethodInterfaceKeys(t *testing.T) {
	for _, tt := range []struct {
		in      any
		want    string
		wantErr bool
	}{
		{in: map[fmt.Stringer]int{textStringerKey(1): 1}, want: `{"k1":1}`},
		{in: map[fmt.Stringer]int{plainStringerKey(1): 1}, want: `{"1":1}`},
		{in: map[error]int{}, want: `{}`},
		{in: map[fmt.Stringer]int{nil: 2}, wantErr: true},
	} {
		got, err := Marshal(tt.in, Deterministic(true))
		if (err != nil) != tt.wantErr || !tt.wantErr && string(got) != tt.want {
			t.Errorf("%T: got %s %v, want %s (error %v)", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}
