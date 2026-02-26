package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/joetifa2003/inertigo/cmd/inertigo/generator"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Inertigo project",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			projectScope  string
			goPackageName string
			framework     string
		)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project scope (e.g. @myapp)").
					Description("Used for npm package naming: @myapp/frontend, @myapp/server, etc.").
					Value(&projectScope).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("project scope is required")
						}
						if s[0] != '@' {
							return fmt.Errorf("project scope must start with @")
						}
						if len(s) < 2 {
							return fmt.Errorf("project scope must have a name after @")
						}
						return nil
					}),

				huh.NewInput().
					Title("Go package name (e.g. github.com/user/myapp)").
					Description("Used for go.mod module name and Go imports.").
					Value(&goPackageName).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("go package name is required")
						}
						return nil
					}),

				huh.NewSelect[string]().
					Title("Select Framework").
					Options(
						huh.NewOption("React", "react"),
						huh.NewOption("Svelte", "svelte"),
					).
					Value(&framework),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Generating project with %s framework...\n", framework)

		err = generator.Generate(projectScope, goPackageName, framework)
		if err != nil {
			fmt.Printf("Error generating project: %v\n", err)
			os.Exit(1)
		}

		// Derive directory name from scope (e.g. "@myapp" -> "myapp")
		dirName := projectScope[1:]

		fmt.Println("\n✅ Project generated successfully!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", dirName)
		fmt.Println("  pnpm i")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  pnpm dev        - Start development servers (Go backend + Vite dev server)")
		fmt.Println("  pnpm dev:watch  - Same as dev but with turbo watch (auto-rebuilds on changes)")
		fmt.Println("  pnpm build      - Build everything for production")
		fmt.Println("  pnpm preview    - Build and run the production server")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
