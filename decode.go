package saferimg

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"io"
)

type Opts struct {
	// MaxWidth, in pixels, of the image
	MaxWidth int

	// MaxHeight, in pixels, of the image
	MaxHeight int

	// MaxMemory, in bytes, that this image can take up after being decoded
	MaxMemory int
}

// Decoder decodes images using disintegration/imaging, but performs safety checks first.
type Decoder struct {
	opts Opts
}

func NewDecoder(opts Opts) *Decoder {
	return &Decoder{opts: opts}
}

// Decode is a convenience method to decode using DefaultOpts (32mb memory limit).
func Decode(r io.Reader, opts ...imaging.DecodeOption) (image.Image, error) {
	return NewDecoder(DefaultOpts).Decode(r, opts...)
}

// Decode performs pre-flight safety checks before decoding an image from r.
// Decoding can be aborted if the image exceeds MaxWidth, MaxHeight or MaxMemory.
func (d *Decoder) Decode(r io.Reader, opts ...imaging.DecodeOption) (image.Image, error) {
	config, _, r2, err := PeekConfig(r)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	// Dimension checks
	if d.opts.MaxWidth != 0 && config.Width > d.opts.MaxWidth {
		return nil, fmt.Errorf("%w: width %d > %d", ErrExceedsMaxDimension, config.Width, d.opts.MaxWidth)
	}
	if d.opts.MaxHeight != 0 && config.Height > d.opts.MaxHeight {
		return nil, fmt.Errorf("%w: height %d > %d", ErrExceedsMaxDimension, config.Height, d.opts.MaxHeight)
	}

	// Memory check
	requiredBytes := EstimatedMemory(config)
	if d.opts.MaxMemory != 0 && requiredBytes > d.opts.MaxMemory {
		return nil, fmt.Errorf("%w: %db > %db", ErrExceedsMaxMemory, requiredBytes, d.opts.MaxMemory)
	}

	return imaging.Decode(r2, opts...)
}

// PeekConfig decodes and the image config from r.
// Returns the io.Reader r2 which re-plays the original stream as if it had not been consumed.
// Used to inspect the config before invoking a Decode method, which expects to read the config too.
func PeekConfig(r io.Reader) (config image.Config, format string, r2 io.Reader, err error) {
	b := new(bytes.Buffer)

	tee := io.TeeReader(r, b)
	config, format, err = image.DecodeConfig(tee)
	if err != nil {
		return
	}

	r2 = io.MultiReader(b, r)
	return
}

// EstimatedMemory returns the amount of memory required to store a decoded image with this image.Config.
// Does not account for any additional overheads that the decoder may have during decode.
func EstimatedMemory(config image.Config) int {
	return config.Width * config.Height * BytesPerPixel(config.ColorModel)
}

// BytesPerPixel returns the color.Model's bytes per pixel, or a conservative worst-case (16) if unknown.
func BytesPerPixel(cm color.Model) int {
	switch cm {
	case color.RGBAModel, color.NRGBAModel:
		return 4
	case color.RGBA64Model, color.NRGBA64Model:
		return 8
	case color.GrayModel:
		return 1
	case color.Gray16Model:
		return 2
	case color.CMYKModel:
		return 4
	case color.YCbCrModel:
		// YCbCr uses 3 components (8 bits each) but decoders may expand output
		return 3
	case color.NYCbCrAModel:
		return 4
	default:
		return 16 // safe worst-case
	}
}

var DefaultOpts = Opts{
	MaxMemory: 32 * 1024 * 1024,
}

var ErrExceedsMaxMemory = fmt.Errorf("required memory exceeds max")
var ErrExceedsMaxDimension = fmt.Errorf("dimension exceeds max")
