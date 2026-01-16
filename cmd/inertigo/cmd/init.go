package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/joetifa2003/inertigo/cmd/inertigo/generator"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Inertigo project",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			projectName    string
			framework      string
			enableSSR      bool
			installDeps    bool
			packageManager string
		)

		cwd, _ := os.Getwd()
		defaultName := filepath.Base(cwd)

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project Name").
					Value(&projectName).
					Placeholder(defaultName),

				huh.NewSelect[string]().
					Title("Select Framework").
					Options(
						huh.NewOption("React", "react"),
						huh.NewOption("Vue", "vue"),
						huh.NewOption("Svelte", "svelte"),
					).
					Value(&framework),

				huh.NewConfirm().
					Title("Enable SSR?").
					Value(&enableSSR),
			),

			huh.NewGroup(
				huh.NewConfirm().
					Title("Install dependencies now?").
					Value(&installDeps),
			),

			// Third group for package manager, only shown if installDeps is true
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select Package Manager").
					Options(
						huh.NewOption("npm", "npm"),
						huh.NewOption("pnpm", "pnpm"),
						huh.NewOption("yarn", "yarn"),
						huh.NewOption("bun", "bun"),
					).
					Value(&packageManager),
			).WithHideFunc(func() bool {
				return !installDeps
			}),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if projectName == "" {
			projectName = defaultName
		}

		fmt.Printf("Generating project %s with %s (SSR: %v)...\n", projectName, framework, enableSSR)

		err = generator.Generate(projectName, framework, enableSSR, installDeps, packageManager)
		if err != nil {
			fmt.Printf("Error generating project: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Project generated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
