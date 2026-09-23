package encoder

import (
	"testing"
	"unsafe"
)

// The types whose addresses are hashed to the same set of the recent opcodes must not evict each other while
// two of them are encoded by turns, which is what a value of a type with values of interface{} in it does:
// which types share a set depends on where the binary has the types, so the ones found here are used.

type (
	recentType0  int
	recentType1  int
	recentType2  int
	recentType3  int
	recentType4  int
	recentType5  int
	recentType6  int
	recentType7  int
	recentType8  int
	recentType9  int
	recentType10 int
	recentType11 int
	recentType12 int
	recentType13 int
	recentType14 int
	recentType15 int
	recentType16 int
	recentType17 int
	recentType18 int
	recentType19 int
	recentType20 int
	recentType21 int
	recentType22 int
	recentType23 int
	recentType24 int
	recentType25 int
	recentType26 int
	recentType27 int
	recentType28 int
	recentType29 int
	recentType30 int
	recentType31 int
	recentType32 int
	recentType33 int
	recentType34 int
	recentType35 int
	recentType36 int
	recentType37 int
	recentType38 int
	recentType39 int
	recentType40 int
	recentType41 int
	recentType42 int
	recentType43 int
	recentType44 int
	recentType45 int
	recentType46 int
	recentType47 int
	recentType48 int
	recentType49 int
	recentType50 int
	recentType51 int
	recentType52 int
	recentType53 int
	recentType54 int
	recentType55 int
	recentType56 int
	recentType57 int
	recentType58 int
	recentType59 int
	recentType60 int
	recentType61 int
	recentType62 int
	recentType63 int
)

var recentTypes = []interface{}{
	recentType0(0), recentType1(0), recentType2(0), recentType3(0),
	recentType4(0), recentType5(0), recentType6(0), recentType7(0),
	recentType8(0), recentType9(0), recentType10(0), recentType11(0),
	recentType12(0), recentType13(0), recentType14(0), recentType15(0),
	recentType16(0), recentType17(0), recentType18(0), recentType19(0),
	recentType20(0), recentType21(0), recentType22(0), recentType23(0),
	recentType24(0), recentType25(0), recentType26(0), recentType27(0),
	recentType28(0), recentType29(0), recentType30(0), recentType31(0),
	recentType32(0), recentType33(0), recentType34(0), recentType35(0),
	recentType36(0), recentType37(0), recentType38(0), recentType39(0),
	recentType40(0), recentType41(0), recentType42(0), recentType43(0),
	recentType44(0), recentType45(0), recentType46(0), recentType47(0),
	recentType48(0), recentType49(0), recentType50(0), recentType51(0),
	recentType52(0), recentType53(0), recentType54(0), recentType55(0),
	recentType56(0), recentType57(0), recentType58(0), recentType59(0),
	recentType60(0), recentType61(0), recentType62(0), recentType63(0),
}

func typeptrOf(v interface{}) uintptr {
	return uintptr((*emptyInterface)(unsafe.Pointer(&v)).typ)
}

// typesOfSameSet returns the types among recentTypes which are hashed to the same set of the recent opcodes,
// three of them: it is next to impossible for none of the sets to have three of 64 types.
func typesOfSameSet(t *testing.T) []uintptr {
	t.Helper()
	bySet := map[uint64][]uintptr{}
	for _, v := range recentTypes {
		typeptr := typeptrOf(v)
		set := recentCodeSetIndex(typeptr)
		bySet[set] = append(bySet[set], typeptr)
		if len(bySet[set]) == 3 {
			return bySet[set]
		}
	}
	t.Skip("no three types of the same set in this binary")
	return nil
}

func TestRecentCodeSetsHoldTwoTypesOfASet(t *testing.T) {
	typeptrs := typesOfSameSet(t)
	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)
	codeSets := make([]*OpcodeSet, len(typeptrs))
	for i, typeptr := range typeptrs {
		codeSet, err := CompileToGetCodeSet(ctx, typeptr)
		if err != nil {
			t.Fatal(err)
		}
		codeSets[i] = codeSet
	}
	// the two types encoded last are in the set, and the one before them is evicted.
	for i, typeptr := range typeptrs {
		recent := ctx.recentCodeSet(typeptr)
		if i == 0 {
			if recent != nil {
				t.Fatalf("the type encoded before the last two is still in the set")
			}
			continue
		}
		if recent != codeSets[i] {
			t.Fatalf("the type encoded last but %d is not in the set", len(typeptrs)-1-i)
		}
	}
	// two types encoded by turns stay: neither is compiled or looked up in the shared table again.
	for i := 0; i < 4; i++ {
		for _, typeptr := range typeptrs[1:] {
			if ctx.recentCodeSet(typeptr) == nil {
				t.Fatalf("type %#x was evicted by the other type of its set", typeptr)
			}
			if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
				t.Fatal(err)
			}
		}
	}
}
