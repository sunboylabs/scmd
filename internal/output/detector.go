package output

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// TerminalInfo contains information about the terminal environment
type TerminalInfo struct {
	IsTTY          bool
	ColorProfile   termenv.Profile
	Theme          string // "dark", "light", or "auto"
	SupportsColor  bool
	SupportsImages bool
	Width          int
	Height         int
}

// DetectTerminal detects terminal capabilities and environment
func DetectTerminal() *TerminalInfo {
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	info := &TerminalInfo{
		IsTTY:         isTTY,
		ColorProfile:  termenv.ColorProfile(),
		Theme:         DetectTheme(),
		SupportsColor: isTTY && os.Getenv("NO_COLOR") == "",
		Width:         80,
		Height:        24,
	}

	// Get terminal dimensions if available
	if isTTY {
		if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			info.Width = width
			info.Height = height
		}
	}

	// Determine if terminal supports images
	info.SupportsImages = supportsImages()

	return info
}

// DetectTheme detects terminal theme (dark or light)
func DetectTheme() string {
	// Check environment variable first
	if theme := os.Getenv("SCMD_THEME"); theme != "" {
		theme = strings.ToLower(theme)
		if theme == "dark" || theme == "light" {
			return theme
		}
	}

	// Check COLORFGBG environment variable (common in many terminals)
	// Format: "foreground;background" where 0-7 are dark, 8-15 are light
	if fgbg := os.Getenv("COLORFGBG"); fgbg != "" {
		parts := strings.Split(fgbg, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			// Background colors 0-7 indicate light text on dark background
			// Background colors 8-15 indicate dark text on light background
			if bg >= "0" && bg <= "7" {
				return "dark"
			} else if bg >= "8" {
				return "light"
			}
		}
	}

	// Check terminal background color using ANSI escape codes
	// This is experimental and may not work in all terminals
	if detected := detectBackgroundColor(); detected != "" {
		return detected
	}

	// Default to dark theme (most developer terminals are dark)
	return "dark"
}

// detectBackgroundColor attempts to detect if the terminal has a light or dark background
// Returns "dark", "light", or "" if unable to detect
func detectBackgroundColor() string {
	// This is a simplified detection - in reality, querying terminal background
	// requires complex ANSI escape sequence handling
	// For now, we'll skip this and rely on env vars
	return ""
}

// supportsImages checks if the terminal supports image rendering
func supportsImages() bool {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Kitty, iTerm2, and some others support images
	return strings.Contains(term, "kitty") ||
		termProgram == "iTerm.app" ||
		termProgram == "WezTerm"
}

// ShouldUseMarkdown determines if markdown rendering should be used
func ShouldUseMarkdown(formatFlag string, info *TerminalInfo) bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	switch formatFlag {
	case "plain":
		return false
	case "markdown":
		return true
	case "auto":
		// Auto mode: enable markdown in TTY, disable when piped
		return info.IsTTY && info.SupportsColor
	default:
		return false
	}
}

// GetTheme resolves the theme to use
func GetTheme(themeConfig string, info *TerminalInfo) string {
	if themeConfig == "auto" {
		return info.Theme
	}
	themeConfig = strings.ToLower(themeConfig)
	if themeConfig == "dark" || themeConfig == "light" {
		return themeConfig
	}
	return info.Theme
}

// GetWordWrap resolves the word wrap width
func GetWordWrap(wrapConfig int, info *TerminalInfo) int {
	if wrapConfig <= 0 {
		// Use terminal width if available, otherwise default to 80
		if info.Width > 0 && info.Width < 200 {
			return info.Width - 2 // Leave 2 char margin
		}
		return 80
	}
	return wrapConfig
}
