package generator

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/joetifa2003/inertigo/cmd/inertigo/templates"
)

type TemplateData struct {
	ProjectName     string
	InertigoVersion string
	SSR             bool
}

func Generate(projectName, framework string, enableSSR, installDeps bool, packageManager string) error {
	data := TemplateData{
		ProjectName:     projectName,
		InertigoVersion: "v0.1.0", // This should ideally be dynamic or latest
		SSR:             enableSSR,
	}

	// Ensure target directory exists
	targetDir := projectName
	if projectName == "." || projectName == filepath.Base(mustGetwd()) {
		targetDir = "."
	} else {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}
	}

	// Walk through the template directory
	templateRoot := framework
	// If framework is react, use the react template folder
	// Note: The structure in templates/ is currently just 'react', 'vue' (empty), 'svelte' (empty)
	// We assume the caller passes 'react', 'vue', or 'svelte' which matches the folder name.

	err := fs.WalkDir(templates.FS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Relative path from the template root
		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Destination path
		destPath := filepath.Join(targetDir, relPath)

		// Handle file renaming (remove .tmpl)
		if strings.HasSuffix(destPath, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
		}

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read template file
		content, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Check if it's a template file
		if strings.HasSuffix(path, ".tmpl") {
			// Parse and execute template
			tmpl, err := template.New(path).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", path, err)
			}

			f, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", destPath, err)
			}
			defer f.Close()

			if err := tmpl.Execute(f, data); err != nil {
				return fmt.Errorf("failed to execute template %s: %w", path, err)
			}
		} else {
			// Just copy the file
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}

	// Install dependencies if requested
	if installDeps {
		fmt.Printf("Installing dependencies using %s...\n", packageManager)
		cmd := exec.Command(packageManager, "install")
		cmd.Dir = filepath.Join(targetDir, "assets") // Dependencies are in assets folder usually?
		// Wait, cmd/react/assets has package.json. So yes, assets folder.
		// But verify if the generated structure has assets/package.json
		// Yes, template has assets/package.json.tmpl
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	// Also run 'go mod tidy' if we created a go.mod
	if _, err := os.Stat(filepath.Join(targetDir, "go.mod")); err == nil {
		fmt.Println("Running go mod tidy...")
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = targetDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go mod tidy: %v\n", err)
			// Put make this not fatal as user might not have go installed or network issues
		}
	}

	return nil
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}
