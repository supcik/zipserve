// SPDX-FileCopyrightText: 2026 Jacques Supcik <jacques.supcik@ehfr.ch>
//
// SPDX-License-Identifier: MIT

// Package cmd implements the root command for the zipserve application.

package cmd

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var version = "dev"

func serve(prefix string, port int, zipfile string, directory string, skipBrowser bool) {

	log.Debug("Opening file ", zipfile)
	z, err := zip.OpenReader(zipfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		log.Debug("Closing zip file")
		if cerr := z.Close(); cerr != nil {
			log.Fatal(err)
		}
	}()

	// If directory is not set, search for a .prefix file
	if directory == "" {
		log.Debug("Searching for .prefix file to determine directory")
		directory = "."
		err := fs.WalkDir(z, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.Name() == ".prefix" {
				directory = path.Dir(p)
				log.Debug("Found .prefix in ", directory)
				return fs.SkipAll
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Debug("Using directory: ", directory)

	if prefix == "" {
		log.Debug("Reading prefix from .prefix file")
		// Try to read the prefix from the /.prefix file
		f, err := z.Open(path.Join(directory, ".prefix"))
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

	log.Debug("Using prefix: ", prefix)

	if info, err := fs.Stat(z, directory); err != nil || !info.IsDir() {
		fmt.Printf("Error: directory %s not found in zip file\n", directory)
		os.Exit(1)
	}

	fileSystem, err := fs.Sub(z, directory)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.FS(fileSystem))))
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()

	url := fmt.Sprintf("http://localhost:%d%s", port, prefix)
	if !skipBrowser {
		err = browser.OpenURL(url)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("\n✓ Server running at %s\n", url)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("  Press Ctrl+C to stop the server.")
	<-done
}

var rootCmd = &cobra.Command{
	Use:   "zipserve [flags] ZIPFILE",
	Short: "Serve contents of a ZIP file over HTTP",
	Long: `zipserve is a simple tool to serve the contents of a ZIP file over HTTP.
It allows you to quickly share files contained in a ZIP archive (such as a lecture web site) via a web browser.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		port, _ := cmd.Flags().GetInt("port")
		prefix, _ := cmd.Flags().GetString("prefix")
		directory, _ := cmd.Flags().GetString("directory")
		skipBrowser, _ := cmd.Flags().GetBool("skip-browser")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		serve(prefix, port, args[0], directory, skipBrowser)
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
	rootCmd.Flags().StringP("prefix", "q", "", "Path prefix. If not set, the prefix is read from the .prefix file inside the zip file.")
	rootCmd.Flags().StringP("directory", "d", "", "Directory to serve in the zip file. If not set, the directory containing the .prefix file is used")
	rootCmd.Flags().BoolP("skip-browser", "n", false, "Do not open the browser automatically")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
