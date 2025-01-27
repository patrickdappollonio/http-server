package server

import (
	"html/template"
	"path/filepath"
	"strings"
)

// filenameToIcon is a map of filenames to their respective icons
var filenameToIcon = map[string]string{
	"LICENSE": "fas fa-certificate",
}

// extensionToIcon is a map of file extensions to their respective icons
var extensionToIcon = map[string]string{
	".lock": "fas fa-lock",
	".go":   "fab fa-golang",
	".md":   "fab fa-markdown",
	".conf": "fas fa-gear",
	".sh":   "fas fa-terminal",
	".log":  "fas fa-file-lines",
}

// groupToExtension is a map of file groups to their respective extensions
var groupToExtension = map[string][]string{
	"audio": {
		"aac",
		"aiff",
		"flac",
		"m4a",
		"mp3",
		"ogg",
		"wav",
		"wma",
	},
	"code": {
		"c",
		"cpp",
		"cs",
		"css",
		"go",
		"h",
		"html",
		"java",
		"js",
		"json",
		"php",
		"rb",
		"rs",
		"sh",
		"swift",
		"toml",
		"ts",
		"txt",
		"xml",
		"yaml",
		"yml",
	},
	"image": {
		"bmp",
		"gif",
		"ico",
		"jpeg",
		"jpg",
		"png",
		"svg",
		"tif",
		"tiff",
	},
	"video": {
		"avi",
		"flv",
		"mkv",
		"mov",
		"mp4",
		"mpeg",
		"mpg",
		"ogv",
		"webm",
		"wmv",
	},
	"javascript": {
		"js",
		"jsx",
	},
	"python": {
		"py",
	},
	"font": {
		"eot",
		"otf",
		"ttf",
		"woff",
		"woff2",
	},
	"bash": {
		"sh",
		"env",
		"bash_history",
		"bash_profile",
		"bashrc",
	},
	"compressed": {
		"7z",
		"bz2",
		"gz",
		"rar",
		"tar",
		"zip",
	},
}

// groupToFilenames is a map of file groups to their respective filenames
var groupToFilenames = map[string][]string{
	"golang": {"go.mod", "go.sum"},
	"git":    {".gitignore", ".gitconfig"},
	"docker": {"Dockerfile", ".dockerignore"},
	"bash":   {"Makefile", ".profile", ".bashrc", ".bash_history"},
}

// groupToPrefixes is a map of file groups to their respective prefixes
var groupToPrefixes = map[string][]string{
	"docker": {"docker-compose", "Dockerfile", "dockerfile"},
	"python": {"Pipfile", "pipfile"},
}

// groupToIcon is a map of file groups to their respective icons
var groupToIcon = map[string]string{
	"video":      "fas fa-file-video",
	"audio":      "fas fa-file-audio",
	"image":      "fas fa-file-image",
	"code":       "fas fa-code",
	"javascript": "fab fa-js-square",
	"golang":     "fab fa-golang",
	"git":        "fab fa-git-alt",
	"docker":     "fab fa-docker",
	"font":       "fas fa-font",
	"bash":       "fas fa-terminal",
	"python":     "fab fa-python",
	"compressed": "fas fa-file-zipper",
}

// getIconForFile returns the icon class for a file based on its name
// and whether it's a folder or not.
func getIconForFile(isFolder bool, filename string) template.HTMLAttr {
	// If it's a folder, it's a quick find
	if isFolder {
		return template.HTMLAttr("fas fa-folder")
	}

	// Find it based on its filename
	if icon, ok := filenameToIcon[filename]; ok {
		return template.HTMLAttr(icon)
	}

	// Check if filename belongs to a group
	for group, filenames := range groupToFilenames {
		for _, f := range filenames {
			if f == filename {
				if icon, found := groupToIcon[group]; found {
					return template.HTMLAttr(icon)
				}
			}
		}
	}

	// Check if the file prefix belongs to a group
	for group, prefixes := range groupToPrefixes {
		for _, prefix := range prefixes {
			if strings.HasPrefix(filename, prefix) {
				if icon, found := groupToIcon[group]; found {
					return template.HTMLAttr(icon)
				}
			}
		}
	}

	// Get the extension
	extension := filepath.Ext(filename)

	// Check if it belongs to a specific icon overwrite
	if icon, ok := extensionToIcon[extension]; ok {
		return template.HTMLAttr(icon)
	}

	// Check if it belongs to a group
	for group, extensions := range groupToExtension {
		for _, ext := range extensions {
			if "."+ext == extension {
				if icon, found := groupToIcon[group]; found {
					return template.HTMLAttr(icon)
				}
			}
		}
	}

	// If we can't find the extension, use a generic icon
	return template.HTMLAttr("fas fa-file")
}
