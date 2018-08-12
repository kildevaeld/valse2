package httpcontext

import (
	"encoding/json"

	"github.com/kildevaeld/strong"
)

type JsonEncoding struct {
}

func (j *JsonEncoding) Decode(bs []byte, v interface{}) error {
	return json.Unmarshal(bs, v)
}

func (j *JsonEncoding) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

var (
	encoders map[string]Encoder
	decoders map[string]Decoder
)

func init() {
	encoders = make(map[string]Encoder)
	decoders = make(map[string]Decoder)

	jsonEncoder := &JsonEncoding{}
	decoders[strong.MIMEApplicationJSON] = jsonEncoder
	decoders[strong.MIMEApplicationJSONCharsetUTF8] = jsonEncoder
	encoders[strong.MIMEApplicationJSON] = jsonEncoder
	encoders[strong.MIMEApplicationJSONCharsetUTF8] = jsonEncoder
}

type Decoder interface {
	Decode(bs []byte, v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

func RegisterDecoder(contentType string, decoder Decoder) {
	decoders[contentType] = decoder
}

func RegisterEncoder(contentType string, encoder Encoder) {
	encoders[contentType] = encoder
}

func GetDecoder(constType string) Decoder {
	return decoders[constType]
}

func GetEncoder(constType string) Decoder {
	return decoders[constType]
}
