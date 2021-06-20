package decoder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

// PathString represents JSON Path
//
// JSON Path rule
// .     : child operator
// ..    : recursive descent
// [num] : object/element of array by number
// [*]   : all objects/elements for array.
type PathString string

func (s PathString) Build() (Path, error) {
	buf := []rune(s)
	length := len(buf)
	cursor := 0
	builder := &pathBuilder{}
	start := 0
	for cursor < length {
		c := buf[cursor]
		switch c {
		case '.':
			if start < cursor {
				builder.child(string(buf[start:cursor]))
			}
			c, err := builder.parsePathDot(buf, cursor)
			if err != nil {
				return nil, err
			}
			cursor = c
			start = cursor + 1
		case '[':
			if start < cursor {
				builder.child(string(buf[start:cursor]))
			}
			c, err := builder.parsePathIndex(buf, cursor)
			if err != nil {
				return nil, err
			}
			cursor = c
			start = cursor + 1
		}
		cursor++
	}
	if start < cursor {
		builder.child(string(buf[start:cursor]))
	}
	return builder.Build(), nil
}

func (b *pathBuilder) parsePathRecursive(buf []rune, cursor int) (int, error) {
	length := len(buf)
	cursor += 2 // skip .. characters
	start := cursor
	for ; cursor < length; cursor++ {
		c := buf[cursor]
		switch c {
		case '$', '*', ']':
			return 0, fmt.Errorf("specified '%c' after '..' character", c)
		case '.', '[':
			goto end
		}
	}
end:
	if start == cursor {
		return 0, fmt.Errorf("not found recursive selector")
	}
	b.recursive(string(buf[start:cursor]))
	return cursor, nil
}

func (b *pathBuilder) parsePathDot(buf []rune, cursor int) (int, error) {
	length := len(buf)
	if cursor+1 < length && buf[cursor+1] == '.' {
		c, err := b.parsePathRecursive(buf, cursor)
		if err != nil {
			return 0, err
		}
		return c, nil
	}
	cursor++ // skip . character
	start := cursor
	for ; cursor < length; cursor++ {
		c := buf[cursor]
		switch c {
		case '$', '*', ']':
			return 0, fmt.Errorf("specified '%c' after '.' character", c)
		case '.', '[':
			goto end
		}
	}
end:
	if start == cursor {
		return 0, fmt.Errorf("not found child selector")
	}
	b.child(string(buf[start:cursor]))
	return cursor, nil
}

func (b *pathBuilder) parsePathIndex(buf []rune, cursor int) (int, error) {
	length := len(buf)
	cursor++ // skip '[' character
	if length <= cursor {
		return 0, fmt.Errorf("unexpected end of JSON Path")
	}
	c := buf[cursor]
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*':
		start := cursor
		cursor++
		for ; cursor < length; cursor++ {
			c := buf[cursor]
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				continue
			}
			break
		}
		if buf[cursor] != ']' {
			return 0, fmt.Errorf("invalid character %s at %d", string(buf[cursor]), cursor)
		}
		numOrAll := string(buf[start:cursor])
		if numOrAll == "*" {
			b.indexAll()
			return cursor + 1, nil
		}
		num, err := strconv.ParseInt(numOrAll, 10, 64)
		if err != nil {
			return 0, err
		}
		b.index(int(num))
		return cursor + 1, nil
	}
	return 0, fmt.Errorf("invalid character %s at %d", string(c), cursor)
}

type pathBuilder struct {
	root Path
	node Path
}

func (b *pathBuilder) indexAll() {
	node := newIndexAllNode()
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
}

func (b *pathBuilder) recursive(selector string) {
	node := newRecursiveNode(selector)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
}

func (b *pathBuilder) child(name string) {
	node := newSelectorNode(name)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
}

func (b *pathBuilder) index(idx int) {
	node := newIndexNode(idx)
	if b.root == nil {
		b.root = node
		b.node = node
	} else {
		b.node = b.node.chain(node)
	}
}

func (b *pathBuilder) Build() Path {
	return b.root
}

type Path interface {
	fmt.Stringer
	chain(Path) Path
	Index(int) (Path, bool, error)
	Field(string) (Path, bool, error)
	Get(reflect.Value, reflect.Value) error
	target() bool
	allRead() bool
	single() bool
}

type basePathNode struct {
	child Path
}

func (n *basePathNode) allRead() bool {
	return true
}

func (n *basePathNode) chain(node Path) Path {
	n.child = node
	return node
}

func (n *basePathNode) target() bool {
	return n.child == nil
}

func (n *basePathNode) single() bool {
	return true
}

type selectorNode struct {
	*basePathNode
	selector string
}

func newSelectorNode(selector string) *selectorNode {
	return &selectorNode{
		basePathNode: &basePathNode{},
		selector:     strings.ToLower(selector),
	}
}

