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
)

func (this *RestServer) LoadWebUI() {
	fmt.Println("Loading UI...")
	
	// Clear previous web UI file mappings and handlers
	webUIFileMapMutex.Lock()
	// Store old handlers to remove them
	oldPaths := make([]string, 0, len(webUIFileMap))
	for path := range webUIFileMap {
		oldPaths = append(oldPaths, path)
	}
	webUIFileMap = make(map[string]string)
	webUIFileMapMutex.Unlock()
	
	// Determine the web directory path
	webDir := this.getWebDirectory()
	
	// Scan and register all web files
	this.loadWebDir("/", webDir)
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

func (this *RestServer) loadWebDir(path string, webDir string) {
	dirName := concat(webDir, path)
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Println("Error loading web UI:", err)
		return
	}
	
	for _, file := range files {
		webPath := concat(path, file.Name())
		if file.IsDir() {
			this.loadWebDir(concat(webPath, "/"), webDir)
		} else {
			fullFilePath := filepath.Join(webDir, path, file.Name())
			if file.Name() == "index.html" {
				indexPath := path
				if indexPath != "/" && !strings.HasSuffix(indexPath, "/") {
					indexPath += "/"
				}
				fmt.Println("Loaded index.html at path:", indexPath)
				// Store mapping and register handler
				webUIFileMapMutex.Lock()
				webUIFileMap[indexPath] = fullFilePath
				webUIFileMapMutex.Unlock()
				
				http.HandleFunc(indexPath, func(w http.ResponseWriter, r *http.Request) {
					// Add cache-busting headers
					w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
					http.ServeFile(w, r, fullFilePath)
				})
			} else {
				fmt.Println("Loaded file:", webPath)
				// Store mapping and register handler
				webUIFileMapMutex.Lock()
				webUIFileMap[webPath] = fullFilePath
				webUIFileMapMutex.Unlock()
				
				http.HandleFunc(webPath, func(w http.ResponseWriter, r *http.Request) {
					// Add cache-busting headers
					w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					w.Header().Set("Pragma", "no-cache")
					w.Header().Set("Expires", "0")
					http.ServeFile(w, r, fullFilePath)
				})
			}
		}
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
