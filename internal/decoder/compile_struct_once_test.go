package decoder

import (
	"reflect"
	"testing"
)

// Structs built of embeds only, each embedding the two before it: a struct embedded k levels down is met through
// about 1.6^k paths of embeds.
type (
	embedPaths0 struct{ Y0 int }
	embedPaths1 struct {
		embedPaths0
		Y1 int
	}
	embedPaths2 struct {
		embedPaths1
		embedPaths0
		Y2 int
	}
	embedPaths3 struct {
		embedPaths2
		embedPaths1
		Y3 int
	}
	embedPaths4 struct {
		embedPaths3
		embedPaths2
		Y4 int
	}
	embedPaths5 struct {
		embedPaths4
		embedPaths3
		Y5 int
	}
	embedPaths6 struct {
		embedPaths5
		embedPaths4
		Y6 int
	}
	embedPaths7 struct {
		embedPaths6
		embedPaths5
		Y7 int
	}
	embedPaths8 struct {
		embedPaths7
		embedPaths6
		Y8 int
	}
	embedPaths9 struct {
		embedPaths8
		embedPaths7
		Y9 int
	}
	embedPaths10 struct {
		embedPaths9
		embedPaths8
		Y10 int
	}
	embedPaths11 struct {
		embedPaths10
		embedPaths9
		Y11 int
	}
	embedPaths12 struct {
		embedPaths11
		embedPaths10
		Y12 int
	}
	embedPaths13 struct {
		embedPaths12
		embedPaths11
		Y13 int
	}
	embedPaths14 struct {
		embedPaths13
		embedPaths12
		Y14 int
	}
	embedPaths15 struct {
		embedPaths14
		embedPaths13
		Y15 int
	}
	embedPaths16 struct {
		embedPaths15
		embedPaths14
		Y16 int
	}
)

// Structs which each embed the next one and hold a slice of themselves, which every struct they are embedded in
// holds too.
type (
	embedChain0 struct {
		embedChain1
		S0 []embedChain0
	}
	embedChain1 struct {
		embedChain2
		S1 []embedChain1
	}
	embedChain2 struct {
		embedChain3
		S2 []embedChain2
	}
	embedChain3 struct {
		embedChain4
		S3 []embedChain3
	}
	embedChain4 struct {
		embedChain5
		S4 []embedChain4
	}
	embedChain5 struct {
		embedChain6
		S5 []embedChain5
	}
	embedChain6 struct {
		embedChain7
		S6 []embedChain6
	}
	embedChain7 struct {
		embedChain8
		S7 []embedChain7
	}
	embedChain8 struct {
		embedChain9
		S8 []embedChain8
	}
	embedChain9 struct {
		embedChain10
		S9 []embedChain9
	}
	embedChain10 struct {
		embedChain11
		S10 []embedChain10
	}
	embedChain11 struct {
		embedChain12
		S11 []embedChain11
	}
	embedChain12 struct {
		embedChain13
		S12 []embedChain12
	}
	embedChain13 struct {
		embedChain14
		S13 []embedChain13
	}
	embedChain14 struct {
		embedChain15
		S14 []embedChain14
	}
	embedChain15 struct {
		embedChain16
		S15 []embedChain15
	}
	embedChain16 struct{ X int }
)

// Structs which each have two fields of the next one.
type (
	fieldPair0  struct{ A, B fieldPair1 }
	fieldPair1  struct{ A, B fieldPair2 }
	fieldPair2  struct{ A, B fieldPair3 }
	fieldPair3  struct{ A, B fieldPair4 }
	fieldPair4  struct{ A, B fieldPair5 }
	fieldPair5  struct{ A, B fieldPair6 }
	fieldPair6  struct{ A, B fieldPair7 }
	fieldPair7  struct{ A, B fieldPair8 }
	fieldPair8  struct{ A, B fieldPair9 }
	fieldPair9  struct{ A, B fieldPair10 }
	fieldPair10 struct{ A, B fieldPair11 }
	fieldPair11 struct{ A, B fieldPair12 }
	fieldPair12 struct{ A, B fieldPair13 }
	fieldPair13 struct{ A, B fieldPair14 }
	fieldPair14 struct{ A, B fieldPair15 }
	fieldPair15 struct{ A, B fieldPair16 }
	fieldPair16 struct{ X int }
)

// A struct type is compiled once for a compile, however many fields and paths of embeds lead to it: the allocations
// of the compile of a struct with twice the levels are a few times as many, as the fields of the structs are, where a
// compile for each field or path makes them tens or hundreds of times as many. They are counted, so that the test
// tells them apart on any machine.
func TestCompileStructTypeOnce(t *testing.T) {
	tests := []struct {
		name        string
		short, long reflect.Type
	}{
		{"embeds of the two structs before", reflect.TypeOf(&embedPaths8{}), reflect.TypeOf(&embedPaths16{})},
		{"embeds of the next struct, with slices of themselves", reflect.TypeOf(&embedChain8{}), reflect.TypeOf(&embedChain0{})},
		{"two fields of the next struct", reflect.TypeOf(&fieldPair8{}), reflect.TypeOf(&fieldPair0{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocs := func(typ reflect.Type) float64 {
				return testing.AllocsPerRun(1, func() {
					if _, err := compileHead(typ, map[uintptr]Decoder{}); err != nil {
						t.Fatal(err)
					}
				})
			}
			short, long := allocs(tt.short), allocs(tt.long)
			t.Logf("%v allocations for 8 levels, %v for 16", short, long)
			if long > 16*short {
				t.Errorf("the compile of 16 levels makes %.1f times the allocations of 8 levels", long/short)
			}
		})
	}
}
