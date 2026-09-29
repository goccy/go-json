package json_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// deadCodeProgram uses the encoders and the decoders of go-json on values with marshalers and unmarshalers.
const deadCodeProgram = `package main

import (
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/goccy/go-json"
)

type stamp struct{ time.Time }

type record struct {
	At    time.Time
	Stamp stamp
	Addr  netip.Addr
	Any   any
	Map   map[string]any
}

func main() {
	v := record{At: time.Unix(0, 0).UTC(), Addr: netip.MustParseAddr("::1"), Any: 1, Map: map[string]any{"a": 1}}
	b, err := json.Marshal(v)
	fmt.Println(string(b), err)
	b, err = json.MarshalIndent(v, "", " ")
	fmt.Println(string(b), err)
	b, err = json.MarshalContext(context.Background(), v)
	fmt.Println(string(b), err)
	var buf bytes.Buffer
	fmt.Println(json.NewEncoder(&buf).Encode(v))
	var r record
	fmt.Println(json.Unmarshal(b, &r), r)
	fmt.Println(json.NewDecoder(&buf).Decode(&r))
}
`

// go-json looks the methods of the marshalers up by reflect with their names as constants only: a lookup by a name
// the compiler can't see as a constant makes the linker keep every exported method of every type of a program
// which imports go-json ( see runtime.MethodLookup ). The linker marks the functions with such a lookup as
// <ReflectMethod> in the dependencies it dumps.
func TestNoReflectMethodLookupByVariableName(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a program")
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	goMod := "module deadcode\n\ngo 1.21\n\nrequire github.com/goccy/go-json v0.0.0\n\nreplace github.com/goccy/go-json => " +
		filepath.ToSlash(root) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(deadCodeProgram), 0o600); err != nil {
		t.Fatal(err)
	}
	goCmd, err := exec.LookPath("go")
	if err != nil {
		t.Skip("the go command is not found")
	}
	cmd := exec.Command(goCmd, "build", "-ldflags=-dumpdep", "-o", filepath.Join(dir, "program"), ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	marked := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		// the dump is "caller -> callee", each followed by <ReflectMethod> if it is marked.
		for _, fn := range strings.Split(line, " -> ") {
			if name, ok := strings.CutSuffix(strings.TrimSpace(fn), " <ReflectMethod>"); ok && strings.HasPrefix(name, "github.com/goccy/go-json") {
				marked[name] = true
			}
		}
	}
	for name := range marked {
		t.Errorf("%s looks a method up by reflect with a name which is not a constant", name)
	}
}
