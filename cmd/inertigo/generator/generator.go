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
	ProjectScope  string // e.g. "@myapp"
	GoPackageName string // e.g. "github.com/user/myapp"
}

func Generate(projectScope, goPackageName, framework string) error {
	data := TemplateData{
		ProjectScope:  projectScope,
		GoPackageName: goPackageName,
	}

	// Derive directory name from scope (e.g. "@myapp" -> "myapp")
	dirName := projectScope[1:]

	// Ensure target directory exists
	targetDir := dirName
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Walk through the template directory
	templateRoot := framework

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

	// Run 'go mod tidy' in the backend directory
	backendDir := filepath.Join(targetDir, "backend")
	if _, err := os.Stat(filepath.Join(backendDir, "go.mod")); err == nil {
		fmt.Println("Running go mod tidy...")
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = backendDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to run go mod tidy: %v\n", err)
		}
	}

	return nil
}
