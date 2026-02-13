package codec

/*
#cgo pkg-config: libjxl libjxl_threads
#include <jxl/decode.h>
#include <jxl/encode.h>
#include <jxl/thread_parallel_runner.h>
#include <stdlib.h>
#include <string.h>

// jxl_decode decodes a JXL image from raw data into RGBA pixels.
// Returns 0 on success, negative on error.
// On success, *out_pixels is allocated with malloc and must be freed by caller.
// *out_width and *out_height are set to the image dimensions.
static int jxl_decode(const uint8_t* data, size_t data_len,
                      uint8_t** out_pixels, uint32_t* out_width, uint32_t* out_height) {
    JxlDecoder* dec = JxlDecoderCreate(NULL);
    if (!dec) return -1;

    void* runner = JxlThreadParallelRunnerCreate(NULL,
        JxlThreadParallelRunnerDefaultNumWorkerThreads());
    if (!runner) {
        JxlDecoderDestroy(dec);
        return -2;
    }

    if (JxlDecoderSetParallelRunner(dec, JxlThreadParallelRunner, runner) != JXL_DEC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlDecoderDestroy(dec);
        return -3;
    }

    if (JxlDecoderSubscribeEvents(dec, JXL_DEC_BASIC_INFO | JXL_DEC_FULL_IMAGE) != JXL_DEC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlDecoderDestroy(dec);
        return -4;
    }

    if (JxlDecoderSetInput(dec, data, data_len) != JXL_DEC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlDecoderDestroy(dec);
        return -5;
    }
    JxlDecoderCloseInput(dec);

    JxlBasicInfo info;
    JxlPixelFormat format = {4, JXL_TYPE_UINT8, JXL_NATIVE_ENDIAN, 0};
    uint8_t* pixels = NULL;

    for (;;) {
        JxlDecoderStatus status = JxlDecoderProcessInput(dec);

        if (status == JXL_DEC_BASIC_INFO) {
            if (JxlDecoderGetBasicInfo(dec, &info) != JXL_DEC_SUCCESS) {
                if (pixels) free(pixels);
                JxlThreadParallelRunnerDestroy(runner);
                JxlDecoderDestroy(dec);
                return -6;
            }
            *out_width = info.xsize;
            *out_height = info.ysize;
            continue;
        }

        if (status == JXL_DEC_NEED_IMAGE_OUT_BUFFER) {
            size_t buffer_size;
            if (JxlDecoderImageOutBufferSize(dec, &format, &buffer_size) != JXL_DEC_SUCCESS) {
                JxlThreadParallelRunnerDestroy(runner);
                JxlDecoderDestroy(dec);
                return -7;
            }
            pixels = (uint8_t*)malloc(buffer_size);
            if (!pixels) {
                JxlThreadParallelRunnerDestroy(runner);
                JxlDecoderDestroy(dec);
                return -8;
            }
            if (JxlDecoderSetImageOutBuffer(dec, &format, pixels, buffer_size) != JXL_DEC_SUCCESS) {
                free(pixels);
                JxlThreadParallelRunnerDestroy(runner);
                JxlDecoderDestroy(dec);
                return -9;
            }
            continue;
        }

        if (status == JXL_DEC_FULL_IMAGE) {
            break;
        }

        if (status == JXL_DEC_SUCCESS) {
            break;
        }

        // Error or unexpected status
        if (pixels) free(pixels);
        JxlThreadParallelRunnerDestroy(runner);
        JxlDecoderDestroy(dec);
        return -10;
    }

    *out_pixels = pixels;
    JxlThreadParallelRunnerDestroy(runner);
    JxlDecoderDestroy(dec);
    return 0;
}

// jxl_encode encodes RGBA pixels into JXL format.
// Returns 0 on success, negative on error.
// On success, *out_data is allocated with malloc and must be freed by caller.
// *out_size is set to the output data length.
static int jxl_encode(const uint8_t* pixels, uint32_t width, uint32_t height,
                      float distance, int lossless, int effort,
                      uint8_t** out_data, size_t* out_size) {
    JxlEncoder* enc = JxlEncoderCreate(NULL);
    if (!enc) return -1;

    void* runner = JxlThreadParallelRunnerCreate(NULL,
        JxlThreadParallelRunnerDefaultNumWorkerThreads());
    if (!runner) {
        JxlEncoderDestroy(enc);
        return -2;
    }

    if (JxlEncoderSetParallelRunner(enc, JxlThreadParallelRunner, runner) != JXL_ENC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -3;
    }

    JxlBasicInfo info;
    JxlEncoderInitBasicInfo(&info);
    info.xsize = width;
    info.ysize = height;
    info.bits_per_sample = 8;
    info.exponent_bits_per_sample = 0;
    info.num_color_channels = 3;
    info.num_extra_channels = 1;
    info.alpha_bits = 8;
    info.alpha_exponent_bits = 0;
    info.uses_original_profile = lossless ? JXL_TRUE : JXL_FALSE;

    if (JxlEncoderSetBasicInfo(enc, &info) != JXL_ENC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -4;
    }

    JxlEncoderFrameSettings* frame_settings = JxlEncoderFrameSettingsCreate(enc, NULL);
    if (!frame_settings) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -5;
    }

    if (lossless) {
        JxlEncoderSetFrameDistance(frame_settings, 0.0f);
        JxlEncoderSetFrameLossless(frame_settings, JXL_TRUE);
    } else {
        JxlEncoderSetFrameDistance(frame_settings, distance);
    }

    if (effort >= 1 && effort <= 9) {
        JxlEncoderFrameSettingsSetOption(frame_settings, JXL_ENC_FRAME_SETTING_EFFORT, effort);
    }

    JxlPixelFormat pixel_format = {4, JXL_TYPE_UINT8, JXL_NATIVE_ENDIAN, 0};
    size_t pixel_size = (size_t)width * (size_t)height * 4;

    if (JxlEncoderAddImageFrame(frame_settings, &pixel_format, pixels, pixel_size) != JXL_ENC_SUCCESS) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -6;
    }

    JxlEncoderCloseInput(enc);

    // Collect output
    size_t buf_cap = 1024 * 1024; // Start with 1MB
    uint8_t* buf = (uint8_t*)malloc(buf_cap);
    if (!buf) {
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -7;
    }
    size_t total = 0;

    for (;;) {
        size_t avail = buf_cap - total;
        uint8_t* next_out = buf + total;
        JxlEncoderStatus status = JxlEncoderProcessOutput(enc, &next_out, &avail);

        total = (size_t)(next_out - buf);

        if (status == JXL_ENC_SUCCESS) {
            break;
        }
        if (status == JXL_ENC_NEED_MORE_OUTPUT) {
            size_t new_cap = buf_cap * 2;
            uint8_t* new_buf = (uint8_t*)realloc(buf, new_cap);
            if (!new_buf) {
                free(buf);
                JxlThreadParallelRunnerDestroy(runner);
                JxlEncoderDestroy(enc);
                return -8;
            }
            buf = new_buf;
            buf_cap = new_cap;
            continue;
        }
        // Error
        free(buf);
        JxlThreadParallelRunnerDestroy(runner);
        JxlEncoderDestroy(enc);
        return -9;
    }

    *out_data = buf;
    *out_size = total;
    JxlThreadParallelRunnerDestroy(runner);
    JxlEncoderDestroy(enc);
    return 0;
}
*/
import "C"

