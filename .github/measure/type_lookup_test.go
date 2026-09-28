package json_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/goccy/go-json"
)

// The benchmarks of the lookups of the opcodes and of the decoders of the types, by the ways a program uses the
// types: one type again and again, many types by turns as a server with many kinds of requests does, the types of
// the values of interface{}, with the goroutines in parallel, and in a program of many types, whose tables are
// large. Every value is small, so that the lookups weigh in the time.

type (
	lookupType0 struct {
		A int
		B string
	}
	lookupType1 struct {
		A int
		B string
	}
	lookupType2 struct {
		A int
		B string
	}
	lookupType3 struct {
		A int
		B string
	}
	lookupType4 struct {
		A int
		B string
	}
	lookupType5 struct {
		A int
		B string
	}
	lookupType6 struct {
		A int
		B string
	}
	lookupType7 struct {
		A int
		B string
	}
	lookupType8 struct {
		A int
		B string
	}
	lookupType9 struct {
		A int
		B string
	}
	lookupType10 struct {
		A int
		B string
	}
	lookupType11 struct {
		A int
		B string
	}
	lookupType12 struct {
		A int
		B string
	}
	lookupType13 struct {
		A int
		B string
	}
	lookupType14 struct {
		A int
		B string
	}
	lookupType15 struct {
		A int
		B string
	}
	lookupType16 struct {
		A int
		B string
	}
	lookupType17 struct {
		A int
		B string
	}
	lookupType18 struct {
		A int
		B string
	}
	lookupType19 struct {
		A int
		B string
	}
	lookupType20 struct {
		A int
		B string
	}
	lookupType21 struct {
		A int
		B string
	}
	lookupType22 struct {
		A int
		B string
	}
	lookupType23 struct {
		A int
		B string
	}
	lookupType24 struct {
		A int
		B string
	}
	lookupType25 struct {
		A int
		B string
	}
	lookupType26 struct {
		A int
		B string
	}
	lookupType27 struct {
		A int
		B string
	}
	lookupType28 struct {
		A int
		B string
	}
	lookupType29 struct {
		A int
		B string
	}
	lookupType30 struct {
		A int
		B string
	}
	lookupType31 struct {
		A int
		B string
	}
	lookupType32 struct {
		A int
		B string
	}
	lookupType33 struct {
		A int
		B string
	}
	lookupType34 struct {
		A int
		B string
	}
	lookupType35 struct {
		A int
		B string
	}
	lookupType36 struct {
		A int
		B string
	}
	lookupType37 struct {
		A int
		B string
	}
	lookupType38 struct {
		A int
		B string
	}
	lookupType39 struct {
		A int
		B string
	}
	lookupType40 struct {
		A int
		B string
	}
	lookupType41 struct {
		A int
		B string
	}
	lookupType42 struct {
		A int
		B string
	}
	lookupType43 struct {
		A int
		B string
	}
	lookupType44 struct {
		A int
		B string
	}
	lookupType45 struct {
		A int
		B string
	}
	lookupType46 struct {
		A int
		B string
	}
	lookupType47 struct {
		A int
		B string
	}
	lookupType48 struct {
		A int
		B string
	}
	lookupType49 struct {
		A int
		B string
	}
	lookupType50 struct {
		A int
		B string
	}
	lookupType51 struct {
		A int
		B string
	}
	lookupType52 struct {
		A int
		B string
	}
	lookupType53 struct {
		A int
		B string
	}
	lookupType54 struct {
		A int
		B string
	}
	lookupType55 struct {
		A int
		B string
	}
	lookupType56 struct {
		A int
		B string
	}
	lookupType57 struct {
		A int
		B string
	}
	lookupType58 struct {
		A int
		B string
	}
	lookupType59 struct {
		A int
		B string
	}
	lookupType60 struct {
		A int
		B string
	}
	lookupType61 struct {
		A int
		B string
	}
	lookupType62 struct {
		A int
		B string
	}
	lookupType63 struct {
		A int
		B string
	}
	lookupType64 struct {
		A int
		B string
	}
	lookupType65 struct {
		A int
		B string
	}
	lookupType66 struct {
		A int
		B string
	}
	lookupType67 struct {
		A int
		B string
	}
	lookupType68 struct {
		A int
		B string
	}
	lookupType69 struct {
		A int
		B string
	}
	lookupType70 struct {
		A int
		B string
	}
	lookupType71 struct {
		A int
		B string
	}
	lookupType72 struct {
		A int
		B string
	}
	lookupType73 struct {
		A int
		B string
	}
	lookupType74 struct {
		A int
		B string
	}
	lookupType75 struct {
		A int
		B string
	}
	lookupType76 struct {
		A int
		B string
	}
	lookupType77 struct {
		A int
		B string
	}
	lookupType78 struct {
		A int
		B string
	}
	lookupType79 struct {
		A int
		B string
	}
	lookupType80 struct {
		A int
		B string
	}
	lookupType81 struct {
		A int
		B string
	}
	lookupType82 struct {
		A int
		B string
	}
	lookupType83 struct {
		A int
		B string
	}
	lookupType84 struct {
		A int
		B string
	}
	lookupType85 struct {
		A int
		B string
	}
	lookupType86 struct {
		A int
		B string
	}
	lookupType87 struct {
		A int
		B string
	}
	lookupType88 struct {
		A int
		B string
	}
	lookupType89 struct {
		A int
		B string
	}
	lookupType90 struct {
		A int
		B string
	}
	lookupType91 struct {
		A int
		B string
	}
	lookupType92 struct {
		A int
		B string
	}
	lookupType93 struct {
		A int
		B string
	}
	lookupType94 struct {
		A int
		B string
	}
	lookupType95 struct {
		A int
		B string
	}
	lookupType96 struct {
		A int
		B string
	}
	lookupType97 struct {
		A int
		B string
	}
	lookupType98 struct {
		A int
		B string
	}
	lookupType99 struct {
		A int
		B string
	}
	lookupType100 struct {
		A int
		B string
	}
	lookupType101 struct {
		A int
		B string
	}
	lookupType102 struct {
		A int
		B string
	}
	lookupType103 struct {
		A int
		B string
	}
	lookupType104 struct {
		A int
		B string
	}
	lookupType105 struct {
		A int
		B string
	}
	lookupType106 struct {
		A int
		B string
	}
	lookupType107 struct {
		A int
		B string
	}
	lookupType108 struct {
		A int
		B string
	}
	lookupType109 struct {
		A int
		B string
	}
	lookupType110 struct {
		A int
		B string
	}
	lookupType111 struct {
		A int
		B string
	}
	lookupType112 struct {
		A int
		B string
	}
	lookupType113 struct {
		A int
		B string
	}
	lookupType114 struct {
		A int
		B string
	}
	lookupType115 struct {
		A int
		B string
	}
	lookupType116 struct {
		A int
		B string
	}
	lookupType117 struct {
		A int
		B string
	}
	lookupType118 struct {
		A int
		B string
	}
	lookupType119 struct {
		A int
		B string
	}
	lookupType120 struct {
		A int
		B string
	}
	lookupType121 struct {
		A int
		B string
	}
	lookupType122 struct {
		A int
		B string
	}
	lookupType123 struct {
		A int
		B string
	}
	lookupType124 struct {
		A int
		B string
	}
	lookupType125 struct {
		A int
		B string
	}
	lookupType126 struct {
		A int
		B string
	}
	lookupType127 struct {
		A int
		B string
	}
)

