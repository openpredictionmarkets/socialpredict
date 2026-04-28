package config

import (
	"bytes"
	"errors"

	"gopkg.in/yaml.v3"
)

var (
	ErrNilSource  = errors.New("config source is nil")
	ErrNilDecoder = errors.New("config decoder is nil")
)

// Source returns raw configuration bytes from the current source asset.
type Source interface {
	Bytes() ([]byte, error)
}

type SourceFunc func() ([]byte, error)

func (f SourceFunc) Bytes() ([]byte, error) {
	return f()
}

// StaticSource wraps immutable source bytes for loader construction.
type StaticSource []byte

func (s StaticSource) Bytes() ([]byte, error) {
	return bytes.Clone(s), nil
}

// Decoder transforms source bytes into the owned application-policy snapshot.
type Decoder interface {
	Decode(data []byte) (*AppConfig, error)
}

type DecoderFunc func(data []byte) (*AppConfig, error)

func (f DecoderFunc) Decode(data []byte) (*AppConfig, error) {
	return f(data)
}

// AssetLoader composes a source asset and decoder into the runtime loader seam.
type AssetLoader struct {
	source  Source
	decoder Decoder
}

// NewAssetLoader creates an explicit loader from a source asset and decoder.
func NewAssetLoader(source Source, decoder Decoder) Loader {
	return AssetLoader{
		source:  source,
		decoder: decoder,
	}
}

// NewYAMLLoader loads the owned application-policy config from YAML source bytes.
func NewYAMLLoader(source Source) Loader {
	return NewAssetLoader(source, DecoderFunc(decodeYAML))
}

func (l AssetLoader) Load() (*AppConfig, error) {
	if l.source == nil {
		return nil, ErrNilSource
	}
	if l.decoder == nil {
		return nil, ErrNilDecoder
	}

	data, err := l.source.Bytes()
	if err != nil {
		return nil, err
	}

	return l.decoder.Decode(data)
}

func decodeYAML(data []byte) (*AppConfig, error) {
	cfg := &AppConfig{}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
