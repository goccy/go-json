package encoder

import (
	"context"
	"testing"
)

// Every type compiled once is found by SharedCodeSet, by turns with all the others and wherever the binary has
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
			if got := ctx.SharedCodeSet(typeptr); got != compiled[typeptr] {
				t.Fatalf("SharedCodeSet(%T) = %p, want %p", v, got, compiled[typeptr])
			}
		}
	}
}

// The opcodes with the fields ordered by the encoder are apart from the ones in the order of the struct, and a
// context which may filter the fields finds none, so that CompileToGetCodeSet filters them.
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
	if got := ctx.SharedCodeSet(typeptr); got == inOrder {
		t.Fatal("the opcodes with the fields ordered by the encoder are the ones in the order of the struct")
	}
	ordered, err := CompileToGetCodeSet(ctx, typeptr)
	if err != nil {
		t.Fatal(err)
	}
	if ordered == inOrder || ctx.SharedCodeSet(typeptr) != ordered {
		t.Fatal("the opcodes with the fields ordered by the encoder are not cached apart")
	}
	ctx.Option.Flag = 0
	if ctx.SharedCodeSet(typeptr) != inOrder {
		t.Fatal("the opcodes in the order of the struct are lost")
	}

	ctx.Option.Flag = ContextOption
	ctx.Option.Context = context.Background()
	if got := ctx.SharedCodeSet(typeptr); got != nil {
		t.Fatalf("SharedCodeSet with a context = %p, want nil", got)
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
			codeSet := ctx.SharedCodeSet(typeptr)
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
