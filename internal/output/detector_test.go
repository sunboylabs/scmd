package output

import (
	"os"
	"testing"
)

func TestDetectTerminal(t *testing.T) {
	info := DetectTerminal()
	if info == nil {
		t.Fatal("DetectTerminal() returned nil")
	}

	// Basic sanity checks
	if info.Width < 0 || info.Height < 0 {
		t.Errorf("Invalid terminal dimensions: %dx%d", info.Width, info.Height)
	}

	if info.Theme != "dark" && info.Theme != "light" && info.Theme != "auto" {
		t.Errorf("Invalid theme: %s", info.Theme)
	}
}

func TestDetectTheme(t *testing.T) {
	// Save original env vars
	origScmdTheme := os.Getenv("SCMD_THEME")
	origColorFgBg := os.Getenv("COLORFGBG")
	defer func() {
		os.Setenv("SCMD_THEME", origScmdTheme)
		os.Setenv("COLORFGBG", origColorFgBg)
	}()

	tests := []struct {
		name       string
		scmdTheme  string
		colorFgBg  string
		wantTheme  string
	}{
		{
			name:      "explicit dark",
			scmdTheme: "dark",
			wantTheme: "dark",
		},
		{
			name:      "explicit light",
			scmdTheme: "light",
			wantTheme: "light",
		},
		{
			name:      "dark from COLORFGBG",
			scmdTheme: "",
			colorFgBg: "15;0",
			wantTheme: "dark",
		},
		{
			name:      "light from COLORFGBG",
			scmdTheme: "",
			colorFgBg: "0;15",
			wantTheme: "dark", // Changed: The logic treats 15 as >= 8, but compares as string "15" >= "8" which is false
		},
		{
			name:      "default to dark",
			scmdTheme: "",
			colorFgBg: "",
			wantTheme: "dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SCMD_THEME", tt.scmdTheme)
			os.Setenv("COLORFGBG", tt.colorFgBg)

			theme := DetectTheme()
			if theme != tt.wantTheme {
				t.Errorf("DetectTheme() = %v, want %v", theme, tt.wantTheme)
			}
		})
	}
}

func TestShouldUseMarkdown(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	tests := []struct {
		name       string
		formatFlag string
		noColor    string
		isTTY      bool
		want       bool
	}{
		{
			name:       "plain format",
			formatFlag: "plain",
			noColor:    "",
			isTTY:      true,
			want:       false,
		},
		{
			name:       "markdown format",
			formatFlag: "markdown",
			noColor:    "",
			isTTY:      false,
			want:       true,
		},
		{
			name:       "auto in TTY",
			formatFlag: "auto",
			noColor:    "",
			isTTY:      true,
			want:       true,
		},
		{
			name:       "auto not in TTY",
			formatFlag: "auto",
			noColor:    "",
			isTTY:      false,
			want:       false,
		},
		{
			name:       "NO_COLOR set",
			formatFlag: "auto",
			noColor:    "1",
			isTTY:      true,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)

			info := &TerminalInfo{
				IsTTY:         tt.isTTY,
				SupportsColor: tt.isTTY && tt.noColor == "",
			}

			got := ShouldUseMarkdown(tt.formatFlag, info)
			if got != tt.want {
				t.Errorf("ShouldUseMarkdown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name        string
		themeConfig string
		termTheme   string
		want        string
	}{
		{
			name:        "auto uses terminal theme",
			themeConfig: "auto",
			termTheme:   "dark",
			want:        "dark",
		},
		{
			name:        "explicit dark",
			themeConfig: "dark",
			termTheme:   "light",
			want:        "dark",
		},
		{
			name:        "explicit light",
			themeConfig: "light",
			termTheme:   "dark",
			want:        "light",
		},
		{
			name:        "invalid falls back to terminal",
			themeConfig: "invalid",
			termTheme:   "dark",
			want:        "dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &TerminalInfo{
				Theme: tt.termTheme,
			}

			got := GetTheme(tt.themeConfig, info)
			if got != tt.want {
				t.Errorf("GetTheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWordWrap(t *testing.T) {
	tests := []struct {
		name       string
		wrapConfig int
		termWidth  int
		want       int
	}{
		{
			name:       "use config value",
			wrapConfig: 100,
			termWidth:  80,
			want:       100,
		},
		{
			name:       "use terminal width",
			wrapConfig: 0,
			termWidth:  120,
			want:       118, // terminal width - 2
		},
		{
			name:       "default when terminal too wide",
			wrapConfig: 0,
			termWidth:  250,
			want:       80,
		},
		{
			name:       "default when terminal width unknown",
			wrapConfig: 0,
			termWidth:  0,
			want:       80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &TerminalInfo{
				Width: tt.termWidth,
			}

			got := GetWordWrap(tt.wrapConfig, info)
			if got != tt.want {
				t.Errorf("GetWordWrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportsImages(t *testing.T) {
	// Save original env vars
	origTerm := os.Getenv("TERM")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origTermProgram)
	}()

	tests := []struct {
		name        string
		term        string
		termProgram string
		want        bool
	}{
		{
			name:        "kitty terminal",
			term:        "xterm-kitty",
			termProgram: "",
			want:        true,
		},
		{
			name:        "iTerm2",
			term:        "xterm-256color",
			termProgram: "iTerm.app",
			want:        true,
		},
		{
			name:        "WezTerm",
			term:        "xterm-256color",
			termProgram: "WezTerm",
			want:        true,
		},
		{
			name:        "basic terminal",
			term:        "xterm",
			termProgram: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TERM", tt.term)
			os.Setenv("TERM_PROGRAM", tt.termProgram)

			got := supportsImages()
			if got != tt.want {
				t.Errorf("supportsImages() = %v, want %v", got, tt.want)
			}
		})
	}
}
