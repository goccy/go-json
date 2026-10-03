package encoder

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/goccy/go-json/internal/runtime"
)

// The shapes of the table of the recent opcodes of a context, measured against each other: the number of
// the sets, which the address of a type is hashed to, and the number of the entries of a set, which hold
// the types of the set encoded last. A lookup which misses goes to the table shared by every goroutine, as
// CompileToGetCodeSet does, so a miss costs what it costs the encoder.
//
// The patterns are the types encoded by turns in a Marshal: the type passed to Marshal and the types held by
// its values of interface{}. The types of a pattern are in the same set or in different sets, as their
// addresses in a binary may have them.

type recentVariantEntry struct {
	typeptr uintptr
	codeSet *OpcodeSet
}

// recentVariantTable is a table of the recent opcodes of a shape. A lookup compares the entries of the set of
// the type in their order; an insertion puts the type first and drops the last entry of the set.
type recentVariantTable interface {
	lookup(typeptr uintptr) *OpcodeSet
	insert(typeptr uintptr, codeSet *OpcodeSet)
	shape() string
}

type recentVariant16x1 [16][1]recentVariantEntry
type recentVariant16x2 [16][2]recentVariantEntry
type recentVariant16x4 [16][4]recentVariantEntry
type recentVariant32x2 [32][2]recentVariantEntry
type recentVariant64x2 [64][2]recentVariantEntry
type recentVariant32x4 [32][4]recentVariantEntry

func recentVariantIndex(typeptr uintptr, bits uint) uint64 {
	return (uint64(typeptr) * runtime.TypeHashMultiplier) >> (64 - bits)
}

func (t *recentVariant16x1) shape() string { return "16x1" }
func (t *recentVariant16x1) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 4)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	return nil
}
func (t *recentVariant16x1) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 4)]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

func (t *recentVariant16x2) shape() string { return "16x2" }
func (t *recentVariant16x2) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 4)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	if set[1].typeptr == typeptr {
		return set[1].codeSet
	}
	return nil
}
func (t *recentVariant16x2) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 4)]
	set[1] = set[0]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

func (t *recentVariant32x2) shape() string { return "32x2" }
func (t *recentVariant32x2) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 5)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	if set[1].typeptr == typeptr {
		return set[1].codeSet
	}
	return nil
}
func (t *recentVariant32x2) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 5)]
	set[1] = set[0]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

func (t *recentVariant64x2) shape() string { return "64x2" }
func (t *recentVariant64x2) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 6)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	if set[1].typeptr == typeptr {
		return set[1].codeSet
	}
	return nil
}
func (t *recentVariant64x2) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 6)]
	set[1] = set[0]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

func (t *recentVariant16x4) shape() string { return "16x4" }
func (t *recentVariant16x4) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 4)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	if set[1].typeptr == typeptr {
		return set[1].codeSet
	}
	if set[2].typeptr == typeptr {
		return set[2].codeSet
	}
	if set[3].typeptr == typeptr {
		return set[3].codeSet
	}
	return nil
}
func (t *recentVariant16x4) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 4)]
	set[3], set[2], set[1] = set[2], set[1], set[0]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

func (t *recentVariant32x4) shape() string { return "32x4" }
func (t *recentVariant32x4) lookup(typeptr uintptr) *OpcodeSet {
	set := &t[recentVariantIndex(typeptr, 5)]
	if set[0].typeptr == typeptr {
		return set[0].codeSet
	}
	if set[1].typeptr == typeptr {
		return set[1].codeSet
	}
	if set[2].typeptr == typeptr {
		return set[2].codeSet
	}
	if set[3].typeptr == typeptr {
		return set[3].codeSet
	}
	return nil
}
func (t *recentVariant32x4) insert(typeptr uintptr, codeSet *OpcodeSet) {
	set := &t[recentVariantIndex(typeptr, 5)]
	set[3], set[2], set[1] = set[2], set[1], set[0]
	set[0] = recentVariantEntry{typeptr, codeSet}
}

// recentVariantTypes returns n addresses of types which are in the set of the first one in a table of the
// bits ( sameSet ), or in sets of their own ( !sameSet ): the smaller the table, the more likely the same set.
// They are hashed to the same set of every table with fewer bits too, as the set is the top bits of the hash.
func recentVariantTypes(n int, bits uint, sameSet bool) []uintptr {
	r := rand.New(rand.NewSource(1))
	// what the addresses of the types of a binary look like: 8 byte aligned, in a range of a few MB.
	random := func() uintptr { return 0x1000000 + uintptr(r.Intn(1<<20))*8 }
	typeptrs := []uintptr{random()}
	seen := map[uint64]bool{recentVariantIndex(typeptrs[0], bits): true}
	for len(typeptrs) < n {
		typeptr := random()
		set := recentVariantIndex(typeptr, bits)
		if sameSet != (set == recentVariantIndex(typeptrs[0], bits)) || (!sameSet && seen[set]) {
			continue
		}
		seen[set] = true
		typeptrs = append(typeptrs, typeptr)
	}
	return typeptrs
}

// the patterns of the types encoded by turns: their number, and whether they are in the same set of the
// table of the most sets measured ( then they are in the same set of every table ).
var recentVariantPatterns = []struct {
	types   int
	sameSet bool
}{
	{1, false},
	{2, false},
	{2, true},
	{3, false},
	{3, true},
	{4, true},
	{8, false},
}

func BenchmarkVariant_RecentCodeSets(b *testing.B) {
	tables := []recentVariantTable{
		&recentVariant16x1{},
		&recentVariant16x2{}, &recentVariant32x2{}, &recentVariant64x2{},
		&recentVariant16x4{}, &recentVariant32x4{},
	}
	for _, pattern := range recentVariantPatterns {
		typeptrs := recentVariantTypes(pattern.types, 6, pattern.sameSet)
		// the table shared by every goroutine, which a miss goes to: it has the types, compiled.
		var shared runtime.TypeCache[OpcodeSet]
		for _, typeptr := range typeptrs {
			shared.Store(typeptr, &OpcodeSet{})
		}
		// the sequence of the lookups of a Marshal: the type passed to it, then each of the others, and the
		// first one again for the next Marshal.
		sequence := make([]uintptr, 0, 64)
		for len(sequence) < 64 {
			sequence = append(sequence, typeptrs...)
		}
		set := "separate sets"
		if pattern.sameSet {
			set = "same set"
		}
		for _, table := range tables {
			b.Run(fmt.Sprintf("%s/%d types %s", table.shape(), pattern.types, set), func(b *testing.B) {
				var hits, misses int
				for i := 0; i < b.N; i++ {
					for _, typeptr := range sequence {
						codeSet := table.lookup(typeptr)
						if codeSet == nil {
							codeSet = shared.Load(typeptr)
							table.insert(typeptr, codeSet)
							misses++
						} else {
							hits++
						}
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(sequence)), "ns/lookup")
				b.ReportMetric(float64(misses)/float64(hits+misses)*100, "miss%")
			})
		}
	}
}
