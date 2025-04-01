package git

import ("testing")

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

func TestMockGetBranchName(t *testing.T) {
    git := MockGit{}
    output, err := git.GetBranchName()
    if err != nil {
        panic(err)
    }
    if output != "main" {
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