// lookupValues are values of lookupTypeCount distinct types.
var lookupValues = []any{
	lookupType0{0, "v"}, lookupType1{1, "v"}, lookupType2{2, "v"}, lookupType3{3, "v"},
	lookupType4{4, "v"}, lookupType5{5, "v"}, lookupType6{6, "v"}, lookupType7{7, "v"},
	lookupType8{8, "v"}, lookupType9{9, "v"}, lookupType10{10, "v"}, lookupType11{11, "v"},
	lookupType12{12, "v"}, lookupType13{13, "v"}, lookupType14{14, "v"}, lookupType15{15, "v"},
	lookupType16{16, "v"}, lookupType17{17, "v"}, lookupType18{18, "v"}, lookupType19{19, "v"},
	lookupType20{20, "v"}, lookupType21{21, "v"}, lookupType22{22, "v"}, lookupType23{23, "v"},
	lookupType24{24, "v"}, lookupType25{25, "v"}, lookupType26{26, "v"}, lookupType27{27, "v"},
	lookupType28{28, "v"}, lookupType29{29, "v"}, lookupType30{30, "v"}, lookupType31{31, "v"},
	lookupType32{32, "v"}, lookupType33{33, "v"}, lookupType34{34, "v"}, lookupType35{35, "v"},
	lookupType36{36, "v"}, lookupType37{37, "v"}, lookupType38{38, "v"}, lookupType39{39, "v"},
	lookupType40{40, "v"}, lookupType41{41, "v"}, lookupType42{42, "v"}, lookupType43{43, "v"},
	lookupType44{44, "v"}, lookupType45{45, "v"}, lookupType46{46, "v"}, lookupType47{47, "v"},
	lookupType48{48, "v"}, lookupType49{49, "v"}, lookupType50{50, "v"}, lookupType51{51, "v"},
	lookupType52{52, "v"}, lookupType53{53, "v"}, lookupType54{54, "v"}, lookupType55{55, "v"},
	lookupType56{56, "v"}, lookupType57{57, "v"}, lookupType58{58, "v"}, lookupType59{59, "v"},
	lookupType60{60, "v"}, lookupType61{61, "v"}, lookupType62{62, "v"}, lookupType63{63, "v"},
	lookupType64{64, "v"}, lookupType65{65, "v"}, lookupType66{66, "v"}, lookupType67{67, "v"},
	lookupType68{68, "v"}, lookupType69{69, "v"}, lookupType70{70, "v"}, lookupType71{71, "v"},
	lookupType72{72, "v"}, lookupType73{73, "v"}, lookupType74{74, "v"}, lookupType75{75, "v"},
	lookupType76{76, "v"}, lookupType77{77, "v"}, lookupType78{78, "v"}, lookupType79{79, "v"},
	lookupType80{80, "v"}, lookupType81{81, "v"}, lookupType82{82, "v"}, lookupType83{83, "v"},
	lookupType84{84, "v"}, lookupType85{85, "v"}, lookupType86{86, "v"}, lookupType87{87, "v"},
	lookupType88{88, "v"}, lookupType89{89, "v"}, lookupType90{90, "v"}, lookupType91{91, "v"},
	lookupType92{92, "v"}, lookupType93{93, "v"}, lookupType94{94, "v"}, lookupType95{95, "v"},
	lookupType96{96, "v"}, lookupType97{97, "v"}, lookupType98{98, "v"}, lookupType99{99, "v"},
	lookupType100{100, "v"}, lookupType101{101, "v"}, lookupType102{102, "v"}, lookupType103{103, "v"},
	lookupType104{104, "v"}, lookupType105{105, "v"}, lookupType106{106, "v"}, lookupType107{107, "v"},
	lookupType108{108, "v"}, lookupType109{109, "v"}, lookupType110{110, "v"}, lookupType111{111, "v"},
	lookupType112{112, "v"}, lookupType113{113, "v"}, lookupType114{114, "v"}, lookupType115{115, "v"},
	lookupType116{116, "v"}, lookupType117{117, "v"}, lookupType118{118, "v"}, lookupType119{119, "v"},
	lookupType120{120, "v"}, lookupType121{121, "v"}, lookupType122{122, "v"}, lookupType123{123, "v"},
	lookupType124{124, "v"}, lookupType125{125, "v"}, lookupType126{126, "v"}, lookupType127{127, "v"},
}

