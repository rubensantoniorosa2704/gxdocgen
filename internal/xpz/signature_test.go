package xpz

import (
	"strings"
	"testing"

	"github.com/antchfx/xmlquery"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/model"
)

func TestExtractProcedureSignature_ParmRule(t *testing.T) {
	xmlContent := `
	<Object>
		<Part type="9b0a32a3-de6d-4be1-a4dd-1b85d3741534">
			<Source><![CDATA[Parm(in:&UserID, out:&UserName);]]></Source>
		</Part>
	</Object>
	`
	
	doc, err := xmlquery.Parse(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	sig := ExtractProcedureSignature(doc, "GetUser")
	
	if sig.ExtractionMode != "ParmRule" {
		t.Errorf("Expected extraction mode 'ParmRule', got '%s'", sig.ExtractionMode)
	}
	
	if len(sig.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(sig.Parameters))
	}
	
	if sig.Parameters[0].Name != "UserID" || sig.Parameters[0].Direction != "IN" {
		t.Errorf("First parameter incorrect: %+v", sig.Parameters[0])
	}
	
	if sig.Parameters[1].Name != "UserName" || sig.Parameters[1].Direction != "OUT" {
		t.Errorf("Second parameter incorrect: %+v", sig.Parameters[1])
	}
}

func TestExtractProcedureSignature_IsParm(t *testing.T) {
	xmlContent := `
	<Object>
		<Part type="e4c4ade7-53f0-4a56-bdfd-843735b66f47">
			<Variable Name="UserID">
				<Properties>
					<Property><Name>IsParm</Name><Value>True</Value></Property>
					<Property><Name>Name</Name><Value>UserID</Value></Property>
					<Property><Name>ATTCUSTOMTYPE</Name><Value>bas:Numeric</Value></Property>
					<Property><Name>Description</Name><Value>User identifier</Value></Property>
				</Properties>
			</Variable>
			<Variable Name="UserName">
				<Properties>
					<Property><Name>IsParm</Name><Value>True</Value></Property>
					<Property><Name>Name</Name><Value>UserName</Value></Property>
					<Property><Name>ATTCUSTOMTYPE</Name><Value>bas:Character</Value></Property>
				</Properties>
			</Variable>
		</Part>
	</Object>
	`
	
	doc, err := xmlquery.Parse(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	sig := ExtractProcedureSignature(doc, "GetUser")
	
	if sig.ExtractionMode != "IsParm" {
		t.Errorf("Expected extraction mode 'IsParm', got '%s'", sig.ExtractionMode)
	}
	
	if len(sig.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(sig.Parameters))
	}
	
	// Check type extraction
	if sig.Parameters[0].Type != "Numeric" {
		t.Errorf("Expected type 'Numeric', got '%s'", sig.Parameters[0].Type)
	}
}

func TestExtractProcedureSignature_NoParams(t *testing.T) {
	xmlContent := `<Object></Object>`
	
	doc, err := xmlquery.Parse(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	sig := ExtractProcedureSignature(doc, "DoSomething")
	
	if sig.ExtractionMode != "None" {
		t.Errorf("Expected extraction mode 'None', got '%s'", sig.ExtractionMode)
	}
	
	if len(sig.Parameters) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(sig.Parameters))
	}
	
	if sig.RawSignature != "DoSomething();" {
		t.Errorf("Expected signature 'DoSomething();', got '%s'", sig.RawSignature)
	}
}

func TestCleanType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bas:Boolean", "Boolean"},
		{"bc:User", "User"},
		{"sdt:Messages, GeneXus.Common", "Messages"},
		{"Character", "Character"},
		{"Attribute:UserId", "Attribute:UserId"}, // Keep Attribute: prefix
	}
	
	for _, tt := range tests {
		result := cleanType(tt.input)
		if result != tt.expected {
			t.Errorf("cleanType(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestEnrichWithVariableMetadata(t *testing.T) {
	xmlContent := `
	<Object>
		<Part type="e4c4ade7-53f0-4a56-bdfd-843735b66f47">
			<Variable Name="UserID">
				<Properties>
					<Property><Name>Description</Name><Value>User identifier</Value></Property>
					<Property><Name>ATTCUSTOMTYPE</Name><Value>bas:Numeric</Value></Property>
				</Properties>
			</Variable>
			<Variable Name="IsActive">
				<Properties>
					<Property><Name>Description</Name><Value>Active status</Value></Property>
					<Property><Name>ATTCUSTOMTYPE</Name><Value>bas:Boolean</Value></Property>
				</Properties>
			</Variable>
		</Part>
	</Object>
	`
	
	doc, err := xmlquery.Parse(strings.NewReader(xmlContent))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	params := []model.ParameterDoc{
		{Name: "UserID", Direction: "IN"},
		{Name: "IsActive", Direction: "OUT"},
	}

	enriched := EnrichWithVariableMetadata(params, doc)
	
	if enriched[0].Type != "Numeric" {
		t.Errorf("Expected type 'Numeric' for UserID, got '%s'", enriched[0].Type)
	}
	
	if enriched[0].Description != "User identifier" {
		t.Errorf("Expected description 'User identifier', got '%s'", enriched[0].Description)
	}
	
	if enriched[1].Type != "Boolean" {
		t.Errorf("Expected type 'Boolean' for IsActive, got '%s'", enriched[1].Type)
	}
}

func TestParseParmString(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		procName     string
		expectParams int
	}{
		{
			name:         "Two parameters",
			source:       "Parm(in:&UserID, out:&UserName);",
			procName:     "GetUser",
			expectParams: 2,
		},
		{
			name:         "No parameters",
			source:       "Parm();",
			procName:     "DoSomething",
			expectParams: 0,
		},
		{
			name:         "Mixed directions",
			source:       "Parm(in:&X, out:&Y, inout:&Z);",
			procName:     "Process",
			expectParams: 3,
		},
		{
			name:         "Lowercase parm",
			source:       "parm(in:&VagId,out:&Messages,out:&Sucesso);",
			procName:     "prAlterarSituacaoVaga",
			expectParams: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := parseParmString(tt.source, tt.procName)
			
			if len(sig.Parameters) != tt.expectParams {
				t.Errorf("Expected %d parameters, got %d", tt.expectParams, len(sig.Parameters))
			}
		})
	}
}
