package xpz

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/utils"
)

// GeneXus object type GUIDs
const (
	GXTypeProcedure = "84a12160-f59b-4ad7-a683-ea4481ac23e9"
)

// GeneXus Part type GUIDs
const (
	GXPartSourceCode = "528d1c06-a9c2-420d-bd35-21dca83f12ff" // Source code part
	GXPartRules      = "9b0a32a3-de6d-4be1-a4dd-1b85d3741534" // Rules/Parm part
	GXPartVariables  = "e4c4ade7-53f0-4a56-bdfd-843735b66f47" // Variables part
)

// ExtractResult contains the extraction results
type ExtractResult struct {
	Objects []model.GXObject
	KBName  string
}

// Extract extracts and parses a GeneXus XPZ file
// Returns extraction results including objects and KB name
func Extract(path string) (*ExtractResult, error) {
	// Validate that the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("XPZ file not found: %s", path)
	}

	utils.Info("Opening XPZ file: %s", path)

	// Open the zip archive
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open XPZ archive: %w", err)
	}
	defer reader.Close()

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "gxdocgen-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory
	utils.Info("Extracting to temporary directory: %s", tempDir)

	var objects []model.GXObject
	kbName := ""

	// Iterate through files in the archive
	for _, file := range reader.File {
		// Extract the file
		extractPath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(extractPath, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", extractPath, err)
			}
			continue
		}

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(extractPath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s: %w", extractPath, err)
		}

		// Extract file content
		if err := extractFile(file, extractPath); err != nil {
			return nil, fmt.Errorf("failed to extract %s: %w", file.Name, err)
		}

		// Parse XML files to identify GeneXus objects
		if strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			// Check if this is the main GeneXus export file
			parsedObjects, extractedKBName, err := parseGXExportFileXMLQuery(extractPath)
			if err != nil {
				utils.Warning("Failed to parse %s: %v", file.Name, err)
				continue
			}
			if kbName == "" && extractedKBName != "" {
				kbName = extractedKBName
			}
			if len(parsedObjects) > 0 {
				// This is the main export file with all objects
				objects = append(objects, parsedObjects...)
				utils.Info("Found %d objects in %s", len(parsedObjects), file.Name)
			}
		}
	}

	utils.Success("Extracted %d GeneXus objects", len(objects))
	return &ExtractResult{
		Objects: objects,
		KBName:  kbName,
	}, nil
}

// extractFile extracts a single file from the zip archive
func extractFile(file *zip.File, destPath string) error {
	// Open the file in the archive
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the content
	_, err = io.Copy(destFile, srcFile)
	return err
}

// GeneXus object type GUIDs to human-readable names
var gxTypeMap = map[string]string{
	GXTypeProcedure: "Procedure",
}
