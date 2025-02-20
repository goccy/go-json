module benchmark

go 1.19

require (
	github.com/francoispqt/gojay v1.2.13
	github.com/bytedance/sonic v0.0.0-00010101000000-000000000000
	github.com/json-iterator/go v1.1.10
	github.com/mailru/easyjson v0.0.0-20190312143242-1de009706dbe
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/segmentio/encoding v0.2.4
	github.com/valyala/fastjson v1.6.3
	github.com/wI2L/jettison v0.7.1
)

require (
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
)

replace github.com/bytedance/sonic => ../
