// Package profiles provides a streaming, line-by-line parser for agent profile
// Markdown files that include YAML frontmatter and named body sections.
package profiles

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	maxProfileBytes = 1024 * 1024 // 1 MB
	scanBufInit     = 64 * 1024   // 64 KB initial scanner buffer
	scanBufMax      = 2 * 1024 * 1024 // 2 MB maximum line size
)

// ResolveAgentPath constructs the path for agentName inside profilesDir and
// returns an error if the resolved path would escape profilesDir.
//
// The guard is two-layered:
//  1. Lexical: filepath.Rel-based containment against the absolute profilesDir.
//  2. Symbolic: if the candidate file exists, filepath.EvalSymlinks is applied
//     to both profilesDir and the candidate, and containment is re-checked.
//     This rejects a symlink inside profilesDir that points outside it.
//
// If the candidate does not yet exist (which is common — ResolveAgentPath is
// also called as a pre-read validation step), only the lexical check runs.
// Callers that read the returned path MUST treat the returned path as
// untrusted-until-read if they also accept input that may name non-existent
// profiles; use EnsureContained on any path they compute via alternate means
// (e.g., case-insensitive lookup) before opening it.
func ResolveAgentPath(profilesDir, agentName string) (string, error) {
	if agentName == "" {
		return "", fmt.Errorf("agent name is empty")
	}
	if strings.ContainsAny(agentName, "/\\") {
		return "", fmt.Errorf("agent name %q resolves outside profiles directory", agentName)
	}
	absDir, err := filepath.Abs(profilesDir)
	if err != nil {
		return "", fmt.Errorf("resolving profiles directory: %w", err)
	}

	candidate := filepath.Join(absDir, agentName+".md")
	absCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolving candidate path: %w", err)
	}

	if !lexicallyContained(absDir, absCandidate) {
		return "", fmt.Errorf("agent name %q resolves outside profiles directory", agentName)
	}

	// Symbolic containment: only possible if the file exists. If it does not
	// exist yet, the lexical guard above is sufficient — no symlink to follow.
	if _, statErr := os.Lstat(absCandidate); statErr == nil {
		resolved, err := EnsureContained(absDir, absCandidate)
		if err != nil {
			return "", err
		}
		return resolved, nil
	}

	return absCandidate, nil
}

// EnsureContained resolves base and candidate via filepath.EvalSymlinks and
// confirms the resolved candidate is lexically contained within the resolved
// base. Returns the resolved candidate path on success. Exported so callers
// that compute profile paths by alternate means (e.g., case-insensitive
// lookup) can re-validate before opening the file.
func EnsureContained(base, candidate string) (string, error) {
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("resolve symlinks on base %q: %w", base, err)
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve symlinks on candidate %q: %w", candidate, err)
	}
	absBase, err := filepath.Abs(resolvedBase)
	if err != nil {
		return "", err
	}
	absCandidate, err := filepath.Abs(resolvedCandidate)
	if err != nil {
		return "", err
	}
	if !lexicallyContained(absBase, absCandidate) {
		return "", fmt.Errorf("path %q resolves outside %q after symlink evaluation", candidate, base)
	}
	return absCandidate, nil
}

