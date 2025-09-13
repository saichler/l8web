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
	fs := http.FileServer(http.Dir("web"))
	this.loadWebDir("/", fs)
}

func (this *RestServer) loadWebDir(path string, fs http.Handler) {
	dirName := concat("./web", path)
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Println("Error loading web UI:", err)
		return
	}
	for _, file := range files {
		//filePath := app("./web", path, file.Name())
		webPath := concat(path, file.Name())
		if file.IsDir() {
			this.loadWebDir(concat(webPath, "/"), fs)
		} else {
			if file.Name() == "index.html" {
				fmt.Println("Loaded index.html")
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join("web", "index.html"))
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
