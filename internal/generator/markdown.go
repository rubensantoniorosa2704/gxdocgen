package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/utils"
)

const version = "0.1.1"

// sanitizePackageName ensures package names are safe for use as filenames
func sanitizePackageName(pkg string) string {
	if pkg == "" {
		return "root"
	}
	// Replace characters that are unsafe for filenames
	pkg = strings.ReplaceAll(pkg, "/", "-")
	pkg = strings.ReplaceAll(pkg, "\\", "-")
	pkg = strings.ReplaceAll(pkg, ":", "-")
	pkg = strings.ReplaceAll(pkg, "*", "-")
	pkg = strings.ReplaceAll(pkg, "?", "-")
	pkg = strings.ReplaceAll(pkg, "\"", "-")
	pkg = strings.ReplaceAll(pkg, "<", "-")
	pkg = strings.ReplaceAll(pkg, ">", "-")
	pkg = strings.ReplaceAll(pkg, "|", "-")
	return strings.TrimSpace(pkg)
}

// GenerateDocs generates Markdown documentation from extracted GeneXus objects
func GenerateDocs(objects []model.GXObject, kbName string, outputDir string) error {
	utils.Info("Generating Markdown documentation in: %s", outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Separate Procedures from other objects
	var procedures []model.GXObject
	var otherObjects []model.GXObject
	var undocumentedCount int

	for _, obj := range objects {
		if obj.Type == "Procedure" {
			procedures = append(procedures, obj)
			if obj.Documentation == nil {
				undocumentedCount++
				utils.Warning("Procedure '%s' has no documentation comments", obj.Name)
			}
		} else {
			otherObjects = append(otherObjects, obj)
		}
	}

	// Generate individual Procedure documentation files
	for _, proc := range procedures {
		if err := generateProcedureDoc(proc, outputDir); err != nil {
			utils.Warning("Failed to generate docs for %s: %v", proc.Name, err)
		}
	}

	// Generate package index files
	if err := generatePackageIndexes(procedures, outputDir); err != nil {
		utils.Warning("Failed to generate package indexes: %v", err)
	}

	// Generate main README file with KB name
	readmeFilename := "README.md"
	if kbName != "" {
		readmeFilename = kbName + ".md"
	}
	readmePath := filepath.Join(outputDir, readmeFilename)
	if err := generateReadme(objects, procedures, kbName, readmePath); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	utils.Success("Documentation generated successfully at: %s", outputDir)
	if len(procedures) > 0 {
		utils.Info("Generated %d Procedure documentation file(s)", len(procedures))
		if undocumentedCount > 0 {
			utils.Warning("%d procedure(s) are missing /** */ documentation comments", undocumentedCount)
		}
	}
	return nil
}

// generateReadme creates a README.md file listing all extracted objects
func generateReadme(objects []model.GXObject, procedures []model.GXObject, kbName string, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Build markdown content
	var sb strings.Builder

	// Header
	if kbName != "" {
		sb.WriteString("# " + kbName + " Documentation\n\n")
	} else {
		sb.WriteString("# GeneXus Documentation\n\n")
	}
	sb.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Total Objects: **%d**\n\n", len(objects)))

	// Statistics by type
	typeCount := make(map[string]int)
	for _, obj := range objects {
		objType := obj.Type
		if objType == "" {
			objType = "Unknown"
		}
		typeCount[objType]++
	}

	if len(typeCount) > 0 {
		sb.WriteString("## Object Statistics\n\n")
		sb.WriteString("| Type | Count |\n")
		sb.WriteString("|------|-------|\n")
		for objType, count := range typeCount {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", objType, count))
		}
		sb.WriteString("\n")
	}

	// List packages if we have documented procedures
	if len(procedures) > 0 {
		packageMap := make(map[string]int)
		for _, proc := range procedures {
			if proc.Documentation != nil {
				pkg := sanitizePackageName(proc.Documentation.Package)
				packageMap[pkg]++
			}
		}

		if len(packageMap) > 0 {
			sb.WriteString("## Packages\n\n")
			sb.WriteString("| Package | Procedures |\n")
			sb.WriteString("|---------|------------|\n")
			for pkg, count := range packageMap {
				link := fmt.Sprintf("[%s](./%s.md)", pkg, pkg)
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", link, count))
			}
			sb.WriteString("\n")
		}
	}

	// List all objects
	sb.WriteString("## Extracted Objects\n\n")

	if len(objects) == 0 {
		sb.WriteString("*No objects found in the XPZ file.*\n")
	} else {
		sb.WriteString("| Name | Type | Path |\n")
		sb.WriteString("|------|------|------|\n")

		for _, obj := range objects {
			name := obj.Name
			if name == "" {
				name = "*unnamed*"
			}
			objType := obj.Type
			if objType == "" {
				objType = "Unknown"
			}
			path := obj.Path
			if path == "" {
				path = "-"
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n", name, objType, path))
		}
	}

	sb.WriteString("\n---\n")
	sb.WriteString(fmt.Sprintf("*Generated by GXDocGen v%s*\n", version))

	// Write to file
	_, err = file.WriteString(sb.String())
	return err
}

// generateProcedureDoc generates a Markdown file for a single Procedure
func generateProcedureDoc(proc model.GXObject, outputDir string) error {
	doc := proc.Documentation

	// Create filename from procedure name
	filename := filepath.Join(outputDir, proc.Path+".md")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb strings.Builder

	// Title (from @summary or name)
	title := proc.Name
	if doc != nil && doc.Summary != "" {
		title = doc.Summary
	}
	sb.WriteString("# " + title + "\n\n")

	// Package badge
	if doc != nil && doc.Package != "" {
		pkgName := sanitizePackageName(doc.Package)
		sb.WriteString("**Package:** [`" + doc.Package + "`](./" + pkgName + ".md)\n\n")
	}

	// Function signature
	if proc.ParmSignature != "" {
		sb.WriteString("## Signature\n\n")
		sb.WriteString("```genexus\n")
		sb.WriteString(proc.ParmSignature + "\n")
		sb.WriteString("```\n\n")
	}

	// Deprecation warning
	if doc != nil && doc.Deprecated {
		sb.WriteString("⚠️ **DEPRECATED**")
		if doc.DeprecationNote != "" {
			sb.WriteString(": " + doc.DeprecationNote)
		}
		sb.WriteString("\n\n")
	}

	// Description
	description := ""
	if doc != nil && doc.Description != "" {
		description = doc.Description
	} else if proc.XMLDescription != "" {
		description = proc.XMLDescription
	}

	if description != "" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(description + "\n\n")
	}

	// Parameters
	if doc != nil && len(doc.Parameters) > 0 {
		sb.WriteString("## Parameters\n\n")
		sb.WriteString("| Name | Direction | Type | Description |\n")
		sb.WriteString("|------|-----------|------|-------------|\n")

		for _, param := range doc.Parameters {
			name := param.Name
			if name == "" {
				name = "-"
			}
			direction := param.Direction
			if direction == "" {
				direction = "IN"
			}
			paramType := param.Type
			if paramType == "" {
				paramType = "-"
			}
			desc := param.Description
			if desc == "" {
				desc = "-"
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				name, direction, paramType, desc))
		}
		sb.WriteString("\n")
	}

	// Return type
	if doc != nil && doc.Return != "" {
		sb.WriteString("## Return\n\n")
		sb.WriteString(doc.Return + "\n\n")
	}

	// Metadata footer
	sb.WriteString("---\n\n")
	if doc != nil && !doc.IsAutoGenerated {
		if doc.Author != "" {
			sb.WriteString("**Author:** " + doc.Author + "  \n")
		}
		if doc.Created != "" {
			sb.WriteString("**Created:** " + doc.Created + "  \n")
		}
	} else if doc != nil && doc.IsAutoGenerated {
		// Indicate auto-generated documentation
		sb.WriteString("*⚠️ Auto-generated from XML metadata. Add `/** */` annotations for detailed documentation.*\n\n")
	}

	sb.WriteString(fmt.Sprintf("\n*Generated by GXDocGen v%s*\n", version))

	// Write to file
	_, err = file.WriteString(sb.String())
	return err
}

