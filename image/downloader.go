// Copyright (c) 2017 - Sarjono Mukti Aji <me@simukti.net>
// Unless otherwise noted, this source code license is MIT-License

package image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

// DownloadJob Master image download definition
type DownloadJob struct {
	SourceURL      string
	SourcePath     string
	TargetDir      string
	TargetFilename string
}

// Process Do master image download and remove temporary file on process error
func (job *DownloadJob) Process() error {
	client := fasthttp.Client{
		Name:            "ImageServerDowloader",
		MaxConnsPerHost: 8,
	}

	filename := filepath.Join(job.TargetDir, job.TargetFilename)
	output, err := os.Create(filename)
	if err != nil {
		os.Remove(filename)
		return err
	}
	defer output.Close()

	// Do I need parallel download ??
	status, body, err := client.Get(nil, job.SourceURL)
	if err != nil {
		os.Remove(filename)
		return err
	}

	if status != fasthttp.StatusOK {
		os.Remove(filename)
		return fmt.Errorf(fmt.Sprintf("%s %s", job.SourcePath, fasthttp.StatusMessage(status)))
	}

	if _, err := output.Write(body); err != nil {
		os.Remove(filename)
		return err
	}

	return nil
}
