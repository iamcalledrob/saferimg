// Package disintegration is a convenience wrapper for using saferimg with disintegration/imaging
package disintegration

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/iamcalledrob/saferimg"
	"image"
	"io"
)

// Decoder decodes images using disintegration/imaging, but performs safety checks first.
// Exists as a convenience - individual functions can be used instead for more complex use-cases.
type Decoder struct {
	opts saferimg.Opts
}

func NewDecoder(opts saferimg.Opts) *Decoder {
	return &Decoder{opts: opts}
}

// Decode is a convenience method to decode using DefaultOpts (32mb memory limit).
func Decode(r io.Reader, opts ...imaging.DecodeOption) (image.Image, error) {
	return NewDecoder(saferimg.DefaultOpts).Decode(r, opts...)
}

// Decode performs pre-flight safety checks before decoding an image from r.
// Decoding can be aborted if the image exceeds MaxWidth, MaxHeight or MaxMemory.
func (d *Decoder) Decode(r io.Reader, opts ...imaging.DecodeOption) (image.Image, error) {
	config, _, r2, err := saferimg.PeekConfig(r)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	err = saferimg.ShouldDecode(d.opts, config)
	if err != nil {
		return nil, err
	}

	return imaging.Decode(r2, opts...)
}
