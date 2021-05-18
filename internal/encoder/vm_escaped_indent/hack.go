package vm_escaped_indent

import (
	// HACK: compile order
	// `vm`, `vm_escaped`, `vm_indent`, `vm_escaped_indent`, `vm_debug` packages uses a lot of memory to compile,
	// so forcibly make dependencies and avoid compiling in concurrent.
	// dependency order: vm => vm_escaped => vm_indent => vm_escaped_indent => vm_debug
	_ "github.com/goccy/go-json/internal/encoder/vm_debug"
)
