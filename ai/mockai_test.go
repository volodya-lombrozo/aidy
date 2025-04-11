package ai

import (
	"testing"
"github.com/stretchr/testify/assert"
)

func TestMockGenerateCommitMessage(t *testing.T) {
	mockAI := &MockAI{}
	diff := "some changes"
	branchName := "100"
	expected := "Mock Commit Message for " + diff + " and branch " + branchName
	msg, err := mockAI.GenerateCommitMessage(branchName, diff)
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
	actual, err := mockAI.GenerateIssueLabels("issue", labels)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
    assert.Equal(t, labels, actual)
}


func TestMockGenerateTitle(t *testing.T) {
	mockAI := &MockAI{}
	branchName := "feature-branch"
	expected := "Mock Title for " + branchName

	title, err := mockAI.GenerateTitle(branchName, "diff", "issue")
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
	expected := "Mock Issue Title for " + userInput

	title, err := mockAI.GenerateIssueTitle(userInput)
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

	body, err := mockAI.GenerateIssueBody(userInput)
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

	body, err := mockAI.GenerateBody(branchName, "diff", "issue")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if body != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, body)
	}
}
