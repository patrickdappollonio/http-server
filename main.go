package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultColor = "indigo-red"
	cssURL       = "https://code.getmdl.io/1.3.0/material.%s.min.css"
)

//go:embed template.tmpl
var httpServerTemplate string

var (
	pathFlag       = flag.String("path", "", "The path you want to serve via HTTP")
	pathprefixFlag = flag.String("pathprefix", "/", "A URL path prefix on where to serve these")
	portFlag       = flag.String("port", "5000", "The port you want to serve via HTTP")
	bannerFlag     = flag.String("banner", "", "The HTML code you want to show on the top of the page")
)

func main() {
	// Print usage if the number of parameters is wrong
	if len(flag.Args()) > 2 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse all flags
	flag.Parse()

	var (
		fileServerPath   = firstNonEmpty("/html", *pathFlag, envany("FILE_SERVER_PATH"))
		fileServerPrefix = firstNonEmpty("/", *pathprefixFlag)
		fileServerPort   = firstNonEmpty("5000", *portFlag, envany("FILE_SERVER_PORT", "PORT"))

		fileServerUsername = firstNonEmpty("", envany("FILE_SERVER_USERNAME", "HTTP_USER"))
		fileServerPassword = firstNonEmpty("", envany("FILE_SERVER_PASSWORD", "HTTP_PASS"))

		givenTitle = firstNonEmpty("", envany("FILE_SERVER_TITLE", "PAGE_TITLE"))
		givenColor = firstNonEmpty(defaultColor, envany("FILE_SERVER_COLOR_SET", "COLOR_SET"))

		hideSourceCodeLinks = firstNonEmpty("", envany("FILE_SERVER_HIDE_LINKS", "HIDE_LINKS")) != ""
		bannerCode          = firstNonEmpty("", *bannerFlag, envany("FILE_SERVER_BANNER", "BANNER"))
	)

	// Check if the prefix matches what we want
	if !strings.HasSuffix(fileServerPrefix, "/") || !strings.HasPrefix(fileServerPrefix, "/") {
		log.Fatalf("Unable to start a server with a path prefix not starting or ending in %q... Aborting...", "/")
		return
	}

	// Check if the prefix matches what we want
	if fileServerPrefix == "//" {
		log.Fatalf("Incorrect prefix supplied: %q. Aborting...", fileServerPrefix)
		return
	}

	// Define a default color
	if givenColor != defaultColor {
		if !isAvailableColor(givenColor) {
			log.Fatalf("Unable to set color palette to %q. The color palette does not exist in getmdl.io", givenColor)
			return
		}
	}

	// Convert givenColor to a URL
	givenColor = fmt.Sprintf(cssURL, givenColor)

	// Check if the folder exists
	if !exists(fileServerPath) {
		log.Fatalf("Unable to start server because the path in $FILE_SERVER_PATH or --path doesn't exist: %q", fileServerPath)
		return
	}

	// Generic middlewares for all paths
	paths := chain(logrequest)

	// Check if needs authentication and if so, extend the middlewares
	if fileServerUsername != "" && fileServerPassword != "" {
		paths = paths.extend(basicAuth(fileServerUsername, fileServerPassword))
	}

	// Create the file server
	http.Handle(fileServerPrefix,
		paths.then(
			handler(fileServerPrefix, fileServerPath, givenTitle, givenColor, bannerCode, hideSourceCodeLinks),
		),
	)

	// Check whether or not the fileServerPrefix is set, if so, then
	// simply create a temporary redirect to the new path. Also add a listener
	// for the prefix without the slash at the end so it goes to "/"
	if fileServerPrefix != "/" {
		handleredir := paths.then(redirect("/", fileServerPrefix))
		handlesubpath := paths.then(redirect("*", fileServerPrefix))

		http.Handle("/", handleredir)
		http.Handle(strings.TrimSuffix(fileServerPrefix, "/"), handlesubpath)
	}

	// Graceful shutdown
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, syscall.SIGTERM)

	// Wait signal
	close := make(chan bool, 1)

	// Create a server
	srv := &http.Server{Addr: ":" + fileServerPort}

	// Execute the server
	go func() {
		log.Printf("Starting HTTP Server. Listening at %q", srv.Addr)

		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err.Error())
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
