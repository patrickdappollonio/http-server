package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

//go:generate go run generator.go

var (
	fileServerPath = "/html"
	fileServerPort = "0.0.0.0:5000"

	pathFlag = flag.String("path", "", "The path you want to serve via HTTP")
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
		if flag.Parse(); *pathFlag != "" {
			fileServerPath = *pathFlag
		}
	}

	// Define a default title
	var givenTitle string
	if v := strings.TrimSpace(os.Getenv("FILE_SERVER_TITLE")); v != "" {
		givenTitle = v
	}

	// Check if the folder exists
	if !exists(fileServerPath) {
		log.Fatalf("Unable to start server because the path in $FILE_SERVER_PATH or --path doesn't exist: %q", fileServerPath)
	}

	// Create the file server
	http.Handle("/", logrequest(handler(fileServerPath, givenTitle)))

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