// generatePackageIndexes creates package-level index files
func generatePackageIndexes(procedures []model.GXObject, outputDir string) error {
	// Group procedures by package
	packageMap := make(map[string][]model.GXObject)
	
	for _, proc := range procedures {
		pkg := "root"
		if proc.Documentation != nil && proc.Documentation.Package != "" {
			pkg = sanitizePackageName(proc.Documentation.Package)
		}
		packageMap[pkg] = append(packageMap[pkg], proc)
	}

	// Generate index file for each package
	for pkg, procs := range packageMap {
		filename := filepath.Join(outputDir, pkg+".md")
		if err := generatePackageIndex(pkg, procs, filename); err != nil {
			return err
		}
	}

	return nil
}

// generatePackageIndex creates an index file for a package
func generatePackageIndex(packageName string, procedures []model.GXObject, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var sb strings.Builder

	// Title
	sb.WriteString("# Package: " + packageName + "\n\n")
	sb.WriteString("## Procedures\n\n")

	// Table of procedures
	sb.WriteString("| Procedure | Summary |\n")
	sb.WriteString("|-----------|----------|\n")

	for _, proc := range procedures {
		name := proc.Path
		summary := proc.Name
		if proc.Documentation != nil && proc.Documentation.Summary != "" {
			summary = proc.Documentation.Summary
		}

		// Link to procedure file
		link := fmt.Sprintf("[%s](./%s.md)", name, proc.Path)

		sb.WriteString(fmt.Sprintf("| %s | %s |\n", link, summary))
	}

	sb.WriteString("\n---\n")
	sb.WriteString(fmt.Sprintf("*Generated by GXDocGen v%s*\n", version))

	// Write to file
	_, err = file.WriteString(sb.String())
	return err
}
