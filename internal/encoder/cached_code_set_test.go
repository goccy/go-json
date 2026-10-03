package encoder

import (
	"context"
	"testing"
	"unsafe"
)

// The opcodes of a type are found by CachedCodeSet in the table shared by every goroutine, whatever other types
// the program encodes and wherever the binary has them: many types are used, so that some of them share a
// bucket of the table.

type (
	cachedType0   int
	cachedType1   int
	cachedType2   int
	cachedType3   int
	cachedType4   int
	cachedType5   int
	cachedType6   int
	cachedType7   int
	cachedType8   int
	cachedType9   int
	cachedType10  int
	cachedType11  int
	cachedType12  int
	cachedType13  int
	cachedType14  int
	cachedType15  int
	cachedType16  int
	cachedType17  int
	cachedType18  int
	cachedType19  int
	cachedType20  int
	cachedType21  int
	cachedType22  int
	cachedType23  int
	cachedType24  int
	cachedType25  int
	cachedType26  int
	cachedType27  int
	cachedType28  int
	cachedType29  int
	cachedType30  int
	cachedType31  int
	cachedType32  int
	cachedType33  int
	cachedType34  int
	cachedType35  int
	cachedType36  int
	cachedType37  int
	cachedType38  int
	cachedType39  int
	cachedType40  int
	cachedType41  int
	cachedType42  int
	cachedType43  int
	cachedType44  int
	cachedType45  int
	cachedType46  int
	cachedType47  int
	cachedType48  int
	cachedType49  int
	cachedType50  int
	cachedType51  int
	cachedType52  int
	cachedType53  int
	cachedType54  int
	cachedType55  int
	cachedType56  int
	cachedType57  int
	cachedType58  int
	cachedType59  int
	cachedType60  int
	cachedType61  int
	cachedType62  int
	cachedType63  int
	cachedType64  int
	cachedType65  int
	cachedType66  int
	cachedType67  int
	cachedType68  int
	cachedType69  int
	cachedType70  int
	cachedType71  int
	cachedType72  int
	cachedType73  int
	cachedType74  int
	cachedType75  int
	cachedType76  int
	cachedType77  int
	cachedType78  int
	cachedType79  int
	cachedType80  int
	cachedType81  int
	cachedType82  int
	cachedType83  int
	cachedType84  int
	cachedType85  int
	cachedType86  int
	cachedType87  int
	cachedType88  int
	cachedType89  int
	cachedType90  int
	cachedType91  int
	cachedType92  int
	cachedType93  int
	cachedType94  int
	cachedType95  int
	cachedType96  int
	cachedType97  int
	cachedType98  int
	cachedType99  int
	cachedType100 int
	cachedType101 int
	cachedType102 int
	cachedType103 int
	cachedType104 int
	cachedType105 int
	cachedType106 int
	cachedType107 int
	cachedType108 int
	cachedType109 int
	cachedType110 int
	cachedType111 int
	cachedType112 int
	cachedType113 int
	cachedType114 int
	cachedType115 int
	cachedType116 int
	cachedType117 int
	cachedType118 int
	cachedType119 int
	cachedType120 int
	cachedType121 int
	cachedType122 int
	cachedType123 int
	cachedType124 int
	cachedType125 int
	cachedType126 int
	cachedType127 int
	cachedType128 int
	cachedType129 int
	cachedType130 int
	cachedType131 int
	cachedType132 int
	cachedType133 int
	cachedType134 int
	cachedType135 int
	cachedType136 int
	cachedType137 int
	cachedType138 int
	cachedType139 int
	cachedType140 int
	cachedType141 int
	cachedType142 int
	cachedType143 int
	cachedType144 int
	cachedType145 int
	cachedType146 int
	cachedType147 int
	cachedType148 int
	cachedType149 int
	cachedType150 int
	cachedType151 int
	cachedType152 int
	cachedType153 int
	cachedType154 int
	cachedType155 int
	cachedType156 int
	cachedType157 int
	cachedType158 int
	cachedType159 int
	cachedType160 int
	cachedType161 int
	cachedType162 int
	cachedType163 int
	cachedType164 int
	cachedType165 int
	cachedType166 int
	cachedType167 int
	cachedType168 int
	cachedType169 int
	cachedType170 int
	cachedType171 int
	cachedType172 int
	cachedType173 int
	cachedType174 int
	cachedType175 int
	cachedType176 int
	cachedType177 int
	cachedType178 int
	cachedType179 int
	cachedType180 int
	cachedType181 int
	cachedType182 int
	cachedType183 int
	cachedType184 int
	cachedType185 int
	cachedType186 int
	cachedType187 int
	cachedType188 int
	cachedType189 int
	cachedType190 int
	cachedType191 int
	cachedType192 int
	cachedType193 int
	cachedType194 int
	cachedType195 int
	cachedType196 int
	cachedType197 int
	cachedType198 int
	cachedType199 int
	cachedType200 int
	cachedType201 int
	cachedType202 int
	cachedType203 int
	cachedType204 int
	cachedType205 int
	cachedType206 int
	cachedType207 int
	cachedType208 int
	cachedType209 int
	cachedType210 int
	cachedType211 int
	cachedType212 int
	cachedType213 int
	cachedType214 int
	cachedType215 int
	cachedType216 int
	cachedType217 int
	cachedType218 int
	cachedType219 int
	cachedType220 int
	cachedType221 int
	cachedType222 int
	cachedType223 int
	cachedType224 int
	cachedType225 int
	cachedType226 int
	cachedType227 int
	cachedType228 int
	cachedType229 int
	cachedType230 int
	cachedType231 int
	cachedType232 int
	cachedType233 int
	cachedType234 int
	cachedType235 int
	cachedType236 int
	cachedType237 int
	cachedType238 int
	cachedType239 int
	cachedType240 int
	cachedType241 int
	cachedType242 int
	cachedType243 int
	cachedType244 int
	cachedType245 int
	cachedType246 int
	cachedType247 int
	cachedType248 int
	cachedType249 int
	cachedType250 int
	cachedType251 int
	cachedType252 int
	cachedType253 int
	cachedType254 int
	cachedType255 int
)

