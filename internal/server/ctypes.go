package server

var ctypes = []struct {
	Extension   []string
	ExactNames  []string
	ContentType string
}{
	// Audio Formats
	{[]string{".aac"}, nil, "audio/aac"},
	{[]string{".aiff", ".aif"}, nil, "audio/aiff"}, // Audio Interchange File Format
	{[]string{".flac"}, nil, "audio/flac"},         // Free Lossless Audio Codec
	{[]string{".m4a"}, nil, "audio/mp4"},           // MPEG-4 Audio
	{[]string{".mid", ".midi"}, nil, "audio/midi"},
	{[]string{".mp3"}, nil, "audio/mpeg"},
	{[]string{".oga"}, nil, "audio/ogg"},
	{[]string{".opus"}, nil, "audio/ogg"},
	{[]string{".wav"}, nil, "audio/wav"},
	{[]string{".weba"}, nil, "audio/webm"},
	{[]string{".wma"}, nil, "audio/x-ms-wma"}, // Windows Media Audio
	{[]string{".adp"}, nil, "audio/adpcm"},
	{[]string{".amr"}, nil, "audio/amr"},
	{[]string{".kar"}, nil, "audio/midi"},
	{[]string{".mpga"}, nil, "audio/mpeg"},
	{[]string{".ra"}, nil, "audio/x-realaudio"},
	{[]string{".ram"}, nil, "audio/x-pn-realaudio"},
	{[]string{".s3m"}, nil, "audio/s3m"},
	{[]string{".xm"}, nil, "audio/xm"},

	// Video Formats
	{[]string{".3gp", ".3gpp"}, nil, "video/3gpp"},
	{[]string{".3g2", ".3gpp2"}, nil, "video/3gpp2"},
	{[]string{".avi"}, nil, "video/x-msvideo"},
	{[]string{".flv"}, nil, "video/x-flv"},      // Flash Video
	{[]string{".mkv"}, nil, "video/x-matroska"}, // Matroska Video
	{[]string{".mp4"}, nil, "video/mp4"},
	{[]string{".mov"}, nil, "video/quicktime"},
	{[]string{".mpeg"}, nil, "video/mpeg"},
	{[]string{".ogv"}, nil, "video/ogg"},
	{[]string{".webm"}, nil, "video/webm"},
	{[]string{".ts"}, nil, "video/mp2t"},
	{[]string{".wmv"}, nil, "video/x-ms-wmv"}, // Windows Media Video
	{[]string{".mj2"}, nil, "video/mj2"},
	{[]string{".mjp2"}, nil, "video/mj2"},
	{[]string{".f4v"}, nil, "video/x-f4v"},
	{[]string{".dvb"}, nil, "video/vnd.dvb.file"},
	{[]string{".dpg"}, nil, "application/vnd.dpgraph"},

	// Image Formats
	{[]string{".apng"}, nil, "image/apng"},
	{[]string{".avif"}, nil, "image/avif"},
	{[]string{".bmp"}, nil, "image/bmp"},
	{[]string{".gif"}, nil, "image/gif"},
	{[]string{".heic"}, nil, "image/heic"},
	{[]string{".heif"}, nil, "image/heif"}, // High Efficiency Image Format
	{[]string{".jpeg", ".jpg"}, nil, "image/jpeg"},
	{[]string{".png"}, nil, "image/png"},
	{[]string{".raw"}, nil, "image/x-raw"}, // Raw Image Formats
	{[]string{".svg"}, nil, "image/svg+xml"},
	{[]string{".svgz"}, nil, "image/svg+xml"}, // Compressed SVG
	{[]string{".tif", ".tiff"}, nil, "image/tiff"},
	{[]string{".webp"}, nil, "image/webp"},
	{[]string{".jpe"}, nil, "image/jpeg"},
	{[]string{".jfif"}, nil, "image/jpeg"},
	{[]string{".jxr"}, nil, "image/jxr"},
	{[]string{".jng"}, nil, "image/x-jng"},
	{[]string{".dpx"}, nil, "image/dpx"},
	{[]string{".fits"}, nil, "image/fits"},
	{[]string{".heics"}, nil, "image/heic-sequence"},
	{[]string{".heifs"}, nil, "image/heif-sequence"},

	// Document Formats
	{[]string{".doc"}, nil, "application/msword"},
	{[]string{".docx"}, nil, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	{[]string{".odf"}, nil, "application/vnd.oasis.opendocument.formula"}, // Open Document Format for Formula
	{[]string{".pdf"}, nil, "application/pdf"},
	{[]string{".rtf"}, nil, "application/rtf"},
	{[]string{".txt", ".ini", ".env", ".lock", ".conf", ".dockerfile"}, nil, "text/plain"},
	{[]string{".md"}, nil, "text/markdown"},
	{[]string{".tex"}, nil, "application/x-tex"},
	{[]string{".bib"}, nil, "application/x-bibtex"},
	{[]string{".docm"}, nil, "application/vnd.ms-word.document.macroenabled.12"},
	{[]string{".dot"}, nil, "application/msword"},
	{[]string{".dotm"}, nil, "application/vnd.ms-word.template.macroenabled.12"},
	{[]string{".dotx"}, nil, "application/vnd.openxmlformats-officedocument.wordprocessingml.template"},
	{[]string{".xml"}, nil, "application/xml"},
	{[]string{".xhtml"}, nil, "application/xhtml+xml"},
	{[]string{".xul"}, nil, "application/vnd.mozilla.xul+xml"},
	{[]string{".csv"}, nil, "text/csv"}, // Comma-Separated Values
	{[]string{".json"}, nil, "application/json"},
	{[]string{".json5"}, nil, "application/json5"}, // JSON with comments
	{[]string{".markdown"}, nil, "text/markdown"},
	{[]string{".man"}, nil, "text/troff"},
	{[]string{".dtd"}, nil, "application/xml-dtd"},
	{[]string{".mml"}, nil, "text/mathml"},
	{[]string{".xlf"}, nil, "application/xliff+xml"},

	// Spreadsheet Formats
	{[]string{".ods"}, nil, "application/vnd.oasis.opendocument.spreadsheet"},
	{[]string{".xls"}, nil, "application/vnd.ms-excel"},
	{[]string{".xlsx"}, nil, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	{[]string{".xltx"}, nil, "application/vnd.openxmlformats-officedocument.spreadsheetml.template"}, // Excel Template

	// Presentation Formats
	{[]string{".odp"}, nil, "application/vnd.oasis.opendocument.presentation"},
	{[]string{".ppt"}, nil, "application/vnd.ms-powerpoint"},
	{[]string{".pptx"}, nil, "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	{[]string{".pot"}, nil, "application/vnd.ms-powerpoint"},
	{[]string{".potm"}, nil, "application/vnd.ms-powerpoint.template.macroenabled.12"},
	{[]string{".potx"}, nil, "application/vnd.openxmlformats-officedocument.presentationml.template"},
	{[]string{".ppam"}, nil, "application/vnd.ms-powerpoint.addin.macroenabled.12"},
	{[]string{".ppd"}, nil, "application/vnd.cups-ppd"},

	// Programming Languages
	{[]string{".c"}, nil, "text/x-csrc"},
	{[]string{".css"}, nil, "text/css"},
	{[]string{".cpp"}, nil, "text/x-c++src"},
	{[]string{".h"}, nil, "text/x-chdr"},
	{[]string{".hpp"}, nil, "text/x-c++hdr"},
	{[]string{".dart"}, nil, "application/dart"}, // Dart Language
	{[]string{".go"}, nil, "text/x-go"},
	{[]string{".java"}, nil, "text/x-java-source"},
	{[]string{".kt"}, nil, "text/x-kotlin"},
	{[]string{".pl"}, nil, "application/x-perl"},
	{[]string{".py"}, nil, "text/x-python"},
	{[]string{".rb"}, nil, "application/x-ruby"},
	{[]string{".rs"}, nil, "text/rust"},
	{[]string{".scala"}, nil, "text/x-scala"},
	{[]string{".sh"}, nil, "application/x-sh"},
	{[]string{".swift"}, nil, "text/x-swift"},
	{[]string{".bash"}, nil, "application/x-shellscript"},
	{[]string{".bashrc"}, nil, "application/x-shellscript"},
	{[]string{".js", ".mjs"}, nil, "text/javascript"},
	{[]string{".jsx"}, nil, "text/jsx"},
	{[]string{".tsx"}, nil, "text/tsx"},

	// Font Files
	{[]string{".eot"}, nil, "application/vnd.ms-fontobject"},
	{[]string{".otf"}, nil, "font/otf"},
	{[]string{".ttf"}, nil, "font/ttf"},
	{[]string{".woff"}, nil, "font/woff"},
	{[]string{".woff2"}, nil, "font/woff2"},

	// Compressed and Archive Files
	{[]string{".7z"}, nil, "application/x-7z-compressed"},
	{[]string{".arc"}, nil, "application/x-freearc"},
	{[]string{".bin"}, nil, "application/octet-stream"},
	{[]string{".bz"}, nil, "application/x-bzip"},
	{[]string{".bz2"}, nil, "application/x-bzip2"},
	{[]string{".gz"}, nil, "application/gzip"},
	{[]string{".rar"}, nil, "application/vnd.rar"},
	{[]string{".tar"}, nil, "application/x-tar"},
	{[]string{".xz"}, nil, "application/x-xz"},   // XZ Compressed File
	{[]string{".lz"}, nil, "application/x-lzip"}, // Lzip Compressed File
	{[]string{".zip"}, nil, "application/zip"},
	{[]string{".z"}, nil, "application/x-compress"},
	{[]string{".zlib"}, nil, "application/x-zlib"},

	// Configuration and Dependency Files
	{nil, []string{"Dockerfile"}, "text/x-dockerfile"},
	{nil, []string{"Gemfile"}, "text/plain"},
	{nil, []string{"Makefile"}, "text/x-makefile"},
	{nil, []string{"Pipfile"}, "text/plain"},
	{nil, []string{"package.json"}, "application/json"},
	{nil, []string{"package-lock.json"}, "application/json"},
	{nil, []string{"yarn.lock"}, "text/plain"},
	{nil, []string{"pom.xml"}, "application/xml"},
	{nil, []string{"build.gradle"}, "text/x-gradle"},
	{nil, []string{"requirements.txt"}, "text/plain"}, // Python Requirements File
	{nil, []string{"Cargo.toml"}, "text/plain"},       // Rust Package Manager File

	// Security Files
	{[]string{".crt"}, nil, "application/x-x509-ca-cert"},
	{[]string{".pem"}, nil, "application/x-pem-file"},
	{[]string{".p12"}, nil, "application/x-pkcs12"}, // PKCS#12 File
	{[]string{".pfx"}, nil, "application/x-pkcs12"}, // PKCS#12 File

	// Miscellaneous
	{[]string{".html", ".htm"}, nil, "text/html"},
	{[]string{".ico"}, nil, "image/vnd.microsoft.icon"},
	{[]string{".ics"}, nil, "text/calendar"},
	{[]string{".scss"}, nil, "text/x-scss"},
	{[]string{".sass"}, nil, "text/x-sass"},
	{[]string{".less"}, nil, "text/css"},
	{[]string{".yaml", ".yml"}, nil, "application/x-yaml"},
	{[]string{".log"}, nil, "text/plain"},
	{[]string{".xsd"}, nil, "application/xml"},
	{[]string{".xsl"}, nil, "application/xml"},
	{[]string{".xslt"}, nil, "application/xslt+xml"},
	{[]string{".srt"}, nil, "application/x-subrip"},
	{[]string{".sql"}, nil, "application/sql"},
	{[]string{".tgz"}, nil, "application/x-gzip"},
}

func getContentTypeForFilename(name string) string {
	for _, ct := range ctypes {
		for _, internalName := range ct.ExactNames {
			if name == internalName {
				return ct.ContentType
			}
		}

		for _, ext := range ct.Extension {
			if len(name) >= len(ext) && name[len(name)-len(ext):] == ext {
				return ct.ContentType
			}
		}
	}

	return ""
}
