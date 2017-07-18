package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	fileServerPath = "/html"
	fileServerPort = "0.0.0.0:5000"
)

func init() {
	if v := os.Getenv("FILE_SERVER_PATH"); v != "" {
		fileServerPath = v
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)

	if err != nil || os.IsNotExist(err) {
		return false
	}

	return true
}

func main() {
	// Check if the folder exists
	if !exists(fileServerPath) {
		log.Fatalf("Unable to start server because $FILE_SERVER_PATH doesn't exist: %q", fileServerPath)
	}

	// Create the file server
	fs := http.FileServer(http.Dir(fileServerPath))
	http.Handle("/", fs)

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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Unable to shut down server: " + err.Error())
			close <- true
		}
	}()

	<-close
}
