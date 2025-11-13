package parser

import (
	"testing"
)

func TestParse_ValidComment(t *testing.T) {
	sourceCode := `/**
 * @package users
 * @summary Get User By ID
 * @description Retrieves user information from the database based on the provided user ID.
 * @param UserID IN Numeric:UserId - The unique identifier of the user
 * @param User OUT User - User business component with complete information
 * @author Jane Smith
 * @created 2025-11-13
 */

&User.Load(&UserID)
If &User.Fail()
	&Messages = &User.GetMessages()
EndIf`

	doc, err := Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected doc to be non-nil")
	}

	// Test package
	if doc.Package != "users" {
		t.Errorf("Expected package 'users', got '%s'", doc.Package)
	}

	// Test summary
	if doc.Summary != "Get User By ID" {
		t.Errorf("Expected summary 'Get User By ID', got '%s'", doc.Summary)
	}

	// Test description
	expected := "Retrieves user information from the database based on the provided user ID."
	if doc.Description != expected {
		t.Errorf("Expected description '%s', got '%s'", expected, doc.Description)
	}

	// Test author
	if doc.Author != "Jane Smith" {
		t.Errorf("Expected author 'Jane Smith', got '%s'", doc.Author)
	}

	// Test created
	if doc.Created != "2025-11-13" {
		t.Errorf("Expected created '2025-11-13', got '%s'", doc.Created)
	}

	// Test parameters
	if len(doc.Parameters) != 2 {
		t.Fatalf("Expected 2 parameters, got %d", len(doc.Parameters))
	}

	// Check first parameter
	param1 := doc.Parameters[0]
	if param1.Name != "UserID" {
		t.Errorf("Expected param name 'UserID', got '%s'", param1.Name)
	}
	if param1.Direction != "IN" {
		t.Errorf("Expected param direction 'IN', got '%s'", param1.Direction)
	}
	if param1.Type != "Numeric:UserId" {
		t.Errorf("Expected param type 'Numeric:UserId', got '%s'", param1.Type)
	}
	if param1.Description != "The unique identifier of the user" {
		t.Errorf("Expected param description, got '%s'", param1.Description)
	}

	// Check second parameter
	param2 := doc.Parameters[1]
	if param2.Name != "User" {
		t.Errorf("Expected param name 'User', got '%s'", param2.Name)
	}
	if param2.Direction != "OUT" {
		t.Errorf("Expected param direction 'OUT', got '%s'", param2.Direction)
	}
}

func TestParse_NoComment(t *testing.T) {
	sourceCode := `&User.Load(&UserID)
If &User.Fail()
	&Messages = &User.GetMessages()
EndIf`

	doc, err := Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if doc != nil {
		t.Error("Expected doc to be nil when no comment block exists")
	}
}

func TestParse_DeprecatedTag(t *testing.T) {
	sourceCode := `/**
 * @package legacy
 * @summary Old Login Method
 * @description Legacy authentication procedure
 * @deprecated Use NewAuthenticateUser instead
 */
Parm();`

	doc, err := Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if !doc.Deprecated {
		t.Error("Expected Deprecated to be true")
	}

	if doc.DeprecationNote != "Use NewAuthenticateUser instead" {
		t.Errorf("Expected deprecation note, got '%s'", doc.DeprecationNote)
	}
}

func TestParse_ReturnTag(t *testing.T) {
	sourceCode := `/**
 * @package utils
 * @summary Calculate Total
 * @description Calculates the total amount
 * @return Numeric - The calculated total
 */
Parm();`

	doc, err := Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if doc.Return != "Numeric - The calculated total" {
		t.Errorf("Expected return value, got '%s'", doc.Return)
	}
}

func TestParse_MultipleTags(t *testing.T) {
	sourceCode := `/**
 * @package api
 * @summary Create Customer
 * @description REST endpoint
 * @tag Customers
 * @tag API
 */
Parm();`

	doc, err := Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(doc.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(doc.Tags))
	}

	if doc.Tags[0] != "Customers" {
		t.Errorf("Expected first tag 'Customers', got '%s'", doc.Tags[0])
	}

	if doc.Tags[1] != "API" {
		t.Errorf("Expected second tag 'API', got '%s'", doc.Tags[1])
	}
}

func TestParseParameter_INOUT(t *testing.T) {
	param := parseParameter("OrderData INOUT sdtOrder - Order information to be processed")

	if param == nil {
		t.Fatal("Expected param to be non-nil")
	}

	if param.Name != "OrderData" {
		t.Errorf("Expected name 'OrderData', got '%s'", param.Name)
	}

	if param.Direction != "INOUT" {
		t.Errorf("Expected direction 'INOUT', got '%s'", param.Direction)
	}

	if param.Type != "sdtOrder" {
		t.Errorf("Expected type 'sdtOrder', got '%s'", param.Type)
	}

	if param.Description != "Order information to be processed" {
		t.Errorf("Expected description, got '%s'", param.Description)
	}
}

func TestParseParameter_NoDirection(t *testing.T) {
	param := parseParameter("CustomerID Numeric - Customer identifier")

	if param == nil {
		t.Fatal("Expected param to be non-nil")
	}

	// Should default to IN
	if param.Direction != "IN" {
		t.Errorf("Expected default direction 'IN', got '%s'", param.Direction)
	}

	if param.Type != "Numeric" {
		t.Errorf("Expected type 'Numeric', got '%s'", param.Type)
	}
}

func TestParseParameter_NoDescription(t *testing.T) {
	param := parseParameter("Status OUT Boolean")

	if param == nil {
		t.Fatal("Expected param to be non-nil")
	}

	if param.Description != "" {
		t.Errorf("Expected empty description, got '%s'", param.Description)
	}
}

func TestExtractCommentBlock(t *testing.T) {
	source := `/**
 * @package test
 * @summary Test Summary
 */
Some code here`

	block := extractCommentBlock(source)
	
	if block == "" {
		t.Fatal("Expected non-empty comment block")
	}

	if !containsString(block, "@package test") {
		t.Errorf("Expected comment block to contain '@package test', got: %s", block)
	}
}

func TestExtractCommentBlock_NoComment(t *testing.T) {
	source := `Just some code without comments`

	block := extractCommentBlock(source)
	
	if block != "" {
		t.Errorf("Expected empty comment block, got: %s", block)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && (s[len(s)-len(substr):] == substr || 
		findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
