package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	webUIFileMap = make(map[string]string)
	webUIFileMapMutex sync.RWMutex
	webUIHandlerRegistered = false
)

func (this *RestServer) LoadWebUI() {
	fmt.Println("Loading UI...")
	
	// Clear previous web UI file mappings
	webUIFileMapMutex.Lock()
	webUIFileMap = make(map[string]string)
	webUIFileMapMutex.Unlock()
	
	// Determine the web directory path
	webDir := this.getWebDirectory()
	
	// Scan and register all web files
	this.scanWebDir("/", webDir)
	
	// Register a catch-all handler for web files (only once)
	if !webUIHandlerRegistered {
		http.HandleFunc("/", this.serveWebFile)
		webUIHandlerRegistered = true
	}
}

func (this *RestServer) getWebDirectory() string {
	// Try to find web directory in various locations
	possiblePaths := []string{
		"web",           // Current directory
		"./web",         // Relative to current
		"../web",        // Up one level
		"../../web",     // Up two levels
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// Default to "web" if not found
	return "web"
}

func (this *RestServer) scanWebDir(path string, webDir string) {
	dirName := concat(webDir, path)
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Println("Error scanning web UI:", err)
		return
	}
	
	webUIFileMapMutex.Lock()
	defer webUIFileMapMutex.Unlock()
	
	for _, file := range files {
		webPath := concat(path, file.Name())
		if file.IsDir() {
			this.scanWebDir(concat(webPath, "/"), webDir)
		} else {
			fullFilePath := filepath.Join(webDir, path, file.Name())
			if file.Name() == "index.html" {
				indexPath := path
				if indexPath != "/" && !strings.HasSuffix(indexPath, "/") {
					indexPath += "/"
				}
				webUIFileMap[indexPath] = fullFilePath
				fmt.Println("Mapped index.html:", indexPath, "->", fullFilePath)
			} else {
				webUIFileMap[webPath] = fullFilePath
				fmt.Println("Mapped file:", webPath, "->", fullFilePath)
			}
		}
	}
}

func (this *RestServer) serveWebFile(w http.ResponseWriter, r *http.Request) {
	webUIFileMapMutex.RLock()
	filePath, exists := webUIFileMap[r.URL.Path]
	webUIFileMapMutex.RUnlock()
	
	if !exists {
		// Try to find index.html for directory requests
		if strings.HasSuffix(r.URL.Path, "/") {
			webUIFileMapMutex.RLock()
			filePath, exists = webUIFileMap[r.URL.Path]
			webUIFileMapMutex.RUnlock()
		}
	}
	
	if exists {
		// Add cache-busting headers
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, filePath)
	} else {
		http.NotFound(w, r)
	}
}

func concat(strs ...string) string {
	buff := bytes.Buffer{}
	if strs != nil {
		for _, str := range strs {
			buff.WriteString(str)
		}
	}
	return buff.String()
}
