// Copyright (c) 2017 - Sarjono Mukti Aji <me@simukti.net>
// Unless otherwise noted, this source code license is MIT-License

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/blake2b-simd"
	"github.com/simukti/imageserver/image"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

var (
	sourceServer      string
	masterDir         string
	resultDir         string
	hostPort          string
	timeout           int
	allowedPathPrefix string
	// map storage for splitted allowedPrefix's values
	prefixMap = map[string]bool{}
)

func main() {
	// https://github.com/valyala/fasthttp#performance-optimization-tips-for-multi-core-systems
	runtime.GOMAXPROCS(1)

	flag.StringVar(&sourceServer, "s", "", "Source server base URL. (Example: https://kadalkesit.storage.googleapis.com)")
	flag.StringVar(&masterDir, "m", "/tmp/imgsrv_master", "Directory for master image storage.")
	flag.StringVar(&resultDir, "r", "/tmp/imgsrv_result", "Directory for result image storage.")
	flag.StringVar(&hostPort, "h", "127.0.0.1:8080", "Host port to serve this app.")
	flag.StringVar(&allowedPathPrefix, "app", "*", "Allowed path prefix from source server, comma-separated for multiple values.")
	flag.IntVar(&timeout, "t", 30, "Process timeout per image processing request.")
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			msg := fmt.Sprintf("[ERROR] Flag param -%s is required", f.Name)
			fmt.Println(msg)
			fmt.Println("Usage:")
			flag.PrintDefaults()
			os.Exit(0)
		}
	})

	if err := ensureDir(masterDir); err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	if err := ensureDir(resultDir); err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	if allowedPathPrefix != "*" {
		// store allowed prefix in map for faster lookup
		rp := strings.Split(allowedPathPrefix, ",")
		for _, p := range rp {
			prefixMap[p] = true
		}
	}

	server := fasthttp.Server{
		Handler:            requestHandler,
		Name:               "ImageServer",
		Concurrency:        runtime.NumCPU() * 102400,
		MaxRequestBodySize: 1024 * 1, // 1 KiB
		ReadTimeout:        time.Second * time.Duration(timeout),
		WriteTimeout:       time.Second * time.Duration(timeout),
		ReduceMemoryUsage:  true,
	}

	listener, listenerError := reuseport.Listen("tcp4", hostPort)
	if listenerError != nil {
		log.Fatal(listenerError)
	}

	go func() {
		if serverError := server.Serve(listener); serverError != nil {
			log.Fatalf("Error in ListenAndServe: %s", serverError)
		}
	}()

	select {}
}

// requestHandler Main request handler
func requestHandler(req *fasthttp.RequestCtx) {
	if !req.IsGet() {
		req.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	reqPath := string(req.Path())

	if reqPath == "/" {
		req.SetStatusCode(fasthttp.StatusOK)
		req.SetBody([]byte("ImageServer"))
		return
	}

	if strings.ToLower(reqPath) == "/favicon.ico" {
		req.SetStatusCode(fasthttp.StatusNotFound)
		req.SetBody([]byte(fasthttp.StatusMessage(fasthttp.StatusNotFound)))
		return
	}

	if allowedPathPrefix != "*" {
		ps := strings.SplitN(reqPath, "/", 3)
		if !prefixMap[ps[1]] {
			req.SetStatusCode(fasthttp.StatusForbidden)
			req.SetBody([]byte(fasthttp.StatusMessage(fasthttp.StatusForbidden)))
			return
		}
	}

	reqQuery, err := url.ParseQuery(req.QueryArgs().String())
	if err != nil {
		log.Println(err)
	}

	params := image.ValidateParams(reqQuery)
	source, _ := url.Parse(sourceServer)
	source.Path = path.Join(source.Path, reqPath)

	imageJob := image.Job{
		MasterDir:  masterDir,
		ResultDir:  resultDir,
		SourceURL:  source.String(),
		SourcePath: source.Path,
		Params:     params,
	}

	reqByte, _ := json.Marshal(imageJob)
	imageJob.RequestHash = createHash(reqByte)

	resultFile := filepath.Join(imageJob.ResultDir, imageJob.RequestHash)
	if _, sErr := os.Stat(resultFile); sErr == nil {
		req.Response.Header.DelBytes([]byte("Accept-Encoding"))
		req.SendFile(resultFile)
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- imageJob.Process()
		close(done)
	}()

	select {
	case err := <-done:
		if err != nil {
			req.SetStatusCode(fasthttp.StatusBadRequest)
			req.SetBody([]byte(err.Error()))
			return
		}

		req.Response.Header.DelBytes([]byte("Accept-Encoding"))
		req.SendFile(resultFile)
		return
	case <-time.After(time.Second * time.Duration(timeout)):
		req.SetStatusCode(fasthttp.StatusGatewayTimeout)
		return
	}
}

// ensureDir Make sure directory exist and valid
func ensureDir(dir string) error {
	abs, absErr := filepath.Abs(dir)
	if absErr != nil {
		return fmt.Errorf("Unable to parse %s", abs)
	}

	s, sErr := os.Stat(abs)
	if sErr != nil {
		if os.IsNotExist(sErr) {
			os.Mkdir(abs, 0755)
		}
	} else {
		if !s.IsDir() {
			return fmt.Errorf("'%s' is not a directory", abs)
		}
	}

	return nil
}

// createHash Calculate hash from given request params
func createHash(b []byte) string {
	hash := blake2b.New256()
	hash.Write(b)

	return fmt.Sprintf("%x", hash.Sum(nil))
}
