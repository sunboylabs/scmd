package output

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SyntaxHighlighter handles code syntax highlighting
type SyntaxHighlighter struct {
	style     *chroma.Style
	formatter chroma.Formatter
}

// NewSyntaxHighlighter creates a new syntax highlighter
func NewSyntaxHighlighter(styleName string) *SyntaxHighlighter {
	// Get style
	style := styles.Get(styleName)
	if style == nil {
		// Try some common styles
		if styleName == "dark" {
			style = styles.Get("monokai")
		} else if styleName == "light" {
			style = styles.Get("github")
		}

		if style == nil {
			style = styles.Fallback
		}
	}

	// Get formatter for terminal
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	return &SyntaxHighlighter{
		style:     style,
		formatter: formatter,
	}
}

// HighlightCode applies syntax highlighting to code
func (s *SyntaxHighlighter) HighlightCode(code, language string) (string, error) {
	// Get lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		// Try to match by common aliases
		lexer = matchLexerByAlias(language)
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}
	lexer = chroma.Coalesce(lexer)

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	// Format
	var buf bytes.Buffer
	err = s.formatter.Format(&buf, s.style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}

// HighlightFile highlights code from a file
func (s *SyntaxHighlighter) HighlightFile(code, filename string) (string, error) {
	language := DetectLanguage(filename)
	return s.HighlightCode(code, language)
}

// DetectLanguage tries to detect the programming language from filename
func DetectLanguage(filename string) string {
	if filename == "" {
		return ""
	}

	lexer := lexers.Match(filename)
	if lexer != nil {
		config := lexer.Config()
		if config != nil {
			return strings.ToLower(config.Name)
		}
	}

	// Manual detection for common extensions
	ext := getFileExtension(filename)
	switch ext {
	case "js", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "py":
		return "python"
	case "go":
		return "go"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "java":
		return "java"
	case "c", "h":
		return "c"
	case "cpp", "cc", "cxx", "hpp":
		return "cpp"
	case "cs":
		return "csharp"
	case "php":
		return "php"
	case "swift":
		return "swift"
	case "kt", "kts":
		return "kotlin"
	case "scala":
		return "scala"
	case "sh", "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "ps1":
		return "powershell"
	case "yaml", "yml":
		return "yaml"
	case "json":
		return "json"
	case "xml":
		return "xml"
	case "html", "htm":
		return "html"
	case "css":
		return "css"
	case "scss", "sass":
		return "scss"
	case "sql":
		return "sql"
	case "md", "markdown":
		return "markdown"
	case "tex":
		return "latex"
	case "r":
		return "r"
	case "m":
		return "matlab"
	case "jl":
		return "julia"
	case "lua":
		return "lua"
	case "vim":
		return "vim"
	case "dockerfile":
		return "dockerfile"
	case "makefile":
		return "makefile"
	default:
		return ""
	}
}

// matchLexerByAlias tries to match a lexer by common aliases
func matchLexerByAlias(language string) chroma.Lexer {
	language = strings.ToLower(language)

	aliases := map[string]string{
		"js":         "javascript",
		"jsx":        "javascript",
		"ts":         "typescript",
		"tsx":        "typescript",
		"py":         "python",
		"rb":         "ruby",
		"yml":        "yaml",
		"shell":      "bash",
		"sh":         "bash",
		"zsh":        "bash",
		"c++":        "cpp",
		"c#":         "csharp",
		"objective-c": "objc",
		"objc":       "objc",
		"golang":     "go",
		"dockerfile": "docker",
		"makefile":   "make",
	}

	if mapped, ok := aliases[language]; ok {
		return lexers.Get(mapped)
	}

	return lexers.Get(language)
}

// getFileExtension extracts the file extension
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}

	// Check for special files without extensions
	baseName := strings.ToLower(filename)
	if strings.Contains(baseName, "dockerfile") {
		return "dockerfile"
	}
	if strings.Contains(baseName, "makefile") {
		return "makefile"
	}

	return ""
}

// HighlightDiff highlights diff output
func HighlightDiff(diff string) (string, error) {
	highlighter := NewSyntaxHighlighter("dark")
	return highlighter.HighlightCode(diff, "diff")
}

// HighlightJSON highlights JSON output
func HighlightJSON(json string) (string, error) {
	highlighter := NewSyntaxHighlighter("dark")
	return highlighter.HighlightCode(json, "json")
}

// GetDefaultHighlighter returns a highlighter with default settings
func GetDefaultHighlighter() *SyntaxHighlighter {
	// Detect terminal theme
	style := detectStyle()
	if style == "light" {
		return NewSyntaxHighlighter("github")
	}
	return NewSyntaxHighlighter("monokai")
}