var cachedTypes = []any{
	cachedType0(0), cachedType1(0), cachedType2(0), cachedType3(0),
	cachedType4(0), cachedType5(0), cachedType6(0), cachedType7(0),
	cachedType8(0), cachedType9(0), cachedType10(0), cachedType11(0),
	cachedType12(0), cachedType13(0), cachedType14(0), cachedType15(0),
	cachedType16(0), cachedType17(0), cachedType18(0), cachedType19(0),
	cachedType20(0), cachedType21(0), cachedType22(0), cachedType23(0),
	cachedType24(0), cachedType25(0), cachedType26(0), cachedType27(0),
	cachedType28(0), cachedType29(0), cachedType30(0), cachedType31(0),
	cachedType32(0), cachedType33(0), cachedType34(0), cachedType35(0),
	cachedType36(0), cachedType37(0), cachedType38(0), cachedType39(0),
	cachedType40(0), cachedType41(0), cachedType42(0), cachedType43(0),
	cachedType44(0), cachedType45(0), cachedType46(0), cachedType47(0),
	cachedType48(0), cachedType49(0), cachedType50(0), cachedType51(0),
	cachedType52(0), cachedType53(0), cachedType54(0), cachedType55(0),
	cachedType56(0), cachedType57(0), cachedType58(0), cachedType59(0),
	cachedType60(0), cachedType61(0), cachedType62(0), cachedType63(0),
	cachedType64(0), cachedType65(0), cachedType66(0), cachedType67(0),
	cachedType68(0), cachedType69(0), cachedType70(0), cachedType71(0),
	cachedType72(0), cachedType73(0), cachedType74(0), cachedType75(0),
	cachedType76(0), cachedType77(0), cachedType78(0), cachedType79(0),
	cachedType80(0), cachedType81(0), cachedType82(0), cachedType83(0),
	cachedType84(0), cachedType85(0), cachedType86(0), cachedType87(0),
	cachedType88(0), cachedType89(0), cachedType90(0), cachedType91(0),
	cachedType92(0), cachedType93(0), cachedType94(0), cachedType95(0),
	cachedType96(0), cachedType97(0), cachedType98(0), cachedType99(0),
	cachedType100(0), cachedType101(0), cachedType102(0), cachedType103(0),
	cachedType104(0), cachedType105(0), cachedType106(0), cachedType107(0),
	cachedType108(0), cachedType109(0), cachedType110(0), cachedType111(0),
	cachedType112(0), cachedType113(0), cachedType114(0), cachedType115(0),
	cachedType116(0), cachedType117(0), cachedType118(0), cachedType119(0),
	cachedType120(0), cachedType121(0), cachedType122(0), cachedType123(0),
	cachedType124(0), cachedType125(0), cachedType126(0), cachedType127(0),
	cachedType128(0), cachedType129(0), cachedType130(0), cachedType131(0),
	cachedType132(0), cachedType133(0), cachedType134(0), cachedType135(0),
	cachedType136(0), cachedType137(0), cachedType138(0), cachedType139(0),
	cachedType140(0), cachedType141(0), cachedType142(0), cachedType143(0),
	cachedType144(0), cachedType145(0), cachedType146(0), cachedType147(0),
	cachedType148(0), cachedType149(0), cachedType150(0), cachedType151(0),
	cachedType152(0), cachedType153(0), cachedType154(0), cachedType155(0),
	cachedType156(0), cachedType157(0), cachedType158(0), cachedType159(0),
	cachedType160(0), cachedType161(0), cachedType162(0), cachedType163(0),
	cachedType164(0), cachedType165(0), cachedType166(0), cachedType167(0),
	cachedType168(0), cachedType169(0), cachedType170(0), cachedType171(0),
	cachedType172(0), cachedType173(0), cachedType174(0), cachedType175(0),
	cachedType176(0), cachedType177(0), cachedType178(0), cachedType179(0),
	cachedType180(0), cachedType181(0), cachedType182(0), cachedType183(0),
	cachedType184(0), cachedType185(0), cachedType186(0), cachedType187(0),
	cachedType188(0), cachedType189(0), cachedType190(0), cachedType191(0),
	cachedType192(0), cachedType193(0), cachedType194(0), cachedType195(0),
	cachedType196(0), cachedType197(0), cachedType198(0), cachedType199(0),
	cachedType200(0), cachedType201(0), cachedType202(0), cachedType203(0),
	cachedType204(0), cachedType205(0), cachedType206(0), cachedType207(0),
	cachedType208(0), cachedType209(0), cachedType210(0), cachedType211(0),
	cachedType212(0), cachedType213(0), cachedType214(0), cachedType215(0),
	cachedType216(0), cachedType217(0), cachedType218(0), cachedType219(0),
	cachedType220(0), cachedType221(0), cachedType222(0), cachedType223(0),
	cachedType224(0), cachedType225(0), cachedType226(0), cachedType227(0),
	cachedType228(0), cachedType229(0), cachedType230(0), cachedType231(0),
	cachedType232(0), cachedType233(0), cachedType234(0), cachedType235(0),
	cachedType236(0), cachedType237(0), cachedType238(0), cachedType239(0),
	cachedType240(0), cachedType241(0), cachedType242(0), cachedType243(0),
	cachedType244(0), cachedType245(0), cachedType246(0), cachedType247(0),
	cachedType248(0), cachedType249(0), cachedType250(0), cachedType251(0),
	cachedType252(0), cachedType253(0), cachedType254(0), cachedType255(0),
}

