// Copyright 2022 Jacques Supcik <jacques@supcik.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/pkg/browser"
	flag "github.com/spf13/pflag"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

func main() {
	flag.ErrHelp = errors.New("ZIPFILE is the archive containing the web site")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s ZIPFILE:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var prefix string

	port := flag.Int("port", 8080, "Port Number")
	flag.StringVar(&prefix, "prefix", "", "Path prefix")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(0)
	}

	zipfile := flag.Arg(0)
	z, err := zip.OpenReader(zipfile)
	zfs := zipfs.New(z, "content")

	if err != nil {
		log.Fatal(err)
	}
	defer z.Close()

	if prefix == "" {
		f, err := zfs.Open("/.prefix")
		if err == nil {
			s := bufio.NewScanner(f)
			if s.Scan() {
				prefix = s.Text()
			}
		}
	}

	// Make sure that the prefix starts with a "/" and also ends
	// with a "/"
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	httpfs := httpfs.New(zfs)
	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(httpfs)))
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
		wg.Done()
	}()

	err = browser.OpenURL(fmt.Sprintf("http://localhost:%d/%s", *port, prefix))
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
