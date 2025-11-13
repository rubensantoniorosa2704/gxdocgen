package parser

import (
	"regexp"
	"strings"

	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
)

// Parse extracts and parses documentation comments from GeneXus source code
func Parse(sourceCode string) (*model.DocComment, error) {
	commentBlock := extractCommentBlock(sourceCode)
	if commentBlock == "" {
		return nil, nil
	}

	doc := &model.DocComment{
		Parameters: make([]model.ParameterDoc, 0),
		Tags:       make([]string, 0),
	}

	lines := strings.Split(commentBlock, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "@") {
			parseTag(line, doc)
		}
	}

	return doc, nil
}

// extractCommentBlock finds and extracts the /** ... */ comment block
func extractCommentBlock(source string) string {
	re := regexp.MustCompile(`(?s)/\*\*\s*(.*?)\s*\*/`)
	matches := re.FindStringSubmatch(source)

	if len(matches) < 2 {
		return ""
	}

	block := matches[1]

	// Remove leading * from each line
	lines := strings.Split(block, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

// parseTag processes a single @tag line
func parseTag(line string, doc *model.DocComment) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 1 {
		return
	}

	tag := parts[0]
	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
	}

	switch tag {
	case "@package":
		doc.Package = value
	case "@summary":
		doc.Summary = value
	case "@description":
		doc.Description = value
	case "@author":
		doc.Author = value
	case "@created":
		doc.Created = value
	case "@param":
		param := parseParameter(value)
		if param != nil {
			doc.Parameters = append(doc.Parameters, *param)
		}
	case "@return":
		doc.Return = value
	case "@tag":
		doc.Tags = append(doc.Tags, value)
	case "@deprecated":
		doc.Deprecated = true
		doc.DeprecationNote = value
	}
}

// parseParameter parses a @param line
// Format: @param name [IN|OUT|INOUT] Type:TypeName - Description
func parseParameter(value string) *model.ParameterDoc {
	// Split by " - " to separate description
	parts := strings.SplitN(value, " - ", 2)
	paramPart := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Parse "name direction type"
	tokens := strings.Fields(paramPart)
	if len(tokens) < 2 {
		return nil
	}

	param := &model.ParameterDoc{
		Name:        tokens[0],
		Description: description,
	}

	// Check if second token is direction or type
	direction := strings.ToUpper(tokens[1])
	if direction == "IN" || direction == "OUT" || direction == "INOUT" {
		param.Direction = direction
		if len(tokens) > 2 {
			param.Type = tokens[2]
		}
	} else {
		param.Direction = "IN"
		param.Type = tokens[1]
	}

	return param
}
