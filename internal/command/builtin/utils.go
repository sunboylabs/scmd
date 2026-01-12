package builtin

import (
	"os"
	"path/filepath"
	"strings"
)

// detectLanguage tries to detect programming language from filename and content
func detectLanguage(filename, content string) string {
	// Try to detect from filename extension
	ext := filepath.Ext(filename)
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".rs":
		return "rust"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".cs":
		return "csharp"
	case ".r", ".R":
		return "r"
	case ".m":
		return "matlab"
	case ".jl":
		return "julia"
	case ".lua":
		return "lua"
	case ".sh", ".bash":
		return "bash"
	case ".ps1":
		return "powershell"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".yaml", ".yml":
		return "yaml"
	case ".md", ".markdown":
		return "markdown"
	default:
		// Try to detect from content (basic heuristics)
		if strings.Contains(content, "package main") {
			return "go"
		}
		if strings.Contains(content, "def ") && strings.Contains(content, ":") {
			return "python"
		}
		if strings.Contains(content, "function ") || strings.Contains(content, "const ") {
			return "javascript"
		}
		return ""
	}
}

// isFile checks if a path points to a file (not a directory)
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