const lookupTypeCount = 128

func lookupTargets() []any {
	return []any{
		&lookupType0{}, &lookupType1{}, &lookupType2{}, &lookupType3{},
		&lookupType4{}, &lookupType5{}, &lookupType6{}, &lookupType7{},
		&lookupType8{}, &lookupType9{}, &lookupType10{}, &lookupType11{},
		&lookupType12{}, &lookupType13{}, &lookupType14{}, &lookupType15{},
		&lookupType16{}, &lookupType17{}, &lookupType18{}, &lookupType19{},
		&lookupType20{}, &lookupType21{}, &lookupType22{}, &lookupType23{},
		&lookupType24{}, &lookupType25{}, &lookupType26{}, &lookupType27{},
		&lookupType28{}, &lookupType29{}, &lookupType30{}, &lookupType31{},
		&lookupType32{}, &lookupType33{}, &lookupType34{}, &lookupType35{},
		&lookupType36{}, &lookupType37{}, &lookupType38{}, &lookupType39{},
		&lookupType40{}, &lookupType41{}, &lookupType42{}, &lookupType43{},
		&lookupType44{}, &lookupType45{}, &lookupType46{}, &lookupType47{},
		&lookupType48{}, &lookupType49{}, &lookupType50{}, &lookupType51{},
		&lookupType52{}, &lookupType53{}, &lookupType54{}, &lookupType55{},
		&lookupType56{}, &lookupType57{}, &lookupType58{}, &lookupType59{},
		&lookupType60{}, &lookupType61{}, &lookupType62{}, &lookupType63{},
		&lookupType64{}, &lookupType65{}, &lookupType66{}, &lookupType67{},
		&lookupType68{}, &lookupType69{}, &lookupType70{}, &lookupType71{},
		&lookupType72{}, &lookupType73{}, &lookupType74{}, &lookupType75{},
		&lookupType76{}, &lookupType77{}, &lookupType78{}, &lookupType79{},
		&lookupType80{}, &lookupType81{}, &lookupType82{}, &lookupType83{},
		&lookupType84{}, &lookupType85{}, &lookupType86{}, &lookupType87{},
		&lookupType88{}, &lookupType89{}, &lookupType90{}, &lookupType91{},
		&lookupType92{}, &lookupType93{}, &lookupType94{}, &lookupType95{},
		&lookupType96{}, &lookupType97{}, &lookupType98{}, &lookupType99{},
		&lookupType100{}, &lookupType101{}, &lookupType102{}, &lookupType103{},
		&lookupType104{}, &lookupType105{}, &lookupType106{}, &lookupType107{},
		&lookupType108{}, &lookupType109{}, &lookupType110{}, &lookupType111{},
		&lookupType112{}, &lookupType113{}, &lookupType114{}, &lookupType115{},
		&lookupType116{}, &lookupType117{}, &lookupType118{}, &lookupType119{},
		&lookupType120{}, &lookupType121{}, &lookupType122{}, &lookupType123{},
		&lookupType124{}, &lookupType125{}, &lookupType126{}, &lookupType127{},
	}
}

