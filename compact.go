package json

import (
	"bytes"
)

func compact(dst *bytes.Buffer, src []byte) error {
	length := len(src)
	for cursor := 0; cursor < length; cursor++ {
		c := src[cursor]
		switch c {
		case ' ', '\t', '\n', '\r':
			continue
		case '"':
			if err := dst.WriteByte(c); err != nil {
				return err
			}
			for {
				cursor++
				if err := dst.WriteByte(src[cursor]); err != nil {
					return err
				}
				switch src[cursor] {
				case '\\':
					cursor++
					if err := dst.WriteByte(src[cursor]); err != nil {
						return err
					}
				case '"':
					goto LOOP_END
				case nul:
					return errUnexpectedEndOfJSON("string", int64(length))
				}
			}
		default:
			if err := dst.WriteByte(c); err != nil {
				return err
			}
		}
	LOOP_END:
	}
	return nil
}
