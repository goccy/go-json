package encoder_test

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// The switch of Run of every VM must be compiled to a jump table: the opcode is dispatched by one indirect
// jump, which is what makes the switch faster than a call per opcode. The compiler does it for an integer switch
// of at least 8 cases whose values are dense, so it holds as long as the opcodes are numbered without gaps.
func TestVMDispatchIsJumpTable(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("the go command is not found")
	}
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		t.Skipf("the compiler doesn't use a jump table on %s", runtime.GOARCH)
	}
	for _, vm := range []string{"vm", "vm_indent", "vm_color", "vm_color_indent"} {
		t.Run(vm, func(t *testing.T) {
			cmd := exec.Command("go", "build", "-gcflags=-S", "github.com/goccy/go-json/internal/encoder/"+vm)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build: %v\n%s", err, out)
			}
			text := string(out)
			i := strings.Index(text, vm+".Run STEXT")
			if i < 0 {
				t.Fatal("the code of Run is not found")
			}
			run := text[i:]
			if j := strings.Index(run[20:], " STEXT "); j >= 0 {
				run = run[:20+j]
			}
			if !regexp.MustCompile(`Run\.jump\d+`).MatchString(run) {
				t.Fatal("the switch of Run is not compiled to a jump table")
			}
		})
	}
}
