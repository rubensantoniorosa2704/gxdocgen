package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rubensantoniorosa2704/gxdocgen/internal/generator"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/utils"
	"github.com/rubensantoniorosa2704/gxdocgen/internal/xpz"
)

const (
	version = "0.2.0"
)

func main() {
	// Define command-line flags
	var (
		inputPath  string
		outputPath string
		showHelp   bool
		showVer    bool
	)

	flag.StringVar(&inputPath, "input", "", "Path to the GeneXus XPZ file (required)")
	flag.StringVar(&outputPath, "output", "./docs", "Output directory for generated documentation")
	flag.BoolVar(&showHelp, "help", false, "Show usage information")
	flag.BoolVar(&showHelp, "h", false, "Show usage information (shorthand)")
	flag.BoolVar(&showVer, "version", false, "Show version information")
	flag.BoolVar(&showVer, "v", false, "Show version information (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	// Handle version flag
	if showVer {
		fmt.Printf("GXDocGen version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if showHelp {
		printUsage()
		os.Exit(0)
	}

	// Validate required input flag
	if inputPath == "" {
		utils.Error("Missing required flag: --input")
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	// Validate input file exists and has .xpz extension
	if err := validateInput(inputPath); err != nil {
		utils.Fatal("Invalid input: %v", err)
	}

	// Print banner
	printBanner()

	// Step 1: Extract XPZ file
	utils.Info("Step 1/2: Extracting XPZ file...")
	result, err := xpz.Extract(inputPath)
	if err != nil {
		utils.Fatal("Failed to extract XPZ: %v", err)
	}

	// Step 2: Generate documentation
	utils.Info("Step 2/2: Generating documentation...")
	if err := generator.GenerateDocs(result.Objects, result.KBName, outputPath); err != nil {
		utils.Fatal("Failed to generate documentation: %v", err)
	}

	// Success message
	fmt.Println()
	utils.Success("Documentation generation complete!")
	utils.Info("Output location: %s", outputPath)
}

// validateInput checks if the input file exists and has proper extension
func validateInput(path string) error {
	// Check if file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("expected a file, got a directory: %s", path)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".xpz" {
		return fmt.Errorf("expected .xpz file, got: %s", ext)
	}

	return nil
}

// printBanner prints the application banner
func printBanner() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════╗")
	fmt.Println("║         GXDocGen v" + version + "               ║")
	fmt.Println("║  GeneXus Documentation Generator      ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println()
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("GXDocGen - GeneXus Documentation Generator")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s --input <xpz-file> [options]\n", os.Args[0])
	fmt.Println()
	fmt.Println("REQUIRED FLAGS:")
	fmt.Println("  --input <path>       Path to the GeneXus XPZ file")
	fmt.Println()
	fmt.Println("OPTIONAL FLAGS:")
	fmt.Println("  --output <path>      Output directory (default: ./docs)")
	fmt.Println("  --help, -h           Show this help message")
	fmt.Println("  --version, -v        Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Printf("  %s --input ./export.xpz\n", os.Args[0])
	fmt.Printf("  %s --input ./export.xpz --output ./documentation\n", os.Args[0])
	fmt.Println()
}
