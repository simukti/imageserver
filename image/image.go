// Copyright (c) 2017 - Sarjono Mukti Aji <me@simukti.net>
// Unless otherwise noted, this source code license is MIT-License

package image

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/bimg"
)

// Params Supported request params for transformation
type Params struct {
	Width   int                 `json:"w,omitempty"`
	Height  int                 `json:"h,omitempty"`
	Quality int                 `json:"q,omitempty"`
	Blur    int                 `json:"blur,omitempty"`
	Flip    bool                `json:"flip,omitempty"`
	Flop    bool                `json:"flop,omitempty"`
	Colour  bimg.Interpretation `json:"c,omitempty"`
	Format  bimg.ImageType      `json:"fmt,omitempty"`
}

// Job Main job for image transformation
// Ccontain basic info for source file and requested params
type Job struct {
	RequestHash string `json:"-"`
	MasterDir   string `json:"-"`
	ResultDir   string `json:"-"`
	SourceURL   string `json:"-"`
	SourcePath  string `json:"source_path"`
	Params      Params `json:"image_params,omitempty"`
}

// Process Do image processing from downloading to vips process via bimg
func (job *Job) Process() error {
	download := DownloadJob{
		SourceURL:      job.SourceURL,
		SourcePath:     job.SourcePath,
		TargetDir:      job.MasterDir,
		TargetFilename: strings.Replace(job.SourcePath, "/", "_", -1),
	}

	sourceFilepath := filepath.Join(download.TargetDir, download.TargetFilename)
	if _, err := os.Stat(sourceFilepath); err != nil {
		if err := download.Process(); err != nil {
			return err
		}
	}

	params := job.Params
	options := bimg.Options{
		Enlarge:        false,
		Force:          false,
		NoProfile:      true,
		NoAutoRotate:   true,
		Crop:           true,
		Embed:          false,
		Interlace:      true,
		Compression:    9,
		Interpolator:   bimg.Bicubic,
		Interpretation: params.Colour,
		Type:           params.Format,
		Quality:        params.Quality,
		Flip:           params.Flip,
		Flop:           params.Flop,
	}

	buffer, bufferErr := bimg.Read(sourceFilepath)
	if bufferErr != nil {
		return bufferErr
	}

	image := bimg.NewImage(buffer)
	imageSize, imageError := image.Size()

	if imageError != nil {
		return imageError
	}

	if params.Colour == 0 {
		// see vips <include/image.h>
		options.Interpretation = bimg.InterpretationSRGB
	}

	if params.Width > 0 && params.Height == 0 {
		params.Height = params.Width
	} else if params.Height > 0 && params.Width == 0 {
		params.Width = params.Height
	}

	if params.Width > 0 && params.Width < imageSize.Width {
		options.Width = params.Width
	}

	if params.Height > 0 && options.Height < imageSize.Height {
		options.Height = params.Height
	}

	if params.Blur > 0 {
		options.GaussianBlur = bimg.GaussianBlur{
			Sigma: float64(params.Blur),
		}
	}

	resultFilepath := filepath.Join(job.ResultDir, job.RequestHash)
	result, err := image.Process(options)
	if err != nil {
		return err
	}

	if err := bimg.Write(resultFilepath, result); err != nil {
		return err
	}

	return nil
}
