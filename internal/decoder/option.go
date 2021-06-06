package decoder

type OptionFlag int

const (
	FirstWinOption OptionFlag = 1 << iota
)

type Option struct {
	Flag OptionFlag
}
