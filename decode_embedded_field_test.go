package json_test

import (
	"testing"
)

type (
	// a struct embedded beside dominanceDepth2, which embeds the same struct at the same depth.
	dominanceSameDepth2 struct{ dominanceV }

	// an exported struct which embeds an unexported struct by pointer.
	PromotionPointer struct{ *dominanceV }

	// an integer embedded with the option ,string.
	PromotionInt int

	// structs whose methods of unmarshaling hide each other in a struct which embeds both.
	promotionJSONA struct{ A int }
	promotionJSONB struct{ B int }
	promotionTextA struct{ A int }
	promotionTextB struct{ B int }
)

func (p *promotionJSONA) UnmarshalJSON(b []byte) error { p.A = len(b); return nil }
func (p *promotionJSONB) UnmarshalJSON(b []byte) error { p.B = len(b); return nil }
func (p *promotionTextA) UnmarshalText(b []byte) error { p.A = len(b); return nil }
func (p *promotionTextB) UnmarshalText(b []byte) error { p.B = len(b); return nil }

// Of the fields of the same key in a struct and the structs embedded in it, only the least deep ones are decoded, and
// of those, the one with the key in its tag, or none if they are more than one, as encoding/json does. The structs are
// the ones TestEncodeEmbeddedFieldDominance writes.
func TestDecodeEmbeddedFieldDominance(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		value func() any
	}{
		{"depth 2 against depth 1", `{"V":2}`, func() any {
			return new(struct {
				dominanceDepth2
				dominanceOtherV
			})
		}},
		{"depth 3 against depth 1", `{"V":2}`, func() any {
			return new(struct {
				dominanceDepth3
				dominanceOtherV
			})
		}},
		{"depth 2 against depth 1 behind a pointer", `{"V":2}`, func() any {
			return new(struct {
				dominancePointer
				dominanceOtherV
			})
		}},
		{"a tagged field at depth 2 against an untagged one", `{"V":1}`, func() any {
			return new(struct {
				dominanceTaggedDepth2
				dominanceDepth2
			})
		}},
		{"a deeper tagged field against an untagged one", `{"V":3}`, func() any {
			return new(struct {
				dominanceDepth2
				dominanceTaggedDepth2
				dominanceOtherV
			})
		}},
		{"a tagged field against an untagged one", `{"V":1}`, func() any {
			return new(struct {
				dominanceTagged
				dominanceOtherV
			})
		}},
		{"two untagged fields", `{"V":1}`, func() any {
			return new(struct {
				dominanceV
				dominanceOtherV
			})
		}},
		{"two tagged fields", `{"V":1}`, func() any {
			return new(struct {
				dominanceTagged
				dominanceOtherTagged
			})
		}},
		{"three untagged fields at depth 2 in two embedded structs", `{"V":1}`, func() any {
			return new(struct {
				dominanceTwoV
				dominanceOneV
			})
		}},
		{"the same struct embedded twice at depth 2", `{"V":1}`, func() any {
			return new(struct {
				dominanceDepth2
				dominanceSameDepth2
			})
		}},
		{"a field of the struct itself", `{"V":2}`, func() any {
			return new(struct {
				dominanceDepth2
				V int
			})
		}},
		{"a tagged field of the struct itself against an untagged one", `{"B":2}`, func() any {
			return new(struct {
				A int `json:"B"`
				B int
			})
		}},
		{"an unexported struct embedded by pointer in a struct embedded by value", `{"V":1}`, func() any {
			return new(struct{ dominancePointer })
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compareRecursiveDecode(t, tt.in, tt.value, compareValuesAndErrors)
		})
	}
}

// The fields of the structs embedded without a key are promoted whatever the methods of the structs, and the other
// embedded fields are fields of their own, with their options, as encoding/json decodes them.
func TestDecodeEmbeddedFieldPromotion(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		value func() any
	}{
		{"an unexported struct embedded by pointer in an exported struct embedded by value", `{"V":1}`, func() any {
			return new(struct{ PromotionPointer })
		}},
		{"an integer embedded with the option ,string", `{"PromotionInt":"5"}`, func() any {
			return new(struct {
				PromotionInt `json:",string"`
			})
		}},
		{"an integer embedded with the option ,string, given a number", `{"PromotionInt":5}`, func() any {
			return new(struct {
				PromotionInt `json:",string"`
			})
		}},
		{"structs whose methods UnmarshalJSON hide each other", `{"A":1,"B":2,"W":3,"promotionJSONA":4}`, func() any {
			return new(struct {
				promotionJSONA
				promotionJSONB
				W int
			})
		}},
		{"structs whose methods UnmarshalText hide each other", `{"A":1,"B":2,"W":3,"promotionTextA":"4"}`, func() any {
			return new(struct {
				promotionTextA
				promotionTextB
				W int
			})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compareRecursiveDecode(t, tt.in, tt.value, compareValuesAndErrors)
		})
	}
}