func lexicallyContained(base, candidate string) bool {
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

// ParseProfile reads the profile at path and returns:
//   - fm: the parsed YAML frontmatter as a map
//   - sections: the requested body sections (heading title → content)
//   - error on schema violation, oversized file, or I/O failure
//
// If sections is nil or empty, body parsing is skipped entirely.
// The file size is checked before opening; files larger than 1 MB are rejected.
func ParseProfile(path string, sections []string) (map[string]any, map[string]string, error) {
	// Size guard — do this before opening the file.
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Size() > maxProfileBytes {
		return nil, nil, fmt.Errorf("profile %s is too large (%d bytes)", path, info.Size())
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Use a 64 KB token buffer, with a 2 MB max line size for safety.
	scanner.Buffer(make([]byte, scanBufInit), scanBufMax)

	// --- Frontmatter parsing ------------------------------------------------
	// Read the first non-empty line; it must be exactly "---".
	var firstLine string
	for scanner.Scan() {
		firstLine = strings.TrimSpace(scanner.Text())
		if firstLine != "" {
			break
		}
	}
	if firstLine != "---" {
		// No frontmatter — return empty maps, no error.
		return map[string]any{}, map[string]string{}, nil
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
		return nil, nil, fmt.Errorf("unterminated frontmatter in %s", filepath.Base(path))
	}

	rawFM := strings.Join(fmLines, "\n")
	fm := make(map[string]any)
	if err := yaml.Unmarshal([]byte(rawFM), &fm); err != nil {
		return nil, nil, fmt.Errorf("parsing frontmatter in %s: %w", path, err)
	}

	// --- Schema validation --------------------------------------------------
	if err := validateSchema(fm); err != nil {
		return nil, nil, err
	}

	// --- Body section parsing -----------------------------------------------
	if len(sections) == 0 {
		return fm, map[string]string{}, nil
	}

	// Build a lookup set for the requested section names.
	wanted := make(map[string]bool, len(sections))
	for _, s := range sections {
		wanted[s] = true
	}

	collected := make(map[string]string)
	var (
		currentSection string
		fenceKind      string // "" when not in fence, "```" or "~~~" when in fence
		buf            strings.Builder
	)

	flushSection := func() {
		if currentSection != "" {
			collected[currentSection] = strings.TrimSpace(buf.String())
		}
		buf.Reset()
	}

	allCollected := func() bool {
		if len(collected) < len(sections) {
			return false
		}
		for _, s := range sections {
			if _, ok := collected[s]; !ok {
				return false
			}
		}
		return true
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Toggle fenced code block state, tracking which delimiter opened it.
		if fenceKind == "" {
			// Not currently in a fence — check if this line opens one.
			if strings.HasPrefix(line, "```") {
				fenceKind = "```"
				if currentSection != "" {
					buf.WriteString(line)
					buf.WriteByte('\n')
				}
				continue
			}
			if strings.HasPrefix(line, "~~~") {
				fenceKind = "~~~"
				if currentSection != "" {
					buf.WriteString(line)
					buf.WriteByte('\n')
				}
				continue
			}
		} else {
			// In a fence — only the matching delimiter closes it.
			if strings.HasPrefix(line, fenceKind) {
				if currentSection != "" {
					buf.WriteString(line)
					buf.WriteByte('\n')
				}
				fenceKind = ""
				continue
			}
			// Inside a fence: accumulate and skip heading detection.
			if currentSection != "" {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
			continue
		}

		// Detect ## headings only outside fenced blocks.
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")

			// Flush whatever we were accumulating for the previous section.
			flushSection()
			currentSection = ""

			// Early exit: if all wanted sections are already collected and we
			// hit a new heading that is NOT in the wanted list, we are done.
			if allCollected() {
				break
			}

			if wanted[heading] {
				currentSection = heading
			}
			continue
		}

		// Accumulate body lines for the current wanted section.
		if currentSection != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	// Flush whatever was in flight when the file ended.
	flushSection()

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return fm, collected, nil
}

// validateSchema checks that fm contains name, xp, and role/roles.
// Returns an error listing all missing fields.
func validateSchema(fm map[string]any) error {
	var missing []string

	if _, ok := fm["name"]; !ok {
		missing = append(missing, "name")
	}
	if _, ok := fm["xp"]; !ok {
		missing = append(missing, "xp")
	}
	_, hasRole := fm["role"]
	_, hasRoles := fm["roles"]
	if !hasRole && !hasRoles {
		missing = append(missing, "role")
	}

	if len(missing) > 0 {
		return fmt.Errorf("profile missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// StreamProfiles returns a sorted list of .md file paths found directly inside
// dir. If dir does not exist, returns nil, nil (not an error).
func StreamProfiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading profiles dir %s: %w", dir, err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(paths)
	return paths, nil
}
