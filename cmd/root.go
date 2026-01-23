// SPDX-FileCopyrightText: 2026 Jacques Supcik <jacques.supcik@ehfr.ch>
//
// SPDX-License-Identifier: MIT

// Package cmd implements the root command for the zipserve application.

package cmd

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const (
	prefixFileName   = ".prefix"
	shutdownTimeout  = 5 * time.Second
	defaultPort      = 8080
	defaultDirectory = "."
)

var version = "dev"

// findRootDirectory searches for a .prefix file in the zip and returns its directory.
// If not found, returns the default directory.
func findRootDirectory(fileSystem fs.FS) string {
	log.Debugf("Searching for %s file to determine directory", prefixFileName)
	directory := defaultDirectory
	err := fs.WalkDir(fileSystem, defaultDirectory, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == prefixFileName {
			directory = path.Dir(p)
			log.Debugf("Found %s in %s", prefixFileName, directory)
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to walk directory: %v", err)
	}
	return directory
}

// readPrefixFromFile reads the prefix from a .prefix file in the zip.
// Returns empty string if the file doesn't exist or can't be read.
func readPrefixFromFile(fileSystem fs.FS, directory string) string {
	log.Debugf("Reading prefix from %s file", prefixFileName)
	f, err := fileSystem.Open(path.Join(directory, prefixFileName))
	if err != nil {
		return ""
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			log.Errorf("Failed to close file: %v", cerr)
		}
	}()

	s := bufio.NewScanner(f)
	if s.Scan() {
		return strings.TrimSpace(s.Text())
	}
	if err := s.Err(); err != nil {
		log.Warnf("Error reading %s file: %v", prefixFileName, err)
	}
	return ""
}

// normalizePrefix ensures the prefix starts and ends with "/".
func normalizePrefix(prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return prefix
}

// openBrowser opens the URL in the default browser if skipBrowser is false.
func openBrowser(url string, skipBrowser bool) {
	if skipBrowser {
		return
	}
	if err := browser.OpenURL(url); err != nil {
		log.Warnf("Failed to open browser: %v", err)
	}
}

// waitForShutdown blocks until interrupt signal or server error.
// Returns true if shutdown was triggered by interrupt, false on server error.
func waitForShutdown(serverErr <-chan error) bool {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		log.Info("Shutting down server...")
		return true
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
		return false
	}
}

// gracefulShutdown performs a graceful shutdown of the server with timeout.
func gracefulShutdown(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	} else {
		log.Info("Server stopped gracefully")
	}
}

func serve(prefix string, port int, zipfile string, directory string, skipBrowser bool) {
	log.Debug("Opening file ", zipfile)
	z, err := zip.OpenReader(zipfile)
	if err != nil {
		log.Fatalf("Failed to open zip file: %v", err)
	}

	defer func() {
		log.Debug("Closing zip file")
		if cerr := z.Close(); cerr != nil {
			log.Errorf("Failed to close zip file: %v", cerr)
		}
	}()

	// Determine directory
	if directory == "" {
		directory = findRootDirectory(z)
	}
	log.Debug("Using directory: ", directory)

	// Determine prefix
	if prefix == "" {
		prefix = readPrefixFromFile(z, directory)
	}
	prefix = normalizePrefix(prefix)
	log.Debug("Using prefix: ", prefix)

	// Validate directory
	if info, err := fs.Stat(z, directory); err != nil || !info.IsDir() {
		log.Fatalf("Directory %s not found or is not a directory in zip file", directory)
	}

	// Create sub filesystem
	fileSystem, err := fs.Sub(z, directory)
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	// Setup HTTP handler
	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.FS(fileSystem))))

	// Create and start server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Debugf("Starting HTTP server on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Print server info and open browser
	url := fmt.Sprintf("http://localhost:%d%s", port, prefix)
	openBrowser(url, skipBrowser)

	fmt.Printf("\n✓ Server running at %s\n", url)
	fmt.Println("  Press Ctrl+C to stop the server.")

	// Wait for shutdown signal
	waitForShutdown(serverErr)

	// Graceful shutdown
	gracefulShutdown(server)
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
	rootCmd.Flags().IntP("port", "p", defaultPort, "Port Number")
	rootCmd.Flags().
		StringP("prefix", "q", "", "Path prefix. If not set, the prefix is read from the .prefix file inside the zip file.")
	rootCmd.Flags().
		StringP("directory", "d", "", "Directory to serve in the zip file. If not set, the directory containing the .prefix file is used")
	rootCmd.Flags().BoolP("skip-browser", "n", false, "Do not open the browser automatically")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}