func typeptrOf(v any) uintptr {
	return uintptr((*emptyInterface)(unsafe.Pointer(&v)).typ)
}

// Every type compiled once is found by CachedCodeSet, by turns with all the others, as the opcodes which
// CompileToGetCodeSet returned.
func TestCachedCodeSetFindsEveryCompiledType(t *testing.T) {
	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)
	compiled := map[uintptr]*OpcodeSet{}
	for _, v := range cachedTypes {
		typeptr := typeptrOf(v)
		codeSet, err := CompileToGetCodeSet(ctx, typeptr)
		if err != nil {
			t.Fatal(err)
		}
		compiled[typeptr] = codeSet
	}
	for round := 0; round < 2; round++ {
		for _, v := range cachedTypes {
			typeptr := typeptrOf(v)
			if got := ctx.CachedCodeSet(typeptr); got != compiled[typeptr] {
				t.Fatalf("CachedCodeSet(%T) = %p, want %p", v, got, compiled[typeptr])
			}
		}
	}
}

// The opcodes with the fields ordered by the encoder are apart from the ones in the order of the struct, and a
// context which may filter the fields finds none, so that CompileToGetCodeSet filters them.
func TestCachedCodeSetByOptions(t *testing.T) {
	type orderedFields struct {
		A string
		B int
		C string
	}
	typeptr := typeptrOf(orderedFields{})
	ctx := TakeRuntimeContext()
	defer ReleaseRuntimeContext(ctx)

	ctx.Option.Flag = 0
	inOrder, err := CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Option.Flag = OptimizeFieldOrderOption
	if got := ctx.CachedCodeSet(typeptr); got == inOrder {
		t.Fatal("the opcodes with the fields ordered by the encoder are the ones in the order of the struct")
	}
	ordered, err := CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		t.Fatal(err)
	}
	if ordered == inOrder || ctx.CachedCodeSet(typeptr) != ordered {
		t.Fatal("the opcodes with the fields ordered by the encoder are not cached apart")
	}
	ctx.Option.Flag = 0
	if ctx.CachedCodeSet(typeptr) != inOrder {
		t.Fatal("the opcodes in the order of the struct are lost")
	}

	ctx.Option.Flag = ContextOption
	ctx.Option.Context = context.Background()
	if got := ctx.CachedCodeSet(typeptr); got != nil {
		t.Fatalf("CachedCodeSet with a context = %p, want nil", got)
	}
	if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
		t.Fatal(err)
	}
}
