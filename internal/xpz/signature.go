package xpz

import (
	"regexp"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
)

// Pre-compiled regular expressions for performance
var (
	parmRegex      = regexp.MustCompile(`(?i)parm\s*\((.*?)\)`)
	paramRegex     = regexp.MustCompile(`(?i)^(in|out|inout)\s*:\s*&(.+)$`)
	directionRegex = regexp.MustCompile(`(?i)\b(in|out|inout)\s*:`)
	directionMatch = regexp.MustCompile(`(?i)\b(in|out|inout)`)
	colonSpaceRegex = regexp.MustCompile(`:\s+&`)
	commaSpaceRegex = regexp.MustCompile(`,\s*`)
	typeColonRegex = regexp.MustCompile(`:`)
)

// Signature represents a procedure's parameter signature
type Signature struct {
	Parameters     []model.ParameterDoc
	RawSignature   string
	ExtractionMode string // "ParmRule", "IsParm", "None"
}

// ExtractProcedureSignature extracts procedure signature with multi-layer fallback.
// Priority: ParmRule → Variables[@IsParm] → empty
func ExtractProcedureSignature(objNode *xmlquery.Node, procedureName string) Signature {
	// Try 1: Extract from ParmRule part (most common)
	if sig := extractFromParmRule(objNode, procedureName); sig.Parameters != nil {
		sig.ExtractionMode = "ParmRule"
		return sig
	}

	// Try 2: Extract from Variables with IsParm="true" (legacy format)
	if sig := extractFromIsParmVariables(objNode, procedureName); sig.Parameters != nil {
		sig.ExtractionMode = "IsParm"
		return sig
	}

	// No parameters found
	return Signature{
		Parameters:     []model.ParameterDoc{},
		RawSignature:   procedureName + "();",
		ExtractionMode: "None",
	}
}

// extractFromParmRule extracts parameters from ParmRule part (Rules/Parm section).
// This is the most common location for parameter declarations in GeneXus exports.
func extractFromParmRule(objNode *xmlquery.Node, procedureName string) Signature {
	source := GetText(objNode, "//Part[@type='"+GXPartRules+"']/Source")
	if source == "" {
		return Signature{}
	}
	return parseParmString(source, procedureName)
}

// extractFromIsParmVariables extracts parameters from Variables with IsParm attribute.
// This is a fallback method for older GeneXus export formats.
func extractFromIsParmVariables(objNode *xmlquery.Node, procedureName string) Signature {
	// Find Variables part using constant
	variablesPart := xmlquery.FindOne(objNode, "//Part[@type='"+GXPartVariables+"']")
	if variablesPart == nil {
		return Signature{}
	}

	// Find all Variable elements with IsParm="true"
	variables := xmlquery.Find(variablesPart, "//Variable")
	var params []model.ParameterDoc

	for _, varNode := range variables {
		isParm := false
		var name, varType, description string

		// Check properties
		for _, prop := range xmlquery.Find(varNode, "Properties/Property") {
			propName := GetText(prop, "Name")
			propValue := GetText(prop, "Value")

			switch propName {
			case "IsParm":
				isParm = (propValue == "True" || propValue == "true")
			case "Name":
				name = propValue
			case "Description":
				description = propValue
			case "ATTCUSTOMTYPE":
				varType = cleanType(propValue)
			case "idBasedOn":
				if varType == "" && strings.HasPrefix(propValue, "Attribute:") {
					// Attribute-based type
					varType = "-" // Type not available in XPZ
				}
			}
		}

		// Add parameter if marked as IsParm
		if isParm && name != "" {
			params = append(params, model.ParameterDoc{
				Name:        name,
				Direction:   "IN", // Default direction for IsParm fallback
				Type:        varType,
				Description: description,
			})
		}
	}

	if len(params) == 0 {
		return Signature{}
	}

	// Build raw signature
	rawSig := buildRawSignature(procedureName, params)

	return Signature{
		Parameters:   params,
		RawSignature: rawSig,
	}
}

