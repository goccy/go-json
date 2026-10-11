package benchmark

import "github.com/francoispqt/gojay"

var SmallFixture = []byte(`{"st": 1,"sid": 486,"tt": "active","gr": 0,"uuid": "de305d54-75b4-431b-adb2-eb6b9e546014","ip": "127.0.0.1","ua": "user_agent","tz": -6,"v": 1}`)

// ffjson:skip
type SmallPayload struct {
	St   int
	Sid  int
	Tt   string
	Gr   int
	Uuid string
	Ip   string
	Ua   string
	Tz   int
	V    int
}

type SmallPayloadFFJson struct {
	St   int
	Sid  int
	Tt   string
	Gr   int
	Uuid string
	Ip   string
	Ua   string
	Tz   int
	V    int
}

//easyjson:json
type SmallPayloadEasyJson struct {
	St   int
	Sid  int
	Tt   string
	Gr   int
	Uuid string
	Ip   string
	Ua   string
	Tz   int
	V    int
}

func (t *SmallPayload) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddIntKey("st", t.St)
	enc.AddIntKey("sid", t.Sid)
	enc.AddStringKey("tt", t.Tt)
	enc.AddIntKey("gr", t.Gr)
	enc.AddStringKey("uuid", t.Uuid)
	enc.AddStringKey("ip", t.Ip)
	enc.AddStringKey("ua", t.Ua)
	enc.AddIntKey("tz", t.Tz)
	enc.AddIntKey("v", t.V)
}

func (t *SmallPayload) IsNil() bool {
	return t == nil
}

func (t *SmallPayload) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "st":
		return dec.AddInt(&t.St)
	case "sid":
		return dec.AddInt(&t.Sid)
	case "gr":
		return dec.AddInt(&t.Gr)
	case "tz":
		return dec.AddInt(&t.Tz)
	case "v":
		return dec.AddInt(&t.V)
	case "tt":
		return dec.AddString(&t.Tt)
	case "uuid":
		return dec.AddString(&t.Uuid)
	case "ip":
		return dec.AddString(&t.Ip)
	case "ua":
		return dec.AddString(&t.Ua)
	}
	return nil
}

func (t *SmallPayload) NKeys() int {
	return 9
}

func NewSmallPayload() *SmallPayload {
	return &SmallPayload{
		St:   1,
		Sid:  2,
		Tt:   "TestString",
		Gr:   4,
		Uuid: "8f9a65eb-4807-4d57-b6e0-bda5d62f1429",
		Ip:   "127.0.0.1",
		Ua:   "Mozilla",
		Tz:   8,
		V:    6,
	}
}

func NewSmallPayloadEasyJson() *SmallPayloadEasyJson {
	return &SmallPayloadEasyJson{
		St:   1,
		Sid:  2,
		Tt:   "TestString",
		Gr:   4,
		Uuid: "8f9a65eb-4807-4d57-b6e0-bda5d62f1429",
		Ip:   "127.0.0.1",
		Ua:   "Mozilla",
		Tz:   8,
		V:    6,
	}
}

func NewSmallPayloadFFJson() *SmallPayloadFFJson {
	return &SmallPayloadFFJson{
		St:   1,
		Sid:  2,
		Tt:   "TestString",
		Gr:   4,
		Uuid: "8f9a65eb-4807-4d57-b6e0-bda5d62f1429",
		Ip:   "127.0.0.1",
		Ua:   "Mozilla",
		Tz:   8,
		V:    6,
	}
}
