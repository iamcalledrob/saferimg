package saferimg

import (
	"bytes"
	_ "embed"
	"errors"
	"image"
	"testing"
)

//go:embed asset/pngbomb.png
var pngbomb []byte

//go:embed asset/good.jpg
var good []byte
var goodBounds = image.Rect(0, 0, 958, 720)

func TestDecode(t *testing.T) {
	decoder := NewDecoder(Opts{
		MaxWidth:  goodBounds.Dx(),
		MaxHeight: goodBounds.Dy(),
		MaxMemory: 4 * 1024 * 1024,
	})

	r := bytes.NewReader(good)
	img, err := decoder.Decode(r)
	if err != nil {
		t.Fatalf("decoding good image: %s", err)
	}

	// Mainly to validate test assumptions
	if img.Bounds() != goodBounds {
		t.Fatalf("decoded image does not have expected bounds")
	}
}

func TestDecode_ExceedsMaxDimension(t *testing.T) {
	decoder := NewDecoder(Opts{
		MaxWidth:  goodBounds.Dx() - 1,
		MaxHeight: goodBounds.Dy(),
	})

	r := bytes.NewReader(good)
	_, err := decoder.Decode(r)
	if !errors.Is(err, ErrExceedsMaxDimension) {
		t.Fatalf("expected ErrExceedsMaxDimension, got %v", err)
	}
}

func TestDecode_ExceedsMaxMemory(t *testing.T) {
	decoder := NewDecoder(Opts{
		MaxMemory: 4 * 1024 * 1024,
	})

	r := bytes.NewReader(pngbomb)
	_, err := decoder.Decode(r)
	if !errors.Is(err, ErrExceedsMaxMemory) {
		t.Fatalf("expected ErrExceedsMaxMemory, got %v", err)
	}
}
