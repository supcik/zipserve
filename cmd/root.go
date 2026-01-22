// SPDX-FileCopyrightText: 2026 Jacques Supcik <jacques.supcik@ehfr.ch>
//
// SPDX-License-Identifier: MIT

package cmd

import (
	"archive/zip"
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

var version = "dev"

func serve(prefix string, port int, zipfile string, skipBrowser bool) {

	z, err := zip.OpenReader(zipfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	zfs := zipfs.New(z, "content")

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cerr := z.Close(); cerr != nil {
			log.Fatal(err)
		}
	}()

	if prefix == "" {
		// Try to read the prefix from the /.prefix file
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
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
		wg.Done()
	}()

	if !skipBrowser {
		err = browser.OpenURL(fmt.Sprintf("http://localhost:%d%s", port, prefix))
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("\n✓ Server running at http://localhost:%d%s\n", port, prefix)
	fmt.Println("  Press Ctrl+C to stop the server.")

	wg.Wait()
}

var rootCmd = &cobra.Command{
	Use:   "zipserve [flags] ZIPFILE",
	Short: "",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		port, _ := cmd.Flags().GetInt("port")
		prefix, _ := cmd.Flags().GetString("prefix")
		skipBrowser, _ := cmd.Flags().GetBool("skip-browser")

		serve(prefix, port, args[0], skipBrowser)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.Flags().IntP("port", "p", 8080, "Port Number")
	rootCmd.Flags().StringP("prefix", "q", "/", "Path prefix")
	rootCmd.Flags().BoolP("skip-browser", "s", false, "Do not open the browser automatically")
}
