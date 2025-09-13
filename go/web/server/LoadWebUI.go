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
	webUIHandlerRegistry = make(map[string]http.HandlerFunc)
	webUIHandlerRegistryMutex sync.RWMutex
)

func (this *RestServer) LoadWebUI() {
	fmt.Println("Loading UI...")
	
	// Clear and reload web UI file mappings (but keep handler registry intact)
	webUIFileMapMutex.Lock()
	webUIFileMap = make(map[string]string)
	webUIFileMapMutex.Unlock()
	
	// DO NOT clear handler registry - handlers remain registered in ServeMux
	
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
				// Store mapping
				webUIFileMapMutex.Lock()
				webUIFileMap[indexPath] = fullFilePath
				webUIFileMapMutex.Unlock()
				
				// Check if handler is already registered
				webUIHandlerRegistryMutex.RLock()
				_, exists := webUIHandlerRegistry[indexPath]
				webUIHandlerRegistryMutex.RUnlock()
				
				if !exists {
					handler := this.createDynamicHandler(indexPath)
					webUIHandlerRegistryMutex.Lock()
					webUIHandlerRegistry[indexPath] = handler
					webUIHandlerRegistryMutex.Unlock()
					http.HandleFunc(indexPath, handler)
				}
			} else {
				fmt.Println("Loaded file:", webPath)
				// Store mapping
				webUIFileMapMutex.Lock()
				webUIFileMap[webPath] = fullFilePath
				webUIFileMapMutex.Unlock()
				
				// Check if handler is already registered
				webUIHandlerRegistryMutex.RLock()
				_, exists := webUIHandlerRegistry[webPath]
				webUIHandlerRegistryMutex.RUnlock()
				
				if !exists {
					handler := this.createDynamicHandler(webPath)
					webUIHandlerRegistryMutex.Lock()
					webUIHandlerRegistry[webPath] = handler
					webUIHandlerRegistryMutex.Unlock()
					http.HandleFunc(webPath, handler)
				}
			}
		}
	}
}

func (this *RestServer) createDynamicHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Dynamically look up the current file path
		webUIFileMapMutex.RLock()
		filePath, exists := webUIFileMap[path]
		webUIFileMapMutex.RUnlock()
		
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
