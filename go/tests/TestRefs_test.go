package tests

import (
	"reflect"
	"testing"

	"github.com/saichler/l8web/go/web/webhook"
)

func TestExtractIssueRefs_Fixes(t *testing.T) {
	refs := webhook.ExtractIssueRefs("fixes #42")
	expected := []string{"42"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_Closes(t *testing.T) {
	refs := webhook.ExtractIssueRefs("Closes FEAT-123")
	expected := []string{"FEAT-123"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_UUID(t *testing.T) {
	refs := webhook.ExtractIssueRefs("resolved 550e8400-e29b-41d4-a716-446655440000")
	expected := []string{"550e8400-e29b-41d4-a716-446655440000"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_Multiple(t *testing.T) {
	text := "fixes #1, closes BUG-2, resolves 550e8400-e29b-41d4-a716-446655440000"
	refs := webhook.ExtractIssueRefs(text)
	expected := []string{"1", "BUG-2", "550e8400-e29b-41d4-a716-446655440000"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_Dedup(t *testing.T) {
	refs := webhook.ExtractIssueRefs("fixes #42 and also closes #42")
	expected := []string{"42"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_NoMatch(t *testing.T) {
	refs := webhook.ExtractIssueRefs("added logging and improved performance")
	if refs != nil {
		t.Fatalf("expected nil, got %v", refs)
	}
}

func TestExtractIssueRefs_CaseInsensitive(t *testing.T) {
	refs := webhook.ExtractIssueRefs("FIXES #1")
	expected := []string{"1"}
	if !reflect.DeepEqual(refs, expected) {
		t.Fatalf("expected %v, got %v", expected, refs)
	}
}

func TestExtractIssueRefs_VerbVariants(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"fix #1", []string{"1"}},
		{"fixes #2", []string{"2"}},
		{"fixed #3", []string{"3"}},
		{"close #4", []string{"4"}},
		{"closes #5", []string{"5"}},
		{"closed #6", []string{"6"}},
		{"resolve #7", []string{"7"}},
		{"resolves #8", []string{"8"}},
		{"resolved #9", []string{"9"}},
	}
	for _, tt := range tests {
		refs := webhook.ExtractIssueRefs(tt.input)
		if !reflect.DeepEqual(refs, tt.expected) {
			t.Errorf("ExtractIssueRefs(%q) = %v, want %v", tt.input, refs, tt.expected)
		}
	}
}
