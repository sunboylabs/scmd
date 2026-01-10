package templates

// getBuiltinTemplates returns the default built-in templates
func getBuiltinTemplates() []*Template {
	return []*Template{
		securityReviewTemplate(),
		performanceTemplate(),
		apiDesignTemplate(),
		testingTemplate(),
		documentationTemplate(),
		beginnerExplainTemplate(),
	}
}

func securityReviewTemplate() *Template {
	return &Template{
		Name:        "security-review",
		Version:     "1.0",
		Author:      "scmd",
		Description: "Security-focused code review with OWASP Top 10 emphasis",
		Tags:        []string{"security", "owasp", "review"},
		CompatibleCommands: []string{"review"},
		SystemPrompt: `You are a security expert specializing in application security.
Focus on OWASP Top 10 vulnerabilities and provide actionable remediation.
Use clear severity ratings: Critical, High, Medium, Low.`,
		UserPromptTemplate: `Review the following {{.Language}} code for security issues:

Focus areas:
1. Injection attacks (SQL, Command, XSS)
2. Broken authentication
3. Sensitive data exposure
4. XML external entities (XXE)
5. Broken access control
6. Security misconfiguration
7. Cross-site scripting (XSS)
8. Insecure deserialization
9. Using components with known vulnerabilities
10. Insufficient logging & monitoring

Code:
` + "```{{.Language}}\n{{.Code}}\n```" + `

{{if .Context}}
Additional context:
{{.Context}}
{{end}}

Provide:
- Severity rating for each issue
- Specific vulnerabilities found
- Remediation steps with code examples
- Risk assessment`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language of the code", Default: "auto-detect"},
			{Name: "Code", Description: "The code to review", Required: true},
			{Name: "Context", Description: "Additional context about the codebase", Required: false},
		},
		Output: OutputConfig{
			Format: "markdown",
			Sections: []Section{
				{Title: "Security Assessment", Required: true},
				{Title: "Vulnerabilities Found", Required: true},
				{Title: "Recommendations", Required: true},
			},
		},
		RecommendedModels: []string{"gpt-4", "claude-3", "qwen2.5-7b"},
		Examples: []Example{
			{
				Description: "Review authentication code",
				Command:     "cat login.js | scmd review --template security-review",
			},
			{
				Description: "Review with context",
				Command:     "scmd review auth.py --template security-review --context 'Flask app with JWT'",
			},
		},
	}
}

func performanceTemplate() *Template {
	return &Template{
		Name:        "performance",
		Version:     "1.0",
		Author:      "scmd",
		Description: "Performance optimization and bottleneck analysis",
		Tags:        []string{"performance", "optimization", "profiling"},
		CompatibleCommands: []string{"review", "explain"},
		SystemPrompt: `You are a performance optimization expert.
Analyze code for performance bottlenecks, algorithmic complexity,
and suggest concrete optimizations.
Provide Big O analysis where relevant.`,
		UserPromptTemplate: `Analyze the following {{.Language}} code for performance:

` + "```{{.Language}}\n{{.Code}}\n```" + `

Focus on:
1. Time complexity (Big O)
2. Space complexity
3. Bottlenecks and hot paths
4. Memory allocation patterns
5. Loop optimizations
6. Caching opportunities
7. Algorithmic improvements

{{if .Context}}
Context: {{.Context}}
{{end}}

Provide:
- Current complexity analysis
- Identified bottlenecks
- Optimization suggestions with code examples
- Expected performance impact`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language", Default: "auto-detect"},
			{Name: "Code", Description: "Code to analyze", Required: true},
			{Name: "Context", Description: "Performance requirements or constraints", Required: false},
		},
		Output: OutputConfig{
			Format: "markdown",
		},
		RecommendedModels: []string{"gpt-4", "claude-3", "qwen2.5-7b"},
	}
}

func apiDesignTemplate() *Template {
	return &Template{
		Name:        "api-design",
		Version:     "1.0",
		Author:      "scmd",
		Description: "REST API design review and best practices",
		Tags:        []string{"api", "rest", "design", "architecture"},
		CompatibleCommands: []string{"review"},
		SystemPrompt: `You are an API design expert specializing in RESTful services.
Review APIs for best practices, consistency, and usability.
Focus on REST principles, HTTP semantics, and developer experience.`,
		UserPromptTemplate: `Review the following API implementation:

` + "```{{.Language}}\n{{.Code}}\n```" + `

Evaluate:
1. REST principles adherence
2. HTTP method correctness
3. Status code appropriateness
4. Request/response structure
5. Error handling
6. Versioning strategy
7. Security considerations
8. Documentation completeness

{{if .Context}}
Context: {{.Context}}
{{end}}

Provide:
- Design assessment
- Issues found
- Best practice recommendations
- Example improvements`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language", Default: "auto-detect"},
			{Name: "Code", Description: "API code to review", Required: true},
			{Name: "Context", Description: "API purpose and constraints", Required: false},
		},
		Output: OutputConfig{
			Format: "markdown",
		},
		RecommendedModels: []string{"gpt-4", "claude-3"},
	}
}

