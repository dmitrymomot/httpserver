package httpserver

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// StaticHandler creates a new http.HandlerFunc that serves static files from the specified root directory.
// It does not allow directory listings and optionally supports caching of the served files.
//
// Parameters:
// - publicPath: The URL path prefix from which the static files will be served.
// - root: The http.FileSystem representing the root directory from which files will be served.
// - cacheTTL: The duration for which the client should cache the served files.
//
// Returns:
// An http.HandlerFunc that serves static files with optional caching.
func StaticHandler(publicPath string, root http.FileSystem, cacheTTL time.Duration) http.HandlerFunc {
	return serveStaticHandlerFunc(publicPath, root, cacheTTL)
}

// EmbeddedStaticHandler creates a new http.HandlerFunc that serves static files from an embedded file system.
// It uses the embed.FS type to serve files from the specified directory.
//
// Parameters:
// - publicPath: The URL path prefix from which the static files will be served.
// - fs: The embed.FS representing the embedded file system.
// - cacheTTL: The duration for which the client should cache the served files.
//
// Returns:
// An http.HandlerFunc that serves static files with optional caching.
func EmbeddedStaticHandler(publicPath string, fs embed.FS, cacheTTL time.Duration) http.HandlerFunc {
	return serveStaticHandlerFunc(publicPath, http.FS(fs), cacheTTL)
}

// serveFile serves a single file through HTTP with optional caching.
// It sets appropriate headers for caching based on the cacheTTL parameter.
// If cacheTTL is 0, caching is disabled.
//
// Parameters:
// - w: The http.ResponseWriter to write the response to.
// - r: The *http.Request representing the client's request.
// - file: The http.File representing the file to serve.
// - info: The os.FileInfo containing metadata about the file.
// - cacheTTL: The duration for which the file should be cached by the client.
func serveFile(w http.ResponseWriter, r *http.Request, file http.File, info os.FileInfo, cacheTTL time.Duration) {
	if cacheTTL == 0 {
		// No caching
		http.ServeContent(w, r, info.Name(), info.ModTime(), file)
		return
	}

	// Generate ETag using file info
	etag := fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size())
	lastModified := info.ModTime().UTC().Format(http.TimeFormat)

	// Set headers for caching
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", lastModified)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cacheTTL.Seconds())))
	w.Header().Set("Expires", time.Now().Add(cacheTTL).UTC().Format(http.TimeFormat))
	w.Header().Set("Pragma", "cache")

	// Check if file hasn't been modified since the last request
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Check if file has been modified since the last request based on Last-Modified header
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		if t, err := time.Parse(http.TimeFormat,
			ifModifiedSince); err == nil && info.ModTime().Before(t.Add(1*time.Second)) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Serve the file
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

// serveStaticHandlerFunc creates and returns a http.HandlerFunc that serves static files from a specified root directory.
// It does not allow directory listings and optionally supports caching of the served files.
//
// Parameters:
// - publicPath: The URL path prefix from which the static files will be served.
// - root: The http.FileSystem representing the root directory from which files will be served.
// - cacheTTL: The duration for which the client should cache the served files.
//
// Returns:
// An http.HandlerFunc that serves static files with optional caching.
func serveStaticHandlerFunc(publicPath string, root http.FileSystem, cacheTTL time.Duration) http.HandlerFunc {
	publicPath = strings.TrimRight(publicPath, "/")
	return func(w http.ResponseWriter, r *http.Request) {
		fsPath := strings.TrimPrefix(r.URL.Path, publicPath)
		file, err := root.Open(fsPath)
		if err != nil {
			// File not found
			http.NotFound(w, r)
			return
		}
		defer func(file http.File) {
			if err := file.Close(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(file)

		info, err := file.Stat()
		if err != nil {
			// Error getting file info
			http.NotFound(w, r)
			return
		}

		if info.IsDir() {
			// Path is a directory, return 404
			http.NotFound(w, r)
			return
		}

		// Serve file with caching
		serveFile(w, r, file, info, cacheTTL)
	}
}
