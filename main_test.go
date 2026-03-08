package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	// Setup: expected output filename based on db.xml content
	expectedOutput := "character_id-00001_4.xml"
	// Ensure cleanup
	defer os.Remove(expectedOutput)

	// Run the extraction using the sample db.xml
	if err := run("tests/db.xml"); err != nil {
		t.Fatalf("run() failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		t.Fatalf("Expected output file %s was not created", expectedOutput)
	}

	// Compare generated content with golden file
	goldenFile := filepath.Join("tests", "expected_"+expectedOutput)
	expectedContent, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", goldenFile, err)
	}

	generatedContent, err := os.ReadFile(expectedOutput)
	if err != nil {
		t.Fatalf("Failed to read generated file %s: %v", expectedOutput, err)
	}

	if !bytes.Equal(expectedContent, generatedContent) {
		t.Errorf("Generated content does not match golden file.\nExpected len: %d\nGot len: %d", len(expectedContent), len(generatedContent))
		// Optional: Write diff or print snippet for debugging
		// For large files, printing full diff is too much.
	}
}
