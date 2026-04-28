package config

import (
	"errors"
	"testing"
)

func TestAssetLoaderUsesSourceAndDecoder(t *testing.T) {
	want := &AppConfig{
		Economics: Economics{
			User: User{MaximumDebtAllowed: 500},
		},
	}

	loader := NewAssetLoader(
		SourceFunc(func() ([]byte, error) {
			return []byte("source-bytes"), nil
		}),
		DecoderFunc(func(data []byte) (*AppConfig, error) {
			if got := string(data); got != "source-bytes" {
				t.Fatalf("decoder received %q, want source-bytes", got)
			}
			return want, nil
		}),
	)

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg != want {
		t.Fatalf("Load returned unexpected config pointer")
	}
}

func TestAssetLoaderPropagatesSourceError(t *testing.T) {
	wantErr := errors.New("source failed")

	loader := NewAssetLoader(
		SourceFunc(func() ([]byte, error) {
			return nil, wantErr
		}),
		DecoderFunc(func([]byte) (*AppConfig, error) {
			t.Fatal("decoder should not be called when source fails")
			return nil, nil
		}),
	)

	_, err := loader.Load()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Load error = %v, want %v", err, wantErr)
	}
}

func TestAssetLoaderPropagatesDecoderError(t *testing.T) {
	wantErr := errors.New("decode failed")

	loader := NewAssetLoader(
		StaticSource("source-bytes"),
		DecoderFunc(func(data []byte) (*AppConfig, error) {
			if got := string(data); got != "source-bytes" {
				t.Fatalf("decoder received %q, want source-bytes", got)
			}
			return nil, wantErr
		}),
	)

	_, err := loader.Load()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Load error = %v, want %v", err, wantErr)
	}
}

func TestAssetLoaderRejectsNilSourceOrDecoder(t *testing.T) {
	_, err := NewAssetLoader(nil, DecoderFunc(func([]byte) (*AppConfig, error) {
		return &AppConfig{}, nil
	})).Load()
	if !errors.Is(err, ErrNilSource) {
		t.Fatalf("Load error = %v, want %v", err, ErrNilSource)
	}

	_, err = NewAssetLoader(StaticSource("source-bytes"), nil).Load()
	if !errors.Is(err, ErrNilDecoder) {
		t.Fatalf("Load error = %v, want %v", err, ErrNilDecoder)
	}
}