func (n *selectorNode) Index(idx int) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *selectorNode) Field(fieldName string) (Path, bool, error) {
	if n.selector == fieldName {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *selectorNode) Get(src, dst reflect.Value) error {
	switch src.Type().Kind() {
	case reflect.Map:
		iter := src.MapRange()
		for iter.Next() {
			key, ok := iter.Key().Interface().(string)
			if !ok {
				return fmt.Errorf("invalid map key type %T", src.Type().Key())
			}
			child, found, err := n.Field(strings.ToLower(key))
			if err != nil {
				return err
			}
			if found {
				if child != nil {
					return child.Get(iter.Value(), dst)
				}
				return assignValue(iter.Value(), dst)
			}
		}
	case reflect.Struct:
		typ := src.Type()
		for i := 0; i < typ.Len(); i++ {
			tag := runtime.StructTagFromField(typ.Field(i))
			child, found, err := n.Field(strings.ToLower(tag.Key))
			if err != nil {
				return err
			}
			if found {
				if child != nil {
					return child.Get(src.Field(i), dst)
				}
				return assignValue(src.Field(i), dst)
			}
		}
	case reflect.Ptr:
		return n.Get(src.Elem(), dst)
	case reflect.Interface:
		return n.Get(reflect.ValueOf(src.Interface()), dst)
	case reflect.Float64, reflect.String, reflect.Bool:
		return assignValue(src, dst)
	}
	return fmt.Errorf("failed to get %s value from %s", n.selector, src.Type())
}

func (n *selectorNode) String() string {
	s := fmt.Sprintf(".%s", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type indexNode struct {
	*basePathNode
	selector int
}

func newIndexNode(selector int) *indexNode {
	return &indexNode{
		basePathNode: &basePathNode{},
		selector:     selector,
	}
}

func (n *indexNode) Index(idx int) (Path, bool, error) {
	if n.selector == idx {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *indexNode) Field(fieldName string) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *indexNode) Get(src, dst reflect.Value) error {
	switch src.Type().Kind() {
	case reflect.Array, reflect.Slice:
		if src.Len() > n.selector {
			if n.child != nil {
				return n.child.Get(src.Index(n.selector), dst)
			}
			return assignValue(src.Index(n.selector), dst)
		}
	case reflect.Ptr:
		return n.Get(src.Elem(), dst)
	case reflect.Interface:
		return n.Get(reflect.ValueOf(src.Interface()), dst)
	}
	return fmt.Errorf("failed to get [%d] value from %s", n.selector, src.Type())
}

func (n *indexNode) String() string {
	s := fmt.Sprintf("[%d]", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type indexAllNode struct {
	*basePathNode
}

func newIndexAllNode() *indexAllNode {
	return &indexAllNode{
		basePathNode: &basePathNode{},
	}
}

func (n *indexAllNode) Index(idx int) (Path, bool, error) {
	return n.child, true, nil
}

func (n *indexAllNode) Field(fieldName string) (Path, bool, error) {
	return nil, false, &errors.PathError{}
}

func (n *indexAllNode) Get(src, dst reflect.Value) error {
	switch src.Type().Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < src.Len(); i++ {
			if n.child != nil {
				if err := n.child.Get(src.Index(i), dst); err != nil {
					return err
				}
			} else {
				if err := assignValue(src.Index(i), dst); err != nil {
					return err
				}
			}
		}
		return nil
	case reflect.Ptr:
		return n.Get(src.Elem(), dst)
	case reflect.Interface:
		return n.Get(reflect.ValueOf(src.Interface()), dst)
	}
	return fmt.Errorf("failed to get all value from %s", src.Type())
}

func (n *indexAllNode) String() string {
	s := "[*]"
	if n.child != nil {
		s += n.child.String()
	}
	return s
}

type recursiveNode struct {
	*basePathNode
	selector string
}

func newRecursiveNode(selector string) *recursiveNode {
	node := newSelectorNode(selector)
	return &recursiveNode{
		basePathNode: &basePathNode{
			child: node,
		},
		selector: selector,
	}
}

func (n *recursiveNode) Field(fieldName string) (Path, bool, error) {
	if n.selector == fieldName {
		return n.child, true, nil
	}
	return nil, false, nil
}

func (n *recursiveNode) Index(_ int) (Path, bool, error) {
	return n, true, nil
}

func (n *recursiveNode) Get(src, dst reflect.Value) error {
	if n.child == nil {
		return fmt.Errorf("failed to get by recursive path ..%s", n.selector)
	}
	switch src.Type().Kind() {
	case reflect.Map:
		iter := src.MapRange()
		for iter.Next() {
			key, ok := iter.Key().Interface().(string)
			if !ok {
				return fmt.Errorf("invalid map key type %T", src.Type().Key())
			}
			child, found, err := n.Field(strings.ToLower(key))
			if err != nil {
				return err
			}
			if found {
				_ = child.Get(iter.Value(), dst)
			} else {
				_ = n.Get(iter.Value(), dst)
			}
		}
		return nil
	case reflect.Struct:
		typ := src.Type()
		for i := 0; i < typ.Len(); i++ {
			tag := runtime.StructTagFromField(typ.Field(i))
			child, found, err := n.Field(strings.ToLower(tag.Key))
			if err != nil {
				return err
			}
			if found {
				_ = child.Get(src.Field(i), dst)
			} else {
				_ = n.Get(src.Field(i), dst)
			}
		}
		return nil
	case reflect.Array, reflect.Slice:
		for i := 0; i < src.Len(); i++ {
			_ = n.Get(src.Index(i), dst)
		}
		return nil
	case reflect.Ptr:
		return n.Get(src.Elem(), dst)
	case reflect.Interface:
		return n.Get(reflect.ValueOf(src.Interface()), dst)
	}
	return fmt.Errorf("failed to get %s value from %s", n.selector, src.Type())
}

func (n *recursiveNode) String() string {
	s := fmt.Sprintf("..%s", n.selector)
	if n.child != nil {
		s += n.child.String()
	}
	return s
}
