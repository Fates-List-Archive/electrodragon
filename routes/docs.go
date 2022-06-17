package routes

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wv2/types"

	"github.com/gorilla/mux"
)

// Returns the API document passed to it
func DocsGetDocument(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" {
		w.Write([]byte(invalidMethod))
		return
	}

	path := mux.Vars(r)["rest"]

	if strings.HasSuffix(path, ".md") {
		path = strings.Replace(path, ".md", "", 1)
	}

	if path == "" || path == "/docs" {
		path = "/index"
	}

	// Check if the file exists
	fmt.Println("api-docs/" + path + ".md")
	if _, err := os.Stat("api-docs/" + path + ".md"); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Read the file
	file, err := ioutil.ReadFile("api-docs/" + path + ".md")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalError))
		return
	}

	var data = types.Doc{}

	data.MD = string(file)

	// Look for javascript file in same place
	if _, err := os.Stat("api-docs/" + path + ".js"); err == nil {
		file, err := ioutil.ReadFile("api-docs/" + path + ".js")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(internalError))
			return
		}

		data.JS = string(file)
	}

	json.NewEncoder(w).Encode(data)
}

func DocsGenerateDoctree(w http.ResponseWriter, r *http.Request, opts types.RouteInfo) {
	if r.Method != "GET" {
		w.Write([]byte(invalidMethod))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var doctree []any

	// For every file, append its name into a slice, if its a directory, append its name and its children
	filepath.WalkDir("api-docs", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".js") {
			return nil
		}

		splitted := strings.Split(strings.Replace(path, "api-docs/", "", -1), "/")

		doctree = append(doctree, splitted)
		return nil
	})

	// Convert the slice into a json object
	json.NewEncoder(w).Encode(doctree)

}
