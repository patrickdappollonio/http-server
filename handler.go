package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	tmpl     *template.Template
	tmplName = "list.tmpl"

	tmplFuncs = template.FuncMap{
		"humansize": func(s int64) string {
			return humansize(s)
		},
	}
)

type foldersFirst []os.FileInfo

func (f foldersFirst) Len() int      { return len(f) }
func (f foldersFirst) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f foldersFirst) Less(i, j int) bool {
	if f[i].IsDir() && f[j].IsDir() {
		return f[i].Name() < f[j].Name()
	}

	if !f[i].IsDir() && f[j].IsDir() {
		return false
	}

	if f[i].IsDir() && !f[j].IsDir() {
		return true
	}

	return f[i].Name() < f[j].Name()
}

func init() {
	tmpl = template.Must(template.New(tmplName).Funcs(tmplFuncs).ParseFiles(tmplName))
}

func handler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// If the method is GET, then we continue, we fail with "Method Not Allowed"
		// otherwise, since all request are for files.
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		// Check if the URL ends on "/index.html", if so, redirect to the folder, because
		// we can handle it later down the road
		if strings.HasSuffix(r.URL.Path, "/index.html") {
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "index.html"), http.StatusMovedPermanently)
			return
		}

		// Get the full path to the file or directory, since
		// we don't need the current working directory, we can
		// omit the error
		fullpath, _ := filepath.Abs(filepath.Join(path, r.URL.Path))

		// Find if there's a file or folder here
		info, err := os.Stat(fullpath)
		if err != nil {
			// If when trying to stat a file, the error is "not exists"
			// then we throw a 404
			if os.IsNotExist(err) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if it's a folder, if so, walk and present the contents on screen
		if info.IsDir() || strings.HasSuffix(fullpath, "/index.html") {
			walk(fullpath, w, r)
			return
		}

		serve(fullpath, w, r)
	}
}

func serve(path string, w http.ResponseWriter, r *http.Request) {
	// If there's no info coming, we get it
	info, err := os.Stat(path)
	if err != nil {
		// If when trying to stat a file, the error is "not exists"
		// then we throw a 404
		if os.IsNotExist(err) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Since http.ServeContent can only handle ReadSeekers then we
	// open the file for read.
	f, err := os.Open(path)

	// If we couldn't open, we just fail
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// We serve the content and close the file. ServeContent will handle
	// different headers like Range, Mime types and so on.
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	f.Close()
}

func walk(path string, w http.ResponseWriter, r *http.Request) {
	// Check if there's an index file, and if so, present it on screen
	indexPath := filepath.Join(path, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		serve(indexPath, w, r)
	}

	// If not, construct the UI we need with a list of files from this folder
	// by first opening the folder to get a Go object
	folder, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then listing all the files in it (by passing -1 meaning all)
	list, err := folder.Readdir(-1)
	folder.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Folders first, then alphabetically
	sort.Sort(foldersFirst(list))

	// If we reached this point, we're ready to print the template
	// so we create a bag, and we save the information there
	bag := map[string]interface{}{
		"Path":  r.URL.Path,
		"Files": list,
	}
	if err := tmpl.ExecuteTemplate(w, tmplName, bag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