import (
	"fmt"
	"image"
	"io"
	"unsafe"
)

type jxlDecoder struct{}
type jxlEncoder struct{}

func (d *jxlDecoder) Decode(r io.ReadSeeker) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read jxl data: %w", err)
	}

	var outPixels *C.uint8_t
	var width, height C.uint32_t

	ret := C.jxl_decode(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		&outPixels,
		&width,
		&height,
	)
	if ret != 0 {
		return nil, fmt.Errorf("jxl decode failed (code %d)", int(ret))
	}
	defer C.free(unsafe.Pointer(outPixels))

	w := int(width)
	h := int(height)
	pixelData := C.GoBytes(unsafe.Pointer(outPixels), C.int(w*h*4))

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pixelData)
	return img, nil
}

func (d *jxlDecoder) Format() Format { return JXL }

func (e *jxlEncoder) Encode(w io.Writer, img image.Image, quality int) error {
	return e.EncodeWithOptions(w, img, EncodeOptions{Quality: quality})
}

func (e *jxlEncoder) EncodeWithOptions(w io.Writer, img image.Image, opts EncodeOptions) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert image to RGBA pixels
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			off := (y*width + x) * 4
			pixels[off+0] = uint8(r >> 8)
			pixels[off+1] = uint8(g >> 8)
			pixels[off+2] = uint8(b >> 8)
			pixels[off+3] = uint8(a >> 8)
		}
	}

	// Map quality (1-100) to JXL distance (0.0-15.0)
	// quality 100 -> distance 0.0 (lossless-like)
	// quality 90  -> distance 1.0 (visually lossless)
	// quality 1   -> distance 15.0
	distance := qualityToJXLDistance(opts.Quality)
	lossless := 0
	if opts.Lossless {
		lossless = 1
		distance = 0.0
	}

	var outData *C.uint8_t
	var outSize C.size_t

	ret := C.jxl_encode(
		(*C.uint8_t)(unsafe.Pointer(&pixels[0])),
		C.uint32_t(width),
		C.uint32_t(height),
		C.float(distance),
		C.int(lossless),
		C.int(7), // default effort: squirrel
		&outData,
		&outSize,
	)
	if ret != 0 {
		return fmt.Errorf("jxl encode failed (code %d)", int(ret))
	}
	defer C.free(unsafe.Pointer(outData))

	encoded := C.GoBytes(unsafe.Pointer(outData), C.int(outSize))
	_, err := w.Write(encoded)
	return err
}

func (e *jxlEncoder) Format() Format { return JXL }

// qualityToJXLDistance converts a 1-100 quality scale to JXL distance (0.0-15.0).
// distance 0.0 = mathematically lossless
// distance 1.0 = visually lossless
// distance 15.0 = maximum lossy
func qualityToJXLDistance(quality int) float32 {
	if quality <= 0 {
		quality = 1
	}
	if quality >= 100 {
		return 0.0
	}
	// Use the same mapping as cjxl:
	// quality 100 -> distance 0.0
	// quality 90  -> distance 1.0
	// quality 30  -> distance ~6.4
	// quality 1   -> distance 15.0
	if quality >= 30 {
		return 0.1 + float32(100-quality)*0.09
	}
	return 6.4 + float32(30-quality)*0.297
}

// Ensure jxlEncoder implements AdvancedEncoder.
var _ AdvancedEncoder = (*jxlEncoder)(nil)

// Ensure jxlDecoder satisfies Decoder interface at compile time.
var _ Decoder = (*jxlDecoder)(nil)

func registerJXL(r *Registry) {
	r.RegisterDecoder(&jxlDecoder{})
	r.RegisterEncoder(&jxlEncoder{})
}

