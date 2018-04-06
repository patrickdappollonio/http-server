package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

//go:generate go run generator.go

const (
	defaultColor = "indigo-red"
	cssURL       = "https://code.getmdl.io/1.3.0/material.%s.min.css"
)

var (
	fileServerPath   = "/html"
	fileServerPrefix = "/"
	fileServerPort   = "0.0.0.0:5000"

	pathFlag       = flag.String("path", "", "The path you want to serve via HTTP")
	pathprefixFlag = flag.String("pathprefix", "/", "A URL path prefix on where to serve these")
)

// exists returns whether a folder exists or not in the filesystem
func exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil || os.IsNotExist(err) {
		return false
	}

	return true
}

func main() {
	// Print usage if the number of parameters is wrong
	if len(flag.Args()) > 2 {
		flag.Usage()
		os.Exit(1)
	}

	// If there's an environment variable with the file server
	// path then use it.
	if v := os.Getenv("FILE_SERVER_PATH"); v != "" {
		fileServerPath = v
	} else {
		flag.Parse()

		if *pathFlag != "" {
			fileServerPath = *pathFlag
		}

		if *pathprefixFlag != "/" {
			fileServerPrefix = *pathprefixFlag
		}
	}

	// Check if the prefix matches what we want
	if !strings.HasSuffix(fileServerPrefix, "/") || !strings.HasPrefix(fileServerPrefix, "/") {
		log.Println("Unable to start a server with a path prefix not starting or ending in \"/\"... Aborting...")
		return
	}

	// Check if the prefix matches what we want
	if fileServerPrefix == "//" {
		log.Printf("Incorrect prefix supplied: %q. Aborting...", fileServerPrefix)
		return
	}

	// Define a default title
	var givenTitle string
	if v := strings.TrimSpace(os.Getenv("FILE_SERVER_TITLE")); v != "" {
		givenTitle = v
	}

	// Define a default color
	givenColor := fmt.Sprintf(cssURL, defaultColor)
	if v := strings.TrimSpace(os.Getenv("FILE_SERVER_COLOR_SET")); v != "" {
		// Validate the color is valid, otherwise set it to the default one
		resp, err := http.Get(fmt.Sprintf(cssURL, v))
		if err != nil {
			log.Printf("Unable to set color palette to %q. Error: %s", v, err.Error())
		}

		// Close body right away, since we don't need it
		resp.Body.Close()

		// We can use this color palette only if the status code is 200
		if resp.StatusCode == http.StatusOK {
			givenColor = resp.Request.URL.String()
		} else {
			log.Printf("Unable to set color palette to %q. Server returned status %d %s", resp.StatusCode, resp.Status)
		}
	}

	// Check if the folder exists
	if !exists(fileServerPath) {
		log.Fatalf("Unable to start server because the path in $FILE_SERVER_PATH or --path doesn't exist: %q", fileServerPath)
	}

	// Create the file server
	http.Handle(fileServerPrefix, logrequest(handler(fileServerPrefix, fileServerPath, givenTitle, givenColor)))

	// Graceful shutdown
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, os.Kill)

	// Wait signal
	close := make(chan bool, 1)

	// Create a server
	srv := &http.Server{Addr: fileServerPort}

	// Execute the server
	go func() {
		log.Printf("Starting HTTP Server. Listening at %q", srv.Addr)

		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Println(err.Error())
			} else {
				log.Println("Server closed. Bye!")
			}
			close <- true
		}
	}()

	// Check for a closing signal
	go func() {
		<-sigquit
		log.Printf("Gracefully shutting down server...")

		if srv == nil {
			close <- true
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Unable to shut down server: " + err.Error())
			close <- true
		}
	}()

	<-close
}