func testingTemplate() *Template {
	return &Template{
		Name:        "testing",
		Version:     "1.0",
		Author:      "scmd",
		Description: "Test coverage analysis and test generation",
		Tags:        []string{"testing", "coverage", "quality"},
		CompatibleCommands: []string{"review", "generate"},
		SystemPrompt: `You are a testing expert specializing in test-driven development.
Analyze code for testability, suggest test cases, and identify edge cases.
Focus on comprehensive coverage and meaningful assertions.`,
		UserPromptTemplate: `Analyze the following code for testing:

` + "```{{.Language}}\n{{.Code}}\n```" + `

Provide:
1. Test coverage analysis
2. Missing test scenarios
3. Edge cases to test
4. Test code examples
5. Mocking/stubbing suggestions
6. Performance test considerations

{{if .TestFramework}}
Use {{.TestFramework}} for examples.
{{end}}

Generate comprehensive test cases with clear assertions.`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language", Default: "auto-detect"},
			{Name: "Code", Description: "Code to analyze", Required: true},
			{Name: "TestFramework", Description: "Testing framework to use", Required: false},
		},
		Output: OutputConfig{
			Format: "markdown",
		},
		RecommendedModels: []string{"gpt-4", "claude-3", "qwen2.5-7b"},
	}
}

func documentationTemplate() *Template {
	return &Template{
		Name:        "documentation",
		Version:     "1.0",
		Author:      "scmd",
		Description: "Generate or review documentation",
		Tags:        []string{"docs", "documentation", "comments"},
		CompatibleCommands: []string{"explain", "generate"},
		SystemPrompt: `You are a technical writer specializing in developer documentation.
Create clear, comprehensive documentation that helps developers understand and use code effectively.
Focus on clarity, completeness, and practical examples.`,
		UserPromptTemplate: `{{if eq .Action "generate"}}Generate documentation for:{{else}}Review documentation in:{{end}}

` + "```{{.Language}}\n{{.Code}}\n```" + `

Include:
1. Purpose and overview
2. Function/method descriptions
3. Parameters and return values
4. Usage examples
5. Error handling
6. Performance considerations
7. Related functionality

{{if .Style}}
Documentation style: {{.Style}}
{{end}}

Format as {{if .Format}}{{.Format}}{{else}}markdown{{end}}.`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language", Default: "auto-detect"},
			{Name: "Code", Description: "Code to document", Required: true},
			{Name: "Action", Description: "generate or review", Default: "generate"},
			{Name: "Style", Description: "Documentation style (JSDoc, Sphinx, etc.)", Required: false},
			{Name: "Format", Description: "Output format", Default: "markdown"},
		},
		Output: OutputConfig{
			Format: "markdown",
		},
		RecommendedModels: []string{"gpt-4", "claude-3"},
	}
}

func beginnerExplainTemplate() *Template {
	return &Template{
		Name:        "beginner-explain",
		Version:     "1.0",
		Author:      "scmd",
		Description: "Explain code to beginners with simple language",
		Tags:        []string{"education", "beginner", "explain"},
		CompatibleCommands: []string{"explain"},
		SystemPrompt: `You are a patient programming teacher explaining code to beginners.
Use simple language, avoid jargon, and provide analogies.
Break down complex concepts into digestible pieces.
Include helpful context about why things work the way they do.`,
		UserPromptTemplate: `Explain the following code to a beginner:

` + "```{{.Language}}\n{{.Code}}\n```" + `

{{if .FocusOn}}
Focus especially on: {{.FocusOn}}
{{end}}

Provide:
1. Overall purpose (what does this code do?)
2. Step-by-step breakdown
3. Key concepts explained simply
4. Common beginner mistakes to avoid
5. Simple examples of how to use it
6. Related concepts to learn next

Use analogies and avoid technical jargon where possible.`,
		Variables: []Variable{
			{Name: "Language", Description: "Programming language", Default: "auto-detect"},
			{Name: "Code", Description: "Code to explain", Required: true},
			{Name: "FocusOn", Description: "Specific aspect to focus on", Required: false},
		},
		Output: OutputConfig{
			Format: "markdown",
		},
		RecommendedModels: []string{"gpt-4", "claude-3", "qwen2.5-7b"},
		Examples: []Example{
			{
				Description: "Explain a function",
				Command:     "scmd explain main.go --template beginner-explain",
			},
		},
	}
}