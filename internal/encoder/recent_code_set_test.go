package encoder

import (
	"testing"
	"unsafe"
)

// The types whose addresses are hashed to the same set of the recent opcodes must not evict each other while
// two of them are encoded by turns, which is what a value of a type with values of interface{} in it does:
// which types share a set depends on where the binary has the types, so the ones found here are used.

type (
	recentType0   int
	recentType1   int
	recentType2   int
	recentType3   int
	recentType4   int
	recentType5   int
	recentType6   int
	recentType7   int
	recentType8   int
	recentType9   int
	recentType10  int
	recentType11  int
	recentType12  int
	recentType13  int
	recentType14  int
	recentType15  int
	recentType16  int
	recentType17  int
	recentType18  int
	recentType19  int
	recentType20  int
	recentType21  int
	recentType22  int
	recentType23  int
	recentType24  int
	recentType25  int
	recentType26  int
	recentType27  int
	recentType28  int
	recentType29  int
	recentType30  int
	recentType31  int
	recentType32  int
	recentType33  int
	recentType34  int
	recentType35  int
	recentType36  int
	recentType37  int
	recentType38  int
	recentType39  int
	recentType40  int
	recentType41  int
	recentType42  int
	recentType43  int
	recentType44  int
	recentType45  int
	recentType46  int
	recentType47  int
	recentType48  int
	recentType49  int
	recentType50  int
	recentType51  int
	recentType52  int
	recentType53  int
	recentType54  int
	recentType55  int
	recentType56  int
	recentType57  int
	recentType58  int
	recentType59  int
	recentType60  int
	recentType61  int
	recentType62  int
	recentType63  int
	recentType64  int
	recentType65  int
	recentType66  int
	recentType67  int
	recentType68  int
	recentType69  int
	recentType70  int
	recentType71  int
	recentType72  int
	recentType73  int
	recentType74  int
	recentType75  int
	recentType76  int
	recentType77  int
	recentType78  int
	recentType79  int
	recentType80  int
	recentType81  int
	recentType82  int
	recentType83  int
	recentType84  int
	recentType85  int
	recentType86  int
	recentType87  int
	recentType88  int
	recentType89  int
	recentType90  int
	recentType91  int
	recentType92  int
	recentType93  int
	recentType94  int
	recentType95  int
	recentType96  int
	recentType97  int
	recentType98  int
	recentType99  int
	recentType100 int
	recentType101 int
	recentType102 int
	recentType103 int
	recentType104 int
	recentType105 int
	recentType106 int
	recentType107 int
	recentType108 int
	recentType109 int
	recentType110 int
	recentType111 int
	recentType112 int
	recentType113 int
	recentType114 int
	recentType115 int
	recentType116 int
	recentType117 int
	recentType118 int
	recentType119 int
	recentType120 int
	recentType121 int
	recentType122 int
	recentType123 int
	recentType124 int
	recentType125 int
	recentType126 int
	recentType127 int
	recentType128 int
	recentType129 int
	recentType130 int
	recentType131 int
	recentType132 int
	recentType133 int
	recentType134 int
	recentType135 int
	recentType136 int
	recentType137 int
	recentType138 int
	recentType139 int
	recentType140 int
	recentType141 int
	recentType142 int
	recentType143 int
	recentType144 int
	recentType145 int
	recentType146 int
	recentType147 int
	recentType148 int
	recentType149 int
	recentType150 int
	recentType151 int
	recentType152 int
	recentType153 int
	recentType154 int
	recentType155 int
	recentType156 int
	recentType157 int
	recentType158 int
	recentType159 int
	recentType160 int
	recentType161 int
	recentType162 int
	recentType163 int
	recentType164 int
	recentType165 int
	recentType166 int
	recentType167 int
	recentType168 int
	recentType169 int
	recentType170 int
	recentType171 int
	recentType172 int
	recentType173 int
	recentType174 int
	recentType175 int
	recentType176 int
	recentType177 int
	recentType178 int
	recentType179 int
	recentType180 int
	recentType181 int
	recentType182 int
	recentType183 int
	recentType184 int
	recentType185 int
	recentType186 int
	recentType187 int
	recentType188 int
	recentType189 int
	recentType190 int
	recentType191 int
	recentType192 int
	recentType193 int
	recentType194 int
	recentType195 int
	recentType196 int
	recentType197 int
	recentType198 int
	recentType199 int
	recentType200 int
	recentType201 int
	recentType202 int
	recentType203 int
	recentType204 int
	recentType205 int
	recentType206 int
	recentType207 int
	recentType208 int
	recentType209 int
	recentType210 int
	recentType211 int
	recentType212 int
	recentType213 int
	recentType214 int
	recentType215 int
	recentType216 int
	recentType217 int
	recentType218 int
	recentType219 int
	recentType220 int
	recentType221 int
	recentType222 int
	recentType223 int
	recentType224 int
	recentType225 int
	recentType226 int
	recentType227 int
	recentType228 int
	recentType229 int
	recentType230 int
	recentType231 int
	recentType232 int
	recentType233 int
	recentType234 int
	recentType235 int
	recentType236 int
	recentType237 int
	recentType238 int
	recentType239 int
	recentType240 int
	recentType241 int
	recentType242 int
	recentType243 int
	recentType244 int
	recentType245 int
	recentType246 int
	recentType247 int
	recentType248 int
	recentType249 int
	recentType250 int
	recentType251 int
	recentType252 int
	recentType253 int
	recentType254 int
	recentType255 int
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
	recentType64(0), recentType65(0), recentType66(0), recentType67(0),
	recentType68(0), recentType69(0), recentType70(0), recentType71(0),
	recentType72(0), recentType73(0), recentType74(0), recentType75(0),
	recentType76(0), recentType77(0), recentType78(0), recentType79(0),
	recentType80(0), recentType81(0), recentType82(0), recentType83(0),
	recentType84(0), recentType85(0), recentType86(0), recentType87(0),
	recentType88(0), recentType89(0), recentType90(0), recentType91(0),
	recentType92(0), recentType93(0), recentType94(0), recentType95(0),
	recentType96(0), recentType97(0), recentType98(0), recentType99(0),
	recentType100(0), recentType101(0), recentType102(0), recentType103(0),
	recentType104(0), recentType105(0), recentType106(0), recentType107(0),
	recentType108(0), recentType109(0), recentType110(0), recentType111(0),
	recentType112(0), recentType113(0), recentType114(0), recentType115(0),
	recentType116(0), recentType117(0), recentType118(0), recentType119(0),
	recentType120(0), recentType121(0), recentType122(0), recentType123(0),
	recentType124(0), recentType125(0), recentType126(0), recentType127(0),
	recentType128(0), recentType129(0), recentType130(0), recentType131(0),
	recentType132(0), recentType133(0), recentType134(0), recentType135(0),
	recentType136(0), recentType137(0), recentType138(0), recentType139(0),
	recentType140(0), recentType141(0), recentType142(0), recentType143(0),
	recentType144(0), recentType145(0), recentType146(0), recentType147(0),
	recentType148(0), recentType149(0), recentType150(0), recentType151(0),
	recentType152(0), recentType153(0), recentType154(0), recentType155(0),
	recentType156(0), recentType157(0), recentType158(0), recentType159(0),
	recentType160(0), recentType161(0), recentType162(0), recentType163(0),
	recentType164(0), recentType165(0), recentType166(0), recentType167(0),
	recentType168(0), recentType169(0), recentType170(0), recentType171(0),
	recentType172(0), recentType173(0), recentType174(0), recentType175(0),
	recentType176(0), recentType177(0), recentType178(0), recentType179(0),
	recentType180(0), recentType181(0), recentType182(0), recentType183(0),
	recentType184(0), recentType185(0), recentType186(0), recentType187(0),
	recentType188(0), recentType189(0), recentType190(0), recentType191(0),
	recentType192(0), recentType193(0), recentType194(0), recentType195(0),
	recentType196(0), recentType197(0), recentType198(0), recentType199(0),
	recentType200(0), recentType201(0), recentType202(0), recentType203(0),
	recentType204(0), recentType205(0), recentType206(0), recentType207(0),
	recentType208(0), recentType209(0), recentType210(0), recentType211(0),
	recentType212(0), recentType213(0), recentType214(0), recentType215(0),
	recentType216(0), recentType217(0), recentType218(0), recentType219(0),
	recentType220(0), recentType221(0), recentType222(0), recentType223(0),
	recentType224(0), recentType225(0), recentType226(0), recentType227(0),
	recentType228(0), recentType229(0), recentType230(0), recentType231(0),
	recentType232(0), recentType233(0), recentType234(0), recentType235(0),
	recentType236(0), recentType237(0), recentType238(0), recentType239(0),
	recentType240(0), recentType241(0), recentType242(0), recentType243(0),
	recentType244(0), recentType245(0), recentType246(0), recentType247(0),
	recentType248(0), recentType249(0), recentType250(0), recentType251(0),
	recentType252(0), recentType253(0), recentType254(0), recentType255(0),
}

func typeptrOf(v interface{}) uintptr {
	return uintptr((*emptyInterface)(unsafe.Pointer(&v)).typ)
}

// typesOfSameSet returns the types among recentTypes which are hashed to the same set of the recent opcodes,
// three of them: it is next to impossible for none of the sets to have three of 256 types.
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
		recent := ctx.RecentCodeSet(typeptr)
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
			if ctx.RecentCodeSet(typeptr) == nil {
				t.Fatalf("type %#x was evicted by the other type of its set", typeptr)
			}
			if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
				t.Fatal(err)
			}
		}
	}
}
