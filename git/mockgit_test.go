package git

import (
    "testing"
    "fmt"
)

func TestMockGetBaseBranchName(t *testing.T){
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
        t.Fatal(fmt.Sprintf("Expected the commit message '%v', but got '%v'", expected, output))
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
