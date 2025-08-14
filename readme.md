# saferimg - Golang image decoding with safety checks
`saferimg` decodes images using disintegration/imaging, but performs optional safety checks first.

## Why is this useful?
Both Golang's stdlib [image.Decode](https://pkg.go.dev/image#Decode) and disintegration/imaging's
[imaging.Decode](https://pkg.go.dev/github.com/disintegration/imaging#Decode) will decode any image
into memory, regardless of its dimensions.

Decoding untrusted input, e.g. a user uploaded image, can lead to unbounded memory usage, which leaves
services vulnerable to denial of service attacks.

It is not sufficient to limit the size of the uploaded data, as images can compress incredibly well.
See the `pngbomb.png` file for an example of this: it's a 32000w x 32000h png, which will result in a
multi-gigabyte decoded image.

## How it works
The image configuration is read before decoding, which provides width, height and color model. These
are then compared against the `MaxWidth`, `MaxHeight`, `MaxMemory` limits set in
[Opts](http://pkg.go.dev/github.com/iamcalledrob/saferimg#Opts).

Memory usage is derived from the dimensions and the BPP of the color model.


## Usage
[Godoc](http://pkg.go.dev/github.com/iamcalledrob/saferimg)


#### Decoding using defaults
Applies a 32mb memory limit as a sane default.
```go
img, err := saferimg.Decode(r)
```


#### Decoding with options
Allows customisation of `MaxWidth`, `MaxHeight` and `MaxMemory`.
```go
opts := saferimg.Opts{
    // Don't decode images larger than 4096px in either dimension
    MaxWidth: 4096,
    MaxHeight: 4096,
    
    // Don't decode images that will decode to more than 16mb
    MaxMemory: 16*1024*1024,
}

decoder := saferimg.NewDecoder(opts)

// Decode the image (or fail)
img, err := decoder.Decode(r)
```

#### Advanced
[PeekConfig](http://pkg.go.dev/github.com/iamcalledrob/saferimg#PeekConfig) and
[EstimatedMemory](http://pkg.go.dev/github.com/iamcalledrob/saferimg#PeekConfig) can be used for more advanced scenarios,
such as using a [weighted semaphore](https://pkg.go.dev/golang.org/x/sync/semaphore) to limit overall memory
usage if decoding concurrently, e.g. in an http handler:

```go
// Allow 512mb of total ram for decoding images
sem := semaphore.NewWeighted(512*1024*1024)

func processImage(r io.Reader) error {
    // Find out how much memory is required to decode
    config, _, r, err := saferimg.PeekConfig(r)
    if err != nil {
        return err
    }
    requiredBytes := saferimg.EstimatedMemory(config)
	
    // Only process images that require less than 128MB ram. 
    if requiredBytes > 128*1024*1024 {
        return fmt.Errorf("image is too large to process")	
    }

    // "Acquire" these bytes, or wait until it's possible to do so.
    err = sem.Acquire(ctx, requiredBytes)
    if err != nil {
        return err
    }
    // When done, release these bytes for others to use
    defer func () { sem.Release(requiredBytes) }

	// Decode 
    image, err := imaging.Decode(r)
    if err != nil {
        return err
    }

    // Do something with the image! 
}
```

## Reference
Each megapixel (million pixels) requires approximately 4MB of memory when decoded, assuming
4 bits per pixel. So an iPhone 16, with a 48MP camera, generates images that require 192MB
of memory to decode, excluding any temporary overheads during decoding from the decoder itself.