package ai

type MockAI struct{}

func (m *MockAI) GenerateTitle(branchName string, diff string) (string, error) {
    return "Mock Title for " + branchName, nil
}

func (m *MockAI) GenerateIssueTitle(userInput string) (string, error) {
    return "Mock Issue Title for " + userInput, nil
}

func (m *MockAI) GenerateIssueBody(userInput string) (string, error) {
    return "Mock Issue Body for " + userInput, nil
}

func (m *MockAI) GenerateBody(branchName string, diff string) (string, error) {
    return "Mock Body for " + branchName, nil
}
