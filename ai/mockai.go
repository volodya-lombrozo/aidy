package ai

type MockAI struct{}

func (m *MockAI) PrTitle(branchName string, diff string, issue string, summary string) (string, error) {
	return "'Mock Title for " + branchName + "'", nil
}

func (m *MockAI) PrBody(branchName string, diff string, issue string, summary string) (string, error) {
	return "Mock Body for " + branchName, nil
}

func (m *MockAI) IssueTitle(userInput string, summary string) (string, error) {
	return "'Mock Issue Title for " + userInput + "'", nil
}

func (m *MockAI) IssueBody(userInput string, summary string) (string, error) {
	return "Mock Issue Body for " + userInput, nil
}

func (m *MockAI) CommitMessage(branchName string, diff string) (string, error) {
	return "Mock Commit Message for " + diff + " and branch " + branchName, nil
}

func (m *MockAI) IssueLabels(issue string, available []string) ([]string, error) {
	return available, nil
}

func (m *MockAI) Summary(readme string) (string, error) {
	return "summary: " + readme, nil
}

func (m *MockAI) SuggestBranch(descr string) (string, error) {
	return "mock-branch-name", nil
}
