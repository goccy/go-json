package vm

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
)

func DebugRun(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, p unsafe.Pointer) ([]byte, error) {
	defer func() {
		code := codeSet.TopCode(ctx.Option)
		if wc := ctx.Option.DebugDOTOut; wc != nil {
			_, _ = io.WriteString(wc, code.DumpDOT())
			wc.Close()
			ctx.Option.DebugDOTOut = nil
		}

		if err := recover(); err != nil {
			w := ctx.Option.DebugOut
			fmt.Fprintln(w, "=============[DEBUG]===============")
			fmt.Fprintln(w, "* [TYPE]")
			fmt.Fprintln(w, codeSet.Type)
			fmt.Fprintf(w, "\n")
			fmt.Fprintln(w, "* [ALL OPCODE]")
			fmt.Fprintln(w, code.Dump())
			fmt.Fprintf(w, "\n")
			fmt.Fprintln(w, "* [CONTEXT]")
			fmt.Fprintf(w, "%+v\n", ctx)
			fmt.Fprintln(w, "===================================")
			panic(err)
		}
	}()

	return Run(ctx, b, codeSet.TopCode(ctx.Option), p, 0)
}
