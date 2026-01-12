package output

import (
	"strings"
	"testing"
)

// Sample markdown content for benchmarks
var (
	simpleMarkdown = "# Hello World\n\nThis is a simple test."

	mediumMarkdown = `# Documentation

## Introduction

This is a medium-sized markdown document with various features.

### Features

- **Bold text** support
- *Italic text* support
- ` + "`inline code`" + ` support
- Lists and nested items

### Code Example

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

More content follows here with additional paragraphs.
`

	largeMarkdown = strings.Repeat(mediumMarkdown, 10)
)

// BenchmarkFormatter_Creation tests formatter creation performance
// Target: < 1ms
func BenchmarkFormatter_Creation(b *testing.B) {
	opts := &FormatterOptions{
		Format:   "auto",
		Theme:    "dark",
		WordWrap: 80,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewFormatter(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFormatter_GetDefault tests default formatter creation
// Target: < 1ms
func BenchmarkFormatter_GetDefault(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetDefaultFormatter()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderer_Creation tests renderer creation (lazy)
// Target: < 1ms (should be instant as it's lazy)
func BenchmarkRenderer_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMarkdownRenderer("dark", 80)
	}
}

// BenchmarkFormatter_RenderSimple tests rendering simple markdown
// Target: < 10ms
func BenchmarkFormatter_RenderSimple(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "markdown",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := formatter.Render(simpleMarkdown)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFormatter_RenderMedium tests rendering medium markdown
// Target: < 10ms
func BenchmarkFormatter_RenderMedium(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "markdown",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := formatter.Render(mediumMarkdown)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFormatter_RenderLarge tests rendering large markdown
func BenchmarkFormatter_RenderLarge(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "markdown",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := formatter.Render(largeMarkdown)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFormatter_RenderPlain tests plain text passthrough
// Target: < 1ms (should be near-instant)
func BenchmarkFormatter_RenderPlain(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "plain",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := formatter.Render(mediumMarkdown)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderer_LazyInit tests lazy initialization overhead
func BenchmarkRenderer_LazyInit(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer := NewMarkdownRenderer("dark", 80)
		_, _ = renderer.Render(simpleMarkdown)
	}
}

// BenchmarkRenderer_Reuse tests reusing initialized renderer
func BenchmarkRenderer_Reuse(b *testing.B) {
	renderer := NewMarkdownRenderer("dark", 80)
	// Initialize once
	_, _ = renderer.Render(simpleMarkdown)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := renderer.Render(mediumMarkdown)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDetectTerminal tests terminal detection performance
// Target: < 1ms
func BenchmarkDetectTerminal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectTerminal()
	}
}

// BenchmarkDetectTheme tests theme detection performance
// Target: < 1ms
func BenchmarkDetectTheme(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DetectTheme()
	}
}

// BenchmarkShouldUseMarkdown tests markdown detection
func BenchmarkShouldUseMarkdown(b *testing.B) {
	info := &TerminalInfo{
		IsTTY:         true,
		SupportsColor: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShouldUseMarkdown("auto", info)
	}
}

// BenchmarkRenderError tests styled message rendering
func BenchmarkRenderError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderError("test error message")
	}
}

// BenchmarkRenderSuccess tests styled message rendering
func BenchmarkRenderSuccess(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderSuccess("test success message")
	}
}

// BenchmarkBufferedRender tests buffered rendering
func BenchmarkBufferedRender(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "plain",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer := NewBufferedRender(formatter)
		buffer.Write([]byte(mediumMarkdown))
		_, _ = buffer.Flush()
	}
}

// BenchmarkFormatter_Parallel tests concurrent formatter usage
func BenchmarkFormatter_Parallel(b *testing.B) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "plain", // Use plain to avoid concurrent issues with renderer
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = formatter.Render(simpleMarkdown)
		}
	})
}

// BenchmarkGetFormatterFromConfig tests config-based creation
func BenchmarkGetFormatterFromConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetFormatterFromConfig("auto", "dark", 80)
		if err != nil {
			b.Fatal(err)
		}
	}
}
