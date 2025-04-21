package ai

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockGenerateCommitMessage(t *testing.T) {
	mockAI := &MockAI{}
	diff := "some changes"
	branchName := "100"
	expected := "Mock Commit Message for " + diff + " and branch " + branchName
	msg, err := mockAI.CommitMessage(branchName, diff)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if msg != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, msg)
	}
}

func TestMockGenerateLabels(t *testing.T) {
	mockAI := &MockAI{}
	labels := []string{"bug", "feature"}
	actual, err := mockAI.IssueLabels("issue", labels)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	assert.Equal(t, labels, actual)
}

func TestMockGenerateTitle(t *testing.T) {
	mockAI := &MockAI{}
	branchName := "feature-branch"
	expected := "'Mock Title for " + branchName + "'"

	title, err := mockAI.PrTitle(branchName, "diff", "issue", "summary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if title != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, title)
	}
}

func TestMockGenerateIssueTitle(t *testing.T) {
	mockAI := &MockAI{}
	userInput := "issue input"
	expected := "'Mock Issue Title for " + userInput + "'"

	title, err := mockAI.IssueTitle(userInput, "summary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if title != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, title)
	}
}

func TestMockGenerateIssueBody(t *testing.T) {
	mockAI := &MockAI{}
	userInput := "issue input"
	expected := "Mock Issue Body for " + userInput

	body, err := mockAI.IssueBody(userInput, "summary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if body != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, body)
	}
}

func TestMockGenerateBody(t *testing.T) {
	mockAI := &MockAI{}
	branchName := "feature-branch"
	expected := "Mock Body for " + branchName

	body, err := mockAI.PrBody(branchName, "diff", "issue", "summary")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if body != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, body)
	}
}
