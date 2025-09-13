package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (this *RestServer) LoadWebUI() {
	fmt.Println("Loading UI...")
	
	// Create a new ServeMux to clear previous handlers
	http.DefaultServeMux = http.NewServeMux()
	
	// Determine the web directory path
	webDir := this.getWebDirectory()
	
	fs := http.FileServer(http.Dir(webDir))
	this.loadWebDir("/", fs, webDir)
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

func (this *RestServer) loadWebDir(path string, fs http.Handler, webDir string) {
	dirName := concat(webDir, path)
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Println("Error loading web UI:", err)
		return
	}
	for _, file := range files {
		//filePath := app("./web", path, file.Name())
		webPath := concat(path, file.Name())
		if file.IsDir() {
			this.loadWebDir(concat(webPath, "/"), fs, webDir)
		} else {
			if file.Name() == "index.html" {
				fmt.Println("Loaded index.html")
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
				})
			} else {
				fmt.Println("Loaded file:", webPath)
				http.DefaultServeMux.HandleFunc(webPath, fs.ServeHTTP)
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
