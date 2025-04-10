package git

import (
	"testing"
)

func TestMockGetBaseBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBaseBranchName()
	if err != nil {
		panic(err)
	}
	if output != "main" {
		t.Fatal("Expected the branch name 'main'")
	}
}

func TestMockGetCurrentCommitMessage(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentCommitMessage()
	if err != nil {
		panic(err)
	}
	expected := "feat(#42): current commit message"
	if output != expected {
		t.Fatalf("Expected the commit message '%v', but got '%v'", expected, output)
	}
}

func TestMockGetBranchName(t *testing.T) {
	git := MockGit{}
	output, err := git.GetBranchName()
	if err != nil {
		panic(err)
	}
	if output != "41_working_branch" {
		t.Fatal("Expected the branch name 'main'")
	}
}

func TestMockGetDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetDiff()
	if err != nil {
		panic(err)
	}
	if output != "mock-diff" {
		t.Fatal("Expected the diff 'mock-diff'")
	}
}

func TestMockGetCurrentDiff(t *testing.T) {
	git := MockGit{}
	output, err := git.GetCurrentDiff()
	if err != nil {
		panic(err)
	}
	if output != "current-mock-diff" {
		t.Fatal("Expected the diff 'mock-diff'")
	}
}

func TestMockGetAllRemoteURLs(t *testing.T) {
	git := MockGit{}
	output, err := git.GetAllRemoteURLs()
	if err != nil {
		panic(err)
	}
	first := "https://github.com/volodya-lombrozo/aidy.git"
	second := "https://github.com/volodya-lombrozo/forked-aidy.git"
	if output[0] != first {
		t.Fatalf("Expected the diff '%s'", first)
	}
	if output[1] != second {
		t.Fatalf("Expected the diff '%s'", second)
	}
}
