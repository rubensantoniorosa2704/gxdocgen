package xpz

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/parser"
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
			parsedObjects, extractedKBName, err := parseGXExportFile(extractPath)
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

// GXPart represents a part of a GeneXus object in the XML structure
type GXPart struct {
	Type      string       `xml:"type,attr"`
	Source    string       `xml:"Source"`
	Variables []GXVariable `xml:"Variable"`
}

// GXVariable represents a variable declaration in a Part
type GXVariable struct {
	Name       string        `xml:"Name,attr"`
	Properties []GXProperty2 `xml:"Properties>Property"`
}

// GXProperty2 represents a property of a variable
type GXProperty2 struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// extractParmSignature extracts the Parm() signature from a Part and replaces Parm with the procedure name
func extractParmSignature(parts []GXPart, procedureName string) string {
	for _, part := range parts {
		// Rules part contains the Parm() declaration
		if part.Type == GXPartRules {
			signature := strings.TrimSpace(part.Source)
			// Replace "Parm(" with "ProcedureName("
			if strings.HasPrefix(signature, "Parm(") {
				signature = procedureName + "(" + strings.TrimPrefix(signature, "Parm(")
			}
			return signature
		}
	}
	return ""
}

// parseParmParameters extracts parameter information from Parm() signature
// Format: Parm(in:&Param1, out:&Param2, inout:&Param3);
func parseParmParameters(parmSignature string) []model.ParameterDoc {
	if parmSignature == "" {
		return nil
	}

	// Extract content between parentheses
	start := strings.Index(parmSignature, "(")
	end := strings.LastIndex(parmSignature, ")")
	if start == -1 || end == -1 || start >= end {
		return nil
	}

	paramsStr := parmSignature[start+1 : end]
	if strings.TrimSpace(paramsStr) == "" {
		return nil
	}

	// Split by comma
	paramParts := strings.Split(paramsStr, ",")
	var params []model.ParameterDoc

	for _, part := range paramParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse direction:&Name format
		if colonIdx := strings.Index(part, ":"); colonIdx != -1 {
			direction := strings.ToUpper(strings.TrimSpace(part[:colonIdx]))
			name := strings.TrimSpace(part[colonIdx+1:])
			// Remove & prefix if present
			name = strings.TrimPrefix(name, "&")

			params = append(params, model.ParameterDoc{
				Name:      name,
				Direction: direction,
			})
		}
	}

	return params
}

// enrichParametersWithVariableInfo adds type and description from Variable elements
func enrichParametersWithVariableInfo(params []model.ParameterDoc, parts []GXPart) []model.ParameterDoc {
	if len(params) == 0 {
		return params
	}

	// Find the Variables part (type e4c4ade7-53f0-4a56-bdfd-843735b66f47)
	var variables []GXVariable
	for _, part := range parts {
		if part.Type == "e4c4ade7-53f0-4a56-bdfd-843735b66f47" {
			variables = part.Variables
			break
		}
	}

	if len(variables) == 0 {
		return params
	}

	// Build a map of variable info by name
	varMap := make(map[string]struct {
		Description string
		Type        string
	})

	for _, v := range variables {
		var desc, typ string
		for _, prop := range v.Properties {
			switch prop.Name {
			case "Description":
				desc = prop.Value
			case "ATTCUSTOMTYPE":
				// Format: "bas:Boolean", "bc:User", "sdt:Messages, GeneXus.Common"
				// Extract the type after the colon
				if colonIdx := strings.Index(prop.Value, ":"); colonIdx != -1 {
					typ = strings.TrimSpace(prop.Value[colonIdx+1:])
					// Remove package suffix if present (e.g., "Messages, GeneXus.Common" -> "Messages")
					if commaIdx := strings.Index(typ, ","); commaIdx != -1 {
						typ = strings.TrimSpace(typ[:commaIdx])
					}
				} else {
					typ = prop.Value
				}
			}
		}
		varMap[v.Name] = struct {
			Description string
			Type        string
		}{Description: desc, Type: typ}
	}

	// Enrich parameters with variable info
	for i := range params {
		if info, exists := varMap[params[i].Name]; exists {
			if params[i].Description == "" {
				params[i].Description = info.Description
			}
			if params[i].Type == "" {
				params[i].Type = info.Type
			}
		}
	}

	return params
}

