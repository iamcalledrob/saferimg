package saferimg

import (
	"bytes"
	_ "embed"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"testing"
)

//go:embed asset/pngbomb.png
var pngbomb []byte

//go:embed asset/good.jpg
var good []byte
var goodBounds = image.Rect(0, 0, 958, 720)

func TestShouldDecode(t *testing.T) {
	opts := Opts{
		MaxWidth:  goodBounds.Dx(),
		MaxHeight: goodBounds.Dy(),
		MaxMemory: 4 * 1024 * 1024,
	}
	r := bytes.NewReader(good)
	config := requirePeekConfig(t, r)

	err := ShouldDecode(opts, config)
	if err != nil {
		t.Fatalf("ShouldDecode unexpectedly returned error: %v", err)
	}
}

func TestShouldDecode_ExceedsMaxDimension(t *testing.T) {
	opts := Opts{
		MaxWidth:  goodBounds.Dx() - 1,
		MaxHeight: goodBounds.Dy(),
		MaxMemory: 4 * 1024 * 1024,
	}
	r := bytes.NewReader(good)
	config := requirePeekConfig(t, r)

	err := ShouldDecode(opts, config)
	if !errors.Is(err, ErrExceedsMaxDimension) {
		t.Fatalf("expected ErrExceedsMaxDimension, got %v", err)
	}
}

func TestShouldDecode_ExceedsMaxMemory(t *testing.T) {
	opts := Opts{
		MaxMemory: 4 * 1024 * 1024,
	}
	r := bytes.NewReader(pngbomb)
	config := requirePeekConfig(t, r)

	err := ShouldDecode(opts, config)
	if !errors.Is(err, ErrExceedsMaxMemory) {
		t.Fatalf("expected ErrExceedsMaxMemory, got %v", err)
	}
}

func requirePeekConfig(t *testing.T, r io.Reader) image.Config {
	config, _, _, err := PeekConfig(r)
	if err != nil {
		t.Fatalf("peeking config: %v", err)
	}
	return config
}