var lookupInput = []byte(`{"A":1,"B":"v"}`)

func BenchmarkTypeLookups(b *testing.B) {
	benchTypeLookups(b)
}

// BenchmarkTypeLookupsLargeProgram is BenchmarkTypeLookups in a program which encoded and decoded many other types
// before, which the tables shared by the goroutines hold: it is to be run in a process of its own, since the types
// stay in the tables.
func BenchmarkTypeLookupsLargeProgram(b *testing.B) {
	fillLookupTables(4096)
	benchTypeLookups(b)
}

// fillLookupTables encodes and decodes n types made by reflect, as a large program has.
func fillLookupTables(n int) {
	for i := 0; i < n; i++ {
		typ := reflect.StructOf([]reflect.StructField{
			{Name: "F" + strconv.Itoa(i), Type: reflect.TypeOf(0), Tag: `json:"f"`},
		})
		v := reflect.New(typ)
		if _, err := json.Marshal(v.Interface()); err != nil {
			panic(err)
		}
		if err := json.Unmarshal([]byte(`{"f":1}`), v.Interface()); err != nil {
			panic(err)
		}
	}
}

func benchTypeLookups(b *testing.B) {
	for _, k := range []int{1, 2, 8, 32, 128} {
		values := lookupValues[:k]
		b.Run(fmt.Sprintf("encode/top-level/%d types", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := json.Marshal(values[i%k]); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("decode/top-level/%d types", k), func(b *testing.B) {
			targets := lookupTargets()[:k]
			for i := 0; i < b.N; i++ {
				if err := json.Unmarshal(lookupInput, targets[i%k]); err != nil {
					b.Fatal(err)
				}
			}
			probeDecode(b.Name(), targets)
		})
	}
	for _, k := range []int{1, 6, 24, 96} {
		values := lookupValues[:k]
		b.Run(fmt.Sprintf("encode/interface/%d types", k), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := json.Marshal(values); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("encode/top-level/32 types/parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for i := 0; pb.Next(); i++ {
				if _, err := json.Marshal(lookupValues[i%32]); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
	b.Run("decode/top-level/32 types/parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			targets := lookupTargets()
			for i := 0; pb.Next(); i++ {
				if err := json.Unmarshal(lookupInput, targets[i%32]); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
	b.Run("encode/interface/24 types/parallel", func(b *testing.B) {
		values := lookupValues[:24]
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := json.Marshal(values); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
