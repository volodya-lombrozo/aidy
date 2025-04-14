package ai

type MockAI struct{}

func (m *MockAI) GenerateTitle(branchName string, diff string, issue string) (string, error) {
	return "'Mock Title for " + branchName + "'", nil
}

func (m *MockAI) GenerateBody(branchName string, diff string, issue string) (string, error) {
	return "Mock Body for " + branchName, nil
}

func (m *MockAI) GenerateIssueTitle(userInput string) (string, error) {
	return "'Mock Issue Title for " + userInput + "'", nil
}

func (m *MockAI) GenerateIssueBody(userInput string) (string, error) {
	return "Mock Issue Body for " + userInput, nil
}

func (m *MockAI) GenerateCommitMessage(branchName string, diff string) (string, error) {
	return "Mock Commit Message for " + diff + " and branch " + branchName, nil
}

func (m *MockAI) GenerateIssueLabels(issue string, available []string) ([]string, error) {
	return available, nil
}