// parseParmString parses a Parm(...) declaration from source text.
// It handles various formats: parm(...), Parm(...), in:/In:, out:/Out:, etc.
func parseParmString(source, procedureName string) Signature {
	// Filter out commented lines (starting with //)
	lines := strings.Split(source, "\n")
	var activeLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//") {
			activeLines = append(activeLines, line)
		}
	}
	source = strings.Join(activeLines, "\n")

	// Match: Parm(in:&Name, out:&Name, inout:&Name) using pre-compiled regex
	matches := parmRegex.FindStringSubmatch(source)
	if len(matches) < 2 {
		return Signature{}
	}

	paramsStr := matches[1]
	if strings.TrimSpace(paramsStr) == "" {
		return Signature{
			Parameters:   []model.ParameterDoc{},
			RawSignature: procedureName + "();",
		}
	}

	// Split parameters
	parts := strings.Split(paramsStr, ",")
	var params []model.ParameterDoc

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse direction:&Name or direction: &Name using pre-compiled regex
		matches := paramRegex.FindStringSubmatch(part)
		if len(matches) == 3 {
			direction := strings.ToUpper(matches[1])
			name := strings.TrimSpace(matches[2])

			params = append(params, model.ParameterDoc{
				Name:      name,
				Direction: direction,
				Type:      "", // Type will be enriched later
			})
		}
	}

	// Build raw signature
	// Replace "parm"/"Parm" (case-insensitive) with actual procedure name
	rawSig := parmRegex.ReplaceAllString(source, procedureName+"("+paramsStr+")")
	// Normalize directions to lowercase using pre-compiled regex
	rawSig = directionRegex.ReplaceAllStringFunc(rawSig, func(match string) string {
		dir := directionMatch.FindString(match)
		return strings.ToLower(dir) + ":"
	})
	// Remove spaces after colons: "in: &" -> "in:&"
	rawSig = colonSpaceRegex.ReplaceAllString(rawSig, ":&")
	// Ensure single space after commas: ",out:" -> ", out:"
	rawSig = commaSpaceRegex.ReplaceAllString(rawSig, ", ")
	rawSig = strings.TrimSpace(rawSig)

	return Signature{
		Parameters:   params,
		RawSignature: rawSig,
	}
}

// cleanType strips GeneXus type prefixes (bas:, bc:, sdt:).
func cleanType(rawType string) string {
	rawType = strings.TrimSpace(rawType)
	
	// Strip prefixes: bas:, bc:, sdt:
	if strings.Contains(rawType, ":") && !strings.HasPrefix(rawType, "Attribute:") {
		parts := strings.SplitN(rawType, ":", 2)
		if len(parts) == 2 {
			rawType = parts[1]
			
			// If type contains package (e.g., "Messages, GeneXus.Common"), keep only type
			if commaIdx := strings.Index(rawType, ","); commaIdx != -1 {
				rawType = strings.TrimSpace(rawType[:commaIdx])
			}
		}
	}
	
	return rawType
}

// buildRawSignature constructs a normalized signature string from parameters.
func buildRawSignature(procedureName string, params []model.ParameterDoc) string {
	if len(params) == 0 {
		return procedureName + "();"
	}

	var parts []string
	for _, p := range params {
		dir := strings.ToLower(p.Direction)
		// Standard format: no space after colon
		parts = append(parts, dir+":&"+p.Name)
	}

	return procedureName + "(" + strings.Join(parts, ", ") + ");"
}

// EnrichWithVariableMetadata adds type and description metadata from Variables part.
// This enriches parameters extracted from Parm() with additional metadata.
func EnrichWithVariableMetadata(params []model.ParameterDoc, objNode *xmlquery.Node) []model.ParameterDoc {
	// Find Variables part using constant
	variablesPart := xmlquery.FindOne(objNode, "//Part[@type='"+GXPartVariables+"']")
	if variablesPart == nil {
		return params
	}

	// Build map of variable metadata
	varMap := make(map[string]struct {
		Type        string
		Description string
	})

	for _, varNode := range xmlquery.Find(variablesPart, "//Variable") {
		name := GetAttrDirect(varNode, "Name")
		if name == "" {
			continue
		}

		var varType, description string
		for _, prop := range xmlquery.Find(varNode, "Properties/Property") {
			propName := GetText(prop, "Name")
			propValue := GetText(prop, "Value")

			switch propName {
			case "Description":
				description = propValue
			case "ATTCUSTOMTYPE":
				varType = cleanType(propValue)
			case "idBasedOn":
				if varType == "" && strings.HasPrefix(propValue, "Attribute:") {
					varType = "-" // Type not in XPZ
				}
			}
		}

		varMap[name] = struct {
			Type        string
			Description string
		}{Type: varType, Description: description}
	}

	// Enrich parameters
	for i := range params {
		if meta, exists := varMap[params[i].Name]; exists {
			if params[i].Type == "" {
				params[i].Type = meta.Type
			}
			if params[i].Description == "" {
				params[i].Description = meta.Description
			}
		}
	}

	return params
}
