package encoder

type Option struct {
	HTMLEscape   bool
	Indent       bool
	UnorderedMap bool
	Debug        bool
	Colorize     bool
	ColorScheme  *ColorScheme
}

type EncodeFormat struct {
	Header string
	Footer string
}

type EncodeFormatScheme struct {
	Int         EncodeFormat
	Uint        EncodeFormat
	Float       EncodeFormat
	Bool        EncodeFormat
	String      EncodeFormat
	Binary      EncodeFormat
	ObjectStart EncodeFormat
	ObjectEnd   EncodeFormat
	ArrayStart  EncodeFormat
	ArrayEnd    EncodeFormat
	Colon       EncodeFormat
	Comma       EncodeFormat
}

type (
	ColorScheme = EncodeFormatScheme
	ColorFormat = EncodeFormat
)
