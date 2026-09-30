// Package codec converts audio payloads for transport and playback.
package codec

// Codec encodes and decodes one audio format.
type Codec interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}
