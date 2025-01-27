package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/patrickdappollonio/http-server/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	configFilePrefix = ".http-server" // no extension, cobra will pick from several options
	envVarPrefix     = "file_server_" // case insensitive, it's uppercased in code
)

var version = "development"

func run() error {
	// Server and settings holder
	var srv server.Server

	// Define the config prefix for config files
	srv.ConfigFilePrefix = configFilePrefix

	// Create a logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Configure max processes
	undoFn, err := maxprocs.Set()
	if err != nil {
		logger.Printf("Unable to set max procs: %s -- will continue without setting them", err)
		undoFn()
	}

	// Create a piped reader/writer for logging
	// then intercept logging statements as they
	// come to prepend dates
	pr, pw := io.Pipe()
	go sendPipeToLogger(logger, pr)

	// Create the root command
	rootCmd := &cobra.Command{
		Use:           "http-server",
		Short:         "A simple HTTP server and a directory listing tool.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,

		// Bind viper settings against the root command
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return bindCobraAndViper(cmd)
		},

		// Execute the server
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Set the message output to the appropriate writer
			srv.LogOutput = cmd.OutOrStdout()
			srv.SetVersion(version)

			// Validate fields to make sure they're correct
			if err := srv.Validate(); err != nil {
				return fmt.Errorf("unable to validate configuration: %w", err)
			}

			// Load redirections file if enabled
			if err := srv.LoadRedirectionsIfEnabled(); err != nil {
				return fmt.Errorf("unable to load redirections file: %w", err)
			}

			// Print some sane defaults and some information about the request
			srv.PrintStartup()

			// Run the server
			return srv.ListenAndServe()
		},
	}

	// Customize writer inside the command
	rootCmd.SetOut(pw)

	// Configure a custom help command to avoid writing to the customized pipe
	originalHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.SetOut(os.Stdout)
		originalHelp(cmd, args)
	})

	// Configure a custom usage command to avoid writing to the customized pipe
	origUsage := rootCmd.UsageFunc()
	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.SetOut(os.Stdout)
		return origUsage(cmd)
	})

	// Define the flags for the root command
	flags := rootCmd.Flags()
	flags.IntVarP(&srv.Port, "port", "p", 5000, "port to configure the server to listen on")
	flags.StringVarP(&srv.Path, "path", "d", "./", "path to the directory you want to serve")
	flags.StringVar(&srv.PathPrefix, "pathprefix", "/", "path prefix for the URL where the server will listen on")
	flags.BoolVar(&srv.CorsEnabled, "cors", false, "enable CORS support by setting the \"Access-Control-Allow-Origin\" header to \"*\"")
	flags.StringVar(&srv.Username, "username", "", "username for basic authentication")
	flags.StringVar(&srv.Password, "password", "", "password for basic authentication")
	flags.StringVar(&srv.PageTitle, "title", "", "title of the directory listing page")
	flags.BoolVar(&srv.HideLinks, "hide-links", false, "hide the links to this project's source code visible in the header and footer")
	flags.BoolVar(&srv.DisableCacheBuster, "disable-cache-buster", false, "disable the cache buster for assets from the directory listing feature")
	flags.BoolVar(&srv.DisableMarkdown, "disable-markdown", false, "disable the markdown rendering feature")
	flags.BoolVar(&srv.MarkdownBeforeDir, "markdown-before-dir", false, "render markdown content before the directory listing")
	flags.StringVar(&srv.JWTSigningKey, "jwt-key", "", "signing key for JWT authentication")
	flags.BoolVar(&srv.ValidateTimedJWT, "ensure-unexpired-jwt", false, "enable time validation for JWT claims \"exp\" and \"nbf\"")
	flags.StringVar(&srv.BannerMarkdown, "banner", "", "markdown text to be rendered at the top of the directory listing page")
	flags.BoolVar(&srv.ETagDisabled, "disable-etag", false, "disable etag header generation")
	flags.StringVar(&srv.ETagMaxSize, "etag-max-size", "5M", "maximum size for etag header generation, where bigger size = more memory usage")
	flags.BoolVar(&srv.GzipEnabled, "gzip", false, "enable gzip compression for supported content-types")
	flags.BoolVar(&srv.DisableRedirects, "disable-redirects", false, "disable redirection file handling")
	flags.BoolVar(&srv.DisableDirectoryList, "disable-directory-listing", false, "disable the directory listing feature and return 404s for directories without index")
	flags.StringVar(&srv.CustomNotFoundPage, "custom-404", "", "custom \"page not found\" to serve")
	flags.IntVar(&srv.CustomNotFoundStatusCode, "custom-404-code", 0, "custtom status code for pages not found")

	//nolint:wrapcheck // no need to wrap this error
	return rootCmd.Execute()
}