// GeneXus object type GUIDs to human-readable names
var gxTypeMap = map[string]string{
	GXTypeProcedure: "Procedure",
}

// parseGXExportFile parses a GeneXus XPZ export file and extracts all objects
func parseGXExportFile(filePath string) ([]model.GXObject, string, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer xmlFile.Close()

	type GXObjectXML struct {
		Name        string   `xml:"name,attr"`
		Type        string   `xml:"type,attr"`
		Description string   `xml:"description,attr"`
		Parent      string   `xml:"parent,attr"`
		ParentType  string   `xml:"parentType,attr"`
		Parts       []GXPart `xml:"Part"`
	}

	type VersionInfo struct {
		Name string `xml:"name,attr"`
	}

	type SourceInfo struct {
		Version VersionInfo `xml:"Version"`
	}

	type ExportFile struct {
		Source  SourceInfo    `xml:"Source"`
		Objects []GXObjectXML `xml:"Objects>Object"`
	}

	var exportFile ExportFile
	decoder := xml.NewDecoder(xmlFile)

	err = decoder.Decode(&exportFile)
	if err != nil {
		return nil, "", err
	}

	kbName := exportFile.Source.Version.Name

	if len(exportFile.Objects) == 0 {
		return nil, kbName, nil
	}

	var objects []model.GXObject
	seenObjects := make(map[string]bool)

	for _, obj := range exportFile.Objects {
		typeName := gxTypeMap[obj.Type]
		if typeName == "" {
			typeName = "Unknown"
		}

		if typeName == "Folder" {
			continue
		}

		if typeName == "Unknown" {
			continue
		}

		objKey := obj.Name + "|" + obj.Type
		if seenObjects[objKey] {
			continue
		}
		seenObjects[objKey] = true

		displayName := obj.Name
		if obj.Description != "" {
			displayName = obj.Description
		}

		// Extract source code for Procedures
		sourceCode := ""
		parmSignature := ""
		var documentation *model.DocComment

		if typeName == "Procedure" {
			sourceCode = extractProcedureSource(obj.Parts)
			parmSignature = extractParmSignature(obj.Parts, obj.Name)
			// Parse documentation from source code
			if sourceCode != "" {
				doc, err := parser.Parse(sourceCode)
				if err != nil {
					utils.Warning("Failed to parse documentation for %s: %v", obj.Name, err)
				} else {
					documentation = doc
				}
			}
			// If no documentation or no parameters in documentation, parse from Parm signature
			if documentation == nil && parmSignature != "" {
				// Create minimal documentation with parsed parameters
				params := parseParmParameters(parmSignature)
				params = enrichParametersWithVariableInfo(params, obj.Parts)
				documentation = &model.DocComment{
					IsAutoGenerated: true,
					Parameters:      params,
					Tags:            make([]string, 0),
				}
			} else if documentation != nil && len(documentation.Parameters) == 0 && parmSignature != "" {
				// Documentation exists but no @param tags, add parsed parameters
				params := parseParmParameters(parmSignature)
				params = enrichParametersWithVariableInfo(params, obj.Parts)
				documentation.Parameters = params
			}
		}

		objects = append(objects, model.GXObject{
			Name:           displayName,
			Type:           typeName,
			Path:           obj.Name,
			SourceCode:     sourceCode,
			ParmSignature:  parmSignature,
			XMLDescription: obj.Description,
			Documentation:  documentation,
		})
	}

	return objects, kbName, nil
}

// extractProcedureSource extracts source code from Procedure Parts
func extractProcedureSource(parts []GXPart) string {
	for _, part := range parts {
		if part.Type == GXPartSourceCode {
			return strings.TrimSpace(part.Source)
		}
	}
	return ""
}
