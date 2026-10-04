package encoder

import (
	"context"
	"fmt"
	"testing"
)

// Every type compiled once is found by SharedCodeSets, by turns with all the others and wherever the binary has
// them, as the opcodes which CompileToGetCodeSet returned: so a type which its set in a context doesn't hold is
// found without a compilation.
func TestSharedCodeSetFindsEveryCompiledType(t *testing.T) {
	ctx := TakeRuntimeContext()
	ctx.Option.Flag, ctx.Option.Context = 0, nil
	defer func() {
		ctx.Option.Flag, ctx.Option.Context = 0, nil
		ReleaseRuntimeContext(ctx)
	}()
	compiled := map[uintptr]*OpcodeSet{}
	for _, v := range recentTypes {
		typeptr := typeptrOf(v)
		codeSet, err := CompileToGetCodeSet(ctx, typeptr)
		if err != nil {
			t.Fatal(err)
		}
		compiled[typeptr] = codeSet
	}
	for round := 0; round < 2; round++ {
		for _, v := range recentTypes {
			typeptr := typeptrOf(v)
			if got := ctx.SharedCodeSets().Load(typeptr); got != compiled[typeptr] {
				t.Fatalf("SharedCodeSets().Load(%T) = %p, want %p", v, got, compiled[typeptr])
			}
		}
	}
}

// The opcodes with the fields ordered by the encoder are apart from the ones in the order of the struct, and a
// context which may filter the fields finds none in its set, so that CompileToGetCodeSet filters them.
func TestSharedCodeSetByOptions(t *testing.T) {
	type orderedFields struct {
		A string
		B int
		C string
	}
	typeptr := typeptrOf(orderedFields{})
	ctx := TakeRuntimeContext()
	ctx.Option.Flag, ctx.Option.Context = 0, nil
	defer func() {
		ctx.Option.Flag, ctx.Option.Context = 0, nil
		ReleaseRuntimeContext(ctx)
	}()

	ctx.Option.Flag = 0
	inOrder, err := CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Option.Flag = OptimizeFieldOrderOption
	if got := ctx.SharedCodeSets().Load(typeptr); got == inOrder {
		t.Fatal("the opcodes with the fields ordered by the encoder are the ones in the order of the struct")
	}
	ordered, err := CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		t.Fatal(err)
	}
	if ordered == inOrder || ctx.SharedCodeSets().Load(typeptr) != ordered {
		t.Fatal("the opcodes with the fields ordered by the encoder are not cached apart")
	}
	ctx.Option.Flag = 0
	if ctx.SharedCodeSets().Load(typeptr) != inOrder {
		t.Fatal("the opcodes in the order of the struct are lost")
	}

	ctx.Option.Flag = ContextOption
	ctx.Option.Context = context.Background()
	if got := ctx.RecentCodeSet(typeptr); got != nil {
		t.Fatalf("RecentCodeSet with a context = %p, want nil", got)
	}
	if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
		t.Fatal(err)
	}
}

// Three types of a set encoded by turns evict each other from the context, and each is found in the shared table
// then, and taken back into its set.
func TestRecentCodeSetsTakeBackTheTypesOfTheSharedTable(t *testing.T) {
	typeptrs := typesOfSameSet(t)
	ctx := TakeRuntimeContext()
	ctx.Option.Flag, ctx.Option.Context = 0, nil
	defer func() {
		ctx.Option.Flag, ctx.Option.Context = 0, nil
		ReleaseRuntimeContext(ctx)
	}()
	for _, typeptr := range typeptrs {
		if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
			t.Fatal(err)
		}
	}
	for round := 0; round < 3; round++ {
		for _, typeptr := range typeptrs {
			if ctx.RecentCodeSet(typeptr) != nil {
				continue
			}
			codeSet := ctx.SharedCodeSets().Load(typeptr)
			if codeSet == nil {
				t.Fatalf("type %#x is neither in its set nor in the shared table", typeptr)
			}
			ctx.RememberCodeSet(typeptr, codeSet)
			if ctx.RecentCodeSet(typeptr) != codeSet {
				t.Fatalf("type %#x is not taken into its set", typeptr)
			}
		}
	}
}

// typesOfSet returns n types among recentTypes which are hashed to the same set of the recent opcodes, or to n
// different sets.
func typesOfSet(tb testing.TB, n int, sameSet bool) []uintptr {
	tb.Helper()
	bySet := map[uint64][]uintptr{}
	var separate []uintptr
	for _, v := range recentTypes {
		typeptr := typeptrOf(v)
		set := recentCodeSetIndex(typeptr)
		bySet[set] = append(bySet[set], typeptr)
		if sameSet && len(bySet[set]) == n {
			return bySet[set]
		}
		if !sameSet && len(bySet[set]) == 1 {
			if separate = append(separate, typeptr); len(separate) == n {
				return separate
			}
		}
	}
	tb.Skipf("no %d types of the same set or of different sets in this binary", n)
	return nil
}

// BenchmarkRecentCodeSetsByTurns looks up the opcodes of types by turns as the VM looks up the types of the
// values of interface{}: in the context, then in the shared table. The types are of one set, which is where the
// binary has them, or of different sets: three or more of one set evict each other from the context.
func BenchmarkRecentCodeSetsByTurns(b *testing.B) {
	for _, c := range []struct {
		n       int
		sameSet bool
		// many: the first n types of recentTypes, over the sets as the binary has them.
		many bool
	}{{2, true, false}, {3, true, false}, {4, true, false}, {6, true, false}, {7, true, false}, {6, false, false},
		{12, false, true}, {24, false, true}, {48, false, true}} {
		name := fmt.Sprintf("%d types of different sets", c.n)
		if c.sameSet {
			name = fmt.Sprintf("%d types of one set", c.n)
		}
		if c.many {
			name = fmt.Sprintf("%d types", c.n)
		}
		b.Run(name, func(b *testing.B) {
			var typeptrs []uintptr
			if c.many {
				for _, v := range recentTypes[:c.n] {
					typeptrs = append(typeptrs, typeptrOf(v))
				}
			} else {
				typeptrs = typesOfSet(b, c.n, c.sameSet)
			}
			ctx := TakeRuntimeContext()
			defer ReleaseRuntimeContext(ctx)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, typeptr := range typeptrs {
					if ctx.RecentCodeSet(typeptr) != nil {
						continue
					}
					if codeSet := ctx.SharedCodeSets().LoadFirst(typeptr); codeSet != nil {
						ctx.RememberCodeSet(typeptr, codeSet)
						continue
					}
					if _, err := CompileToGetCodeSet(ctx, typeptr); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
