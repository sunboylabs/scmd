package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/scmd/scmd/internal/templates"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage prompt templates",
	Long:  "Create, list, and manage prompt templates for customized AI interactions.",
}

// templateListCmd lists available templates
var templateListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available templates",
	RunE:    runTemplateList,
}

// templateShowCmd shows a specific template
var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

// templateCreateCmd creates a new template interactively
var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template interactively",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateCreate,
}

// templateDeleteCmd deletes a template
var templateDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "Delete a template",
	Args:    cobra.ExactArgs(1),
	RunE:    runTemplateDelete,
}

// templateExportCmd exports a template to stdout
var templateExportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export template to stdout",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateExport,
}

// templateImportCmd imports a template from a file
var templateImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import template from file",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateImport,
}

// templateSearchCmd searches templates
var templateSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search templates",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateSearch,
}

// templateInitCmd initializes builtin templates
var templateInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize built-in templates",
	Long:  "Create the default built-in templates in your templates directory.",
	RunE:  runTemplateInit,
}

func init() {
	// Import flags
	templateImportCmd.Flags().Bool("force", false, "Overwrite existing template")

	// Template subcommands
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateExportCmd)
	templateCmd.AddCommand(templateImportCmd)
	templateCmd.AddCommand(templateSearchCmd)
	templateCmd.AddCommand(templateInitCmd)
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	tpls, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(tpls) == 0 {
		fmt.Println("No templates found.")
		fmt.Println("\nInitialize built-in templates with: scmd template init")
		fmt.Println("Or create your own with: scmd template create <name>")
		return nil
	}

	fmt.Println("Available Templates:\n")
	for _, t := range tpls {
		fmt.Printf("  %s (v%s)\n", t.Name, t.Version)
		fmt.Printf("    %s\n", t.Description)
		if len(t.Tags) > 0 {
			fmt.Printf("    Tags: %s\n", strings.Join(t.Tags, ", "))
		}
		fmt.Printf("    Compatible: %s\n\n", strings.Join(t.CompatibleCommands, ", "))
	}

	fmt.Printf("View details: scmd template show <name>\n")
	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	tpl, err := mgr.Load(name)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	fmt.Printf("Template: %s (v%s)\n\n", tpl.Name, tpl.Version)
	fmt.Printf("Author: %s\n", tpl.Author)
	fmt.Printf("Description: %s\n\n", tpl.Description)

	if len(tpl.Tags) > 0 {
		fmt.Printf("Tags: %s\n\n", strings.Join(tpl.Tags, ", "))
	}

	fmt.Printf("Compatible Commands: %s\n\n", strings.Join(tpl.CompatibleCommands, ", "))

	if len(tpl.Variables) > 0 {
		fmt.Println("Variables:")
		for _, v := range tpl.Variables {
			req := ""
			if v.Required {
				req = " (required)"
			}
			fmt.Printf("  - %s%s: %s\n", v.Name, req, v.Description)
			if v.Default != "" {
				fmt.Printf("    Default: %s\n", v.Default)
			}
		}
		fmt.Println()
	}

	if len(tpl.RecommendedModels) > 0 {
		fmt.Printf("Recommended Models: %s\n\n", strings.Join(tpl.RecommendedModels, ", "))
	}

	if len(tpl.Examples) > 0 {
		fmt.Println("Examples:")
		for _, ex := range tpl.Examples {
			fmt.Printf("  %s\n", ex.Description)
			fmt.Printf("  $ %s\n\n", ex.Command)
		}
	}

	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	fmt.Printf("Creating template: %s\n\n", name)

	// Interactive prompts
	description := promptString("Description: ")
	author := promptString("Author: ")
	tags := promptList("Tags (comma-separated): ")
	commands := promptList("Compatible commands (comma-separated): ")

	fmt.Println("\nEnter system prompt (defines AI behavior):")
	fmt.Println("(Press Ctrl+D when done)")
	systemPrompt := promptMultiline()

	fmt.Println("\nEnter user prompt template:")
	fmt.Println("(Use {{.VariableName}} for variables)")
	fmt.Println("(Press Ctrl+D when done)")
	userPrompt := promptMultiline()

	tpl := &templates.Template{
		Name:                name,
		Version:            "1.0",
		Author:             author,
		Description:        description,
		Tags:               tags,
		CompatibleCommands: commands,
		SystemPrompt:       systemPrompt,
		UserPromptTemplate: userPrompt,
	}

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	if err := mgr.Create(tpl); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("\nTemplate '%s' created successfully\n", name)
	fmt.Printf("\nUse with: scmd review --template %s\n", name)
	fmt.Printf("View with: scmd template show %s\n", name)

	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Confirm deletion
	fmt.Printf("Delete template '%s'? (y/N): ", name)
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" {
		fmt.Println("Cancelled.")
		return nil
	}

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	if err := mgr.Delete(name); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("Deleted template '%s'\n", name)
	return nil
}

func runTemplateExport(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	data, err := mgr.Export(name)
	if err != nil {
		return fmt.Errorf("failed to export template: %w", err)
	}

	fmt.Println(data)
	return nil
}

func runTemplateImport(cmd *cobra.Command, args []string) error {
	file := args[0]
	overwrite, _ := cmd.Flags().GetBool("force")

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	if err := mgr.Import(data, overwrite); err != nil {
		return fmt.Errorf("failed to import template: %w", err)
	}

	fmt.Printf("Imported template from %s\n", file)
	return nil
}

func runTemplateSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	matches, err := mgr.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search templates: %w", err)
	}

	if len(matches) == 0 {
		fmt.Printf("No templates found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d template(s) matching '%s':\n\n", len(matches), query)
	for _, t := range matches {
		fmt.Printf("  %s - %s\n", t.Name, t.Description)
	}

	return nil
}

func runTemplateInit(cmd *cobra.Command, args []string) error {
	mgr, err := templates.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create template manager: %w", err)
	}

	fmt.Println("Initializing built-in templates...")

	if err := mgr.InitBuiltinTemplates(); err != nil {
		return fmt.Errorf("failed to initialize templates: %w", err)
	}

	fmt.Println("Built-in templates created successfully!")
	fmt.Printf("Templates directory: %s\n\n", mgr.GetTemplateDir())
	fmt.Println("View templates with: scmd template list")

	return nil
}

// Helper functions
func promptString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptList(prompt string) []string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func promptMultiline() string {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return strings.Join(lines, "\n")
}
