package ai

type MockAI struct{}

func (m *MockAI) GenerateTitle(branchName string) (string, error) {
    return "Mock Title for " + branchName, nil
}

func (m *MockAI) GenerateBody(branchName string) (string, error) {
    return "Mock Body for " + branchName, nil
}
