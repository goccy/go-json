package json

import "fmt"

type encodeOptimizerPlugin interface {
	optimize(*opcode) (*opcode, error)
}

type encodeOptimizer struct {
	optimizers []encodeOptimizerPlugin
}

func newEncodeOptimizer() *encodeOptimizer {
	return &encodeOptimizer{}
}

func (o *encodeOptimizer) addOptimizer(plg encodeOptimizerPlugin) {
	o.optimizers = append(o.optimizers, plg)
}

func (o *encodeOptimizer) optimize(code *opcode) (*opcode, error) {
	for _, optimizer := range o.optimizers {
		c, err := optimizer.optimize(code)
		if err != nil {
			return nil, err
		}
		code = c
	}
	return code, nil
}

type encodeStructHeadOptimizer struct {
}

func (o *encodeStructHeadOptimizer) optimize(code *opcode) (*opcode, error) {
	head := code
	for code.op != opEnd {
		if code.op == opStructFieldHead {
			fmt.Println("StructHead")
		}
		switch code.op.codeType() {
		case codeArrayElem, codeSliceElem:
			code = code.end
		case codeMapKey:
			code = code.end
		default:
			code = code.next
		}
	}
	return head, nil
}