// sendPipeToLogger reads from the pipe and sends the output to the logger
func sendPipeToLogger(logger *log.Logger, pipe io.Reader) {
	// Scan the log messages per line
	scanner := bufio.NewScanner(pipe)

	// Print every new line to the logger
	for scanner.Scan() {
		logger.Println(scanner.Text())
	}

	// Err renders the first non-EOF error found
	if err := scanner.Err(); err != nil {
		logger.Println("Error writing pipe:", err)
	}
}

// A list of cobra flags that should be ignored from automatic
// environment variable binding generation
var ignoredFlags = map[string]struct{}{
	"help":    {},
	"version": {},
}

// A list of cobra flags that need the long form of the environment
// variable name, because the short form can be ambiguous
var skipShortVersionFlag = map[string]struct{}{
	"path": {},
}

// A set of cobra flag names to environment variable aliases
// user to maintain backwards compatibility
var bindingAliases = map[string][]string{
	"username":   {"http_user", envVarPrefix + "username"},
	"password":   {"http_pass", envVarPrefix + "password"},
	"pathprefix": {envVarPrefix + "path_prefix"},
	"title":      {envVarPrefix + "page_title"},
}

// binds the cobra command flags against the viper configuration
func bindCobraAndViper(rootCommand *cobra.Command) error {
	v := viper.New()

	// Attempt to read settings from a config file from multiple
	// different file types
	v.SetConfigName(configFilePrefix)
	v.SetConfigType("yaml")

	// Look for the config file in these locations
	v.AddConfigPath(".")

	// Try to read the config file from any of the locations
	if err := v.ReadInConfig(); err != nil {
		// If the configuration file was not found, it's all good, we ignore
		// the failure and proceed with the default settings
		if !errors.Is(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("unable to read configuration file: %w", err)
		}
	}

	// Anonymous function to potentially log when we bind an
	// environment variable to a cobra flag
	bind := func(flagName, envVar string) {
		v.BindEnv(flagName, envVar)
	}

	// Configure prefixes for environment variables and
	// set backwards-compatible environment variables
	rootCommand.Flags().VisitAll(func(f *pflag.Flag) {
		// Skip those flags that don't need to be bound
		if _, ok := ignoredFlags[f.Name]; ok {
			return
		}

		// Generate a new name without dashes, replaced to underscores
		newName := strings.ReplaceAll(f.Name, "-", "_")

		// Bind the key to the new environment variable name, uppercased,
		// and dashes replaced with underscore
		if _, found := skipShortVersionFlag[f.Name]; !found {
			bind(f.Name, strings.ToUpper(newName))
		}

		// Bind the key to the new environment variable name including
		// the prefix, uppercased, and dashes replaced with underscore
		bind(f.Name, strings.ToUpper(envVarPrefix+newName))

		// Bind potential aliases of the environment variables to maintain
		// backwards compatibility
		if aliases, found := bindingAliases[f.Name]; found {
			for _, alias := range aliases {
				bind(f.Name, strings.ToUpper(alias))
			}
		}

		// If the flag hasn't been changed, and the value is set in
		// the environment, set the flag to the value from the environment
		if !f.Changed && v.IsSet(f.Name) {
			rootCommand.Flags().Set(f.Name, v.GetString(f.Name))
		}
	})

	return nil
}
