package server

var ctypes = []struct {
	Extension   []string
	ContentType string
}{
	// Generic types
	{[]string{".aac"}, "audio/aac"},
	{[]string{".abw"}, "application/x-abiword"},
	{[]string{".apng"}, "image/apng"},
	{[]string{".arc"}, "application/x-freearc"},
	{[]string{".avif"}, "image/avif"},
	{[]string{".avi"}, "video/x-msvideo"},
	{[]string{".azw"}, "application/vnd.amazon.ebook"},
	{[]string{".bin"}, "application/octet-stream"},
	{[]string{".bmp"}, "image/bmp"},
	{[]string{".bz"}, "application/x-bzip"},
	{[]string{".bz2"}, "application/x-bzip2"},
	{[]string{".cda"}, "application/x-cdf"},
	{[]string{".csh"}, "application/x-csh"},
	{[]string{".css"}, "text/css"},
	{[]string{".csv"}, "text/csv"},
	{[]string{".doc"}, "application/msword"},
	{[]string{".docx"}, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	{[]string{".eot"}, "application/vnd.ms-fontobject"},
	{[]string{".epub"}, "application/epub+zip"},
	{[]string{".gz"}, "application/gzip"},
	{[]string{".gif"}, "image/gif"},
	{[]string{".htm", ".html"}, "text/html"},
	{[]string{".ico"}, "image/vnd.microsoft.icon"},
	{[]string{".ics"}, "text/calendar"},
	{[]string{".jar"}, "application/java-archive"},
	{[]string{".jpeg", ".jpg"}, "image/jpeg"},
	{[]string{".js", ".mjs"}, "text/javascript"},
	{[]string{".jsx"}, "text/jsx"},
	{[]string{".json"}, "application/json"},
	{[]string{".jsonld"}, "application/ld+json"},
	{[]string{".md"}, "text/markdown"},
	{[]string{".mid", ".midi"}, "audio/midi"},
	{[]string{".mp3"}, "audio/mpeg"},
	{[]string{".mp4"}, "video/mp4"},
	{[]string{".mpeg"}, "video/mpeg"},
	{[]string{".mpkg"}, "application/vnd.apple.installer+xml"},
	{[]string{".odp"}, "application/vnd.oasis.opendocument.presentation"},
	{[]string{".ods"}, "application/vnd.oasis.opendocument.spreadsheet"},
	{[]string{".odt"}, "application/vnd.oasis.opendocument.text"},
	{[]string{".oga"}, "audio/ogg"},
	{[]string{".ogv"}, "video/ogg"},
	{[]string{".ogx"}, "application/ogg"},
	{[]string{".opus"}, "audio/ogg"},
	{[]string{".otf"}, "font/otf"},
	{[]string{".png"}, "image/png"},
	{[]string{".pdf"}, "application/pdf"},
	{[]string{".php"}, "application/x-httpd-php"},
	{[]string{".ppt"}, "application/vnd.ms-powerpoint"},
	{[]string{".pptx"}, "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	{[]string{".rar"}, "application/vnd.rar"},
	{[]string{".rtf"}, "application/rtf"},
	{[]string{".sh"}, "application/x-sh"},
	{[]string{".svg"}, "image/svg+xml"},
	{[]string{".tar"}, "application/x-tar"},
	{[]string{".tif", ".tiff"}, "image/tiff"},
	{[]string{".ts"}, "video/mp2t"},
	{[]string{".tsx"}, "text/tsx"},
	{[]string{".ttf"}, "font/ttf"},
	{[]string{".txt", ".ini", ".env", ".lock", ".conf", ".gitignore", ".dockerfile"}, "text/plain"},
	{[]string{".vsd"}, "application/vnd.visio"},
	{[]string{".wav"}, "audio/wav"},
	{[]string{".weba"}, "audio/webm"},
	{[]string{".webm"}, "video/webm"},
	{[]string{".webp"}, "image/webp"},
	{[]string{".woff"}, "font/woff"},
	{[]string{".woff2"}, "font/woff2"},
	{[]string{".xhtml"}, "application/xhtml+xml"},
	{[]string{".xls"}, "application/vnd.ms-excel"},
	{[]string{".xlsx"}, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	{[]string{".xml"}, "application/xml"},
	{[]string{".xul"}, "application/vnd.mozilla.xul+xml"},
	{[]string{".yaml", ".yml"}, "application/x-yaml"},
	{[]string{".zip"}, "application/zip"},
	{[]string{".3gp", ".3gpp"}, "video/3gpp"},
	{[]string{".3g2", ".3gpp2"}, "video/3gpp2"},
	{[]string{".7z"}, "application/x-7z-compressed"},
	{[]string{".scss"}, "text/x-scss"},
	{[]string{".sass"}, "text/x-sass"},
	{[]string{".less"}, "text/css"},
	{[]string{".bat"}, "application/x-msdos-program"},
	{[]string{".bashrc"}, "application/x-shellscript"},

	// Programming languages
	{[]string{".c"}, "text/x-csrc"},
	{[]string{".h"}, "text/x-chdr"},
	{[]string{".cpp"}, "text/x-c++src"},
	{[]string{".hpp"}, "text/x-c++hdr"},
	{[]string{".java"}, "text/x-java-source"},
	{[]string{".py"}, "text/x-python"},
	{[]string{".go"}, "text/x-go"},
	{[]string{".rb"}, "application/x-ruby"},
	{[]string{".pl"}, "application/x-perl"},
	{[]string{".php"}, "application/x-httpd-php"},
	{[]string{".rs"}, "text/rust"},
	{[]string{".swift"}, "text/x-swift"},
	{[]string{".kt"}, "text/x-kotlin"},
	{[]string{".scala"}, "text/x-scala"},
	{[]string{".sh"}, "application/x-sh"},
	{[]string{".bash"}, "application/x-shellscript"},

	// Markup languages
	{[]string{".tex"}, "application/x-tex"},
	{[]string{".bib"}, "application/x-bibtex"},

	// Version control, configurations
	{[]string{"Dockerfile"}, "text/x-dockerfile"},
	{[]string{".gitignore"}, "text/plain"},
	{[]string{"Makefile"}, "text/x-makefile"},
}

func getContentTypeForExtension(extension string) string {
	for _, ct := range ctypes {
		for _, ext := range ct.Extension {
			if ext == extension {
				return ct.ContentType
			}
		}
	}

	return ""
}
