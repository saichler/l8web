package webhook

import (
	"regexp"
	"strings"
)

// issueRefPattern matches common VCS issue reference patterns:
//   - "Fixes #42", "Closes #42", "Resolves #42"
//   - "Fixes L8B-42", "Closes FEAT-99"
//   - "Fixes <uuid>"
var issueRefPattern = regexp.MustCompile(
	`(?i)(?:fix(?:es|ed)?|close[sd]?|resolve[sd]?)\s+` +
		`(#\d+|[A-Za-z]+-\d+|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

// ExtractIssueRefs finds issue references in text using common VCS patterns.
// Matches: "fixes #42", "closes L8B-123", "resolved <uuid>".
// Returns deduplicated refs with # prefix stripped.
func ExtractIssueRefs(text string) []string {
	matches := issueRefPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var refs []string
	for _, m := range matches {
		ref := strings.TrimPrefix(m[1], "#")
		if !seen[ref] {
			seen[ref] = true
			refs = append(refs, ref)
		}
	}
	return refs
}
