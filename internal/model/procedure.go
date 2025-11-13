package model

// Procedure represents a GeneXus Procedure object
type Procedure struct {
	// Name is the procedure's identifier
	Name string

	// Description is the human-readable description
	Description string

	// Path is the relative file path within the XPZ archive
	Path string

	// SourceCode contains the extracted source code
	SourceCode string

	// Documentation contains parsed annotation comments
	Documentation *DocComment
}
