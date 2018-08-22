package httpcontext

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack"

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

type MsgPackEncoding struct {
}

func (m *MsgPackEncoding) Decode(bs []byte, v interface{}) error {
	return msgpack.Unmarshal(bs, v)
}

func (m *MsgPackEncoding) Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
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

	msgPackEncoder := &MsgPackEncoding{}
	encoders[strong.MIMEApplicationMsgpack] = msgPackEncoder
	decoders[strong.MIMEApplicationMsgpack] = msgPackEncoder

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
