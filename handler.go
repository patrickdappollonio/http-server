package main

import (
	"fmt"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	tmpl *template.Template

	tmplFuncs = template.FuncMap{
		"humansize": func(s int64) string {
			return humansize(s)
		},
		"mergepath": func(a ...string) string {
			return path.Clean(path.Join(a...))
		},
		"mergepathtrail": func(a ...string) string {
			m := path.Clean(path.Join(a...)) + "/"
			if m == "//" {
				return "/"
			}
			return m
		},
		"contenttype": func(path string, f os.FileInfo) string {
			if s := detectByName(f.Name()); s != "" {
				return fmt.Sprintf("%s file", s)
			}
			return "File"
		},
		"prettytime": func(t time.Time) string {
			return t.Format(time.RFC1123)
		},
		"genid": func(s string) string {
			h := fnv.New32a()
			h.Write([]byte(s))
			return fmt.Sprintf("%v", h.Sum32())
		},
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
)

type breadcrumbItem struct {
	Name, URL string
}

func handler(prefix, folderPath, givenTitle, givenColor, bannerCode string, hideLinks bool) http.Handler {
	tmpl = template.Must(template.New("http-server").
		Funcs(tmplFuncs).
		Parse(httpServerTemplate))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the method is GET, then we continue, we fail with "Method Not Allowed"
		// otherwise, since all request are for files.
		switch r.Method {
		case http.MethodGet, http.MethodHead:
		default:
			http.Error(w, "only GET and HEAD methods are allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check if the prefix isn't "/", if so, remove it
		if prefix != "/" {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)

			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
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
		fullpath, _ := filepath.Abs(filepath.Join(folderPath, r.URL.Path))

		// Find if there's a file or folder here
		info, err := os.Stat(fullpath)
		if err != nil {
			// If when trying to stat a file, the error is "not exists"
			// then we throw a 404
			if os.IsNotExist(err) {
				http.Error(w, "file or folder not found", http.StatusNotFound)
				return
			}

			http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if it's a folder, if so, walk and present the contents on screen
		if info.IsDir() {
			if !strings.HasSuffix(r.URL.Path, "/") {
				to := path.Join("/", prefix, r.URL.Path) + "/"
				http.Redirect(w, r, to, http.StatusFound)
				return
			}

			walk(prefix, fullpath, givenTitle, givenColor, bannerCode, hideLinks, w, r)
			return
		}

		f, err := os.Open(fullpath)
		if err != nil {
			http.Error(w, "error opening file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		// For content type hinting, we don't need the whole file,
		// just enough to make it work
		buf := make([]byte, 512)
		if _, err = f.Read(buf); err != nil {
			http.Error(w, "error reading file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Having the file name be empty prevents http.ServeContent from guessing
		// the resource's content type
		w.Header().Set("Content-Type", generateContentTypeCharset(info.Name(), buf))
		http.ServeContent(w, r, "", info.ModTime(), f)
	})
}

func walk(prefix, fpath, givenTitle, givenColor, bannerCode string, hideLinks bool, w http.ResponseWriter, r *http.Request) {
	// Check if there's an index file, and if so, present it on screen
	indexPath := filepath.Join(fpath, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	// If not, construct the UI we need with a list of files from this folder
	// by first opening the folder to get a Go object
	folder, err := os.Open(fpath)
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Then listing all the files in it (by passing -1 meaning all)
	list, err := folder.Readdir(-1)
	folder.Close()

	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Folders first, then alphabetically
	sort.Sort(foldersFirst(list))

	// Get the path to a parent folder
	parentFolder := ""
	if p := path.Join(prefix, r.URL.Path); p != "/" && p != prefix {
		// If the path is not root, we're in a folder, but since folders
		// are enforced to use trailing slash then we need to remove it
		// so path.Dir() can work
		parentFolder = path.Dir(strings.TrimSuffix(p, "/"))
		if !strings.HasSuffix(parentFolder, "/") {
			parentFolder += "/"
		}

		// Remove the parent folder if there's a prefix, so we don't have a parent for
		// a root in a prefixed environment
		if parentFolder == "/" && prefix != "/" {
			parentFolder = ""
		}
	}

	// If we reached this point, we're ready to print the template
	// so we create a bag, and we save the information there
	bag := map[string]interface{}{
		"Breadcrumb":  generateBreadcrumb(r.URL.Path),
		"Path":        r.URL.Path,
		"IncludeBack": parentFolder != "",
		"BackURL":     parentFolder,
		"Files":       list,
		"FilePath":    fpath,
		"PageTitle":   "HTTP File Server",
		"TagTitle":    fmt.Sprintf("Browsing directory: %s", r.URL.Path),
		"GivenColor":  givenColor,
		"PathPrefix":  prefix,
		"HideLinks":   hideLinks,
		"Banner":      bannerCode,
	}

	// Check if we need to change the title
	if givenTitle != "" {
		bag["PageTitle"] = givenTitle
		bag["TagTitle"] = givenTitle
	}

	if err := tmpl.Execute(w, bag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func logrequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to get the status code
		lrw := newLRW(w)

		// Capture the start time of this request
		// to measure how long it took
		start := time.Now()

		// Save URL before sending the next handler
		u := r.URL.String()

		// Serve the request
		next.ServeHTTP(lrw, r)

		// Now get the status code and print the log statement
		log.Printf(
			"%s %q -- %s %d %s served in %v",
			r.Method, u, r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), time.Since(start),
		)
	})
}

func generateBreadcrumb(webpath string) []breadcrumbItem {
	// We clean the parts before splitting, removing the initial and trailing slash
	// since we will take care of them later on
	parts := strings.Split(strings.Trim(webpath, "/"), "/")

	// We allocate a slice based on the length of parts plus the initial root slash
	breadcrumb := make([]breadcrumbItem, 0, len(parts)+1)

	// Adding the first element which is the root folder
	breadcrumb = append(breadcrumb, breadcrumbItem{
		Name: "/",
		URL:  "/",
	})

	// Iterate over all other parts
	for i, v := range parts {
		// If the path is empty, we just continue
		// since you can't have folders with empty names
		if v == "" {
			continue
		}

		// Append new breadcrumb and joining the path to the previous item
		breadcrumb = append(breadcrumb, breadcrumbItem{
			Name: v,
			URL:  path.Join(breadcrumb[i].URL, v) + "/",
		})
	}

	return breadcrumb
}
