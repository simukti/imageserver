// Copyright (c) 2017 - Sarjono Mukti Aji <me@simukti.net>
// Unless otherwise noted, this source code license is MIT-License

package image

import (
	"github.com/h2non/bimg"
	"net/url"
	"strconv"
	"strings"
)

var (
	// image type
	jpeg = "jpeg"
	jpg  = "jpg"
	png  = "png"
	webp = "webp"

	// colour space
	srgb = "srgb"
	bw   = "bw"

	// flipflop
	flipHorizontal = "h" // flop
	flipVertical   = "v" // flip

	// default params
	defaultQuality = 75

	// max values
	maxDimension = 8192
	maxQuality   = 90
	maxBlurSigma = 50
)

// ValidateParams Validate requested query parameters
func ValidateParams(query url.Values) Params {
	params := Params{}

	// Validate output image width
	if width, err := strconv.Atoi(query.Get("w")); err == nil {
		if width > 0 && width <= maxDimension {
			params.Width = width
		}
	}

	// Validate output image height
	if height, err := strconv.Atoi(query.Get("h")); err == nil {
		if height > 0 && height <= maxDimension {
			params.Height = height
		}
	}

	// Validate output image quality
	if quality, err := strconv.Atoi(query.Get("q")); err == nil {
		if quality > 0 && quality <= maxQuality {
			params.Quality = quality
		} else {
			params.Quality = defaultQuality
		}
	}

	// Validate output blur level
	if blur, err := strconv.Atoi(query.Get("blur")); err == nil {
		if blur > 0 && blur <= maxBlurSigma {
			params.Blur = blur
		} else if blur > 0 && blur > maxBlurSigma {
			params.Blur = maxBlurSigma
		}
	}

	// Validate output format
	format := strings.ToLower(query.Get("fmt"))
	switch format {
	case png:
		params.Format = bimg.PNG
	case jpeg, jpg:
		params.Format = bimg.JPEG
	case webp:
		params.Format = bimg.WEBP
	}

	// Validate output colour space
	cSpace := strings.ToLower(query.Get("c"))
	switch cSpace {
	case srgb:
		params.Colour = bimg.InterpretationSRGB
	case bw:
		params.Colour = bimg.InterpretationBW
	default:
		params.Colour = bimg.InterpretationSRGB
	}

	// Validate output flip
	flip := strings.ToLower(query.Get("flip"))
	switch flip {
	case flipHorizontal:
		params.Flop = true
	case flipVertical:
		params.Flip = true
	}

	return params
}
