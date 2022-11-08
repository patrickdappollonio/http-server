package server

import (
	"html/template"
	"path/filepath"
	"strings"
)

var filenameToIcon = map[string]string{
	"LICENSE": "fas fa-certificate",
}

var extensionToIcon = map[string]string{
	".lock": "fas fa-lock",
	".go":   "fab fa-golang",
	".md":   "fab fa-markdown",
	".conf": "fas fa-gear",
	".sh":   "fas fa-terminal",
	".log":  "fas fa-file-lines",
}

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

var groupToFilenames = map[string][]string{
	"golang": {"go.mod", "go.sum"},
	"git":    {".gitignore", ".gitconfig"},
	"docker": {"Dockerfile", ".dockerignore"},
	"bash":   {"Makefile", ".profile", ".bashrc", ".bash_history"},
}

var groupToPrefixes = map[string][]string{
	"docker": {"docker-compose", "Dockerfile", "dockerfile"},
	"python": {"Pipfile", "pipfile"},
}

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
