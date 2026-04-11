package profiles

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// UpdateFrontmatterField opens the Markdown file at path, updates the YAML
// frontmatter key to value, and writes the result back atomically using a
// temp file in the same directory (so the rename stays on the same filesystem).
//
// The body is preserved byte-for-byte. Frontmatter delimiters are found
// line-by-line so YAML values that happen to contain "---" do not corrupt
// parsing.
func UpdateFrontmatterField(path string, key string, value any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, scanBufInit), scanBufMax)

	// First line must be exactly "---".
	if !scanner.Scan() {
		return fmt.Errorf("no frontmatter found in %s", path)
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return fmt.Errorf("no frontmatter found in %s", path)
	}

	// Accumulate frontmatter lines until the closing "---".
	var fmLines []string
	closingSeen := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			closingSeen = true
			break
		}
		fmLines = append(fmLines, line)
	}
	if !closingSeen {
		return fmt.Errorf("unterminated frontmatter in %s", filepath.Base(path))
	}

	// Accumulate remaining body lines, preserving each line verbatim.
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	// Close the source file before we overwrite it.
	f.Close()

	// Parse, update, re-marshal frontmatter.
	rawFM := strings.Join(fmLines, "\n")
	fm := make(map[string]any)
	if err := yaml.Unmarshal([]byte(rawFM), &fm); err != nil {
		return fmt.Errorf("parsing frontmatter in %s: %w", path, err)
	}
	fm[key] = value

	marshaled, err := yaml.Marshal(fm)
	if err != nil {
		return fmt.Errorf("marshaling frontmatter: %w", err)
	}

	// Write to a temp file in the same directory for atomic rename.
	tmp, err := os.CreateTemp(filepath.Dir(path), "armies-fm-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Helper to clean up the temp file on any write error.
	writeErr := func(e error) error {
		tmp.Close()
		os.Remove(tmpName)
		return e
	}

	// Write: ---\n<marshaled yaml>---\n<body>
	if _, err := fmt.Fprintf(tmp, "---\n%s---\n", marshaled); err != nil {
		return writeErr(fmt.Errorf("writing frontmatter to temp: %w", err))
	}

	// Re-emit body lines, preserving trailing newline behaviour.
	// bodyLines has each line without its newline; rejoin with \n and add a
	// final \n if there were any lines (matching the original file).
	if len(bodyLines) > 0 {
		body := strings.Join(bodyLines, "\n") + "\n"
		if _, err := tmp.WriteString(body); err != nil {
			return writeErr(fmt.Errorf("writing body to temp: %w", err))
		}
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}

	return nil
}
