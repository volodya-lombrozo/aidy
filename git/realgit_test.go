package git                                                                                                                                                                                     

import (                                                                                                                                                                                             
    "os"                                                                                                                                                                                             
    "os/exec"                             
    "testing"                                                                                                                                                                                        
    "path/filepath"                                                                                                                                                                                  
)                                                                                                                                                                                                    

func setupTestRepo(t *testing.T) (string, func()) {                                                                                                                                                  
    tempDir, err := os.MkdirTemp("", "testrepo")                                                                                                                                                     
    if err != nil {                                                                                                                                                                                  
        t.Fatalf("Failed to create temp dir: %v", err)                                                                                                                                               
    }                                                                                                                                                                                                
    cmd := exec.Command("git", "init")                                                                                                                                                               
    cmd.Dir = tempDir                                                                                                                                                                                
    if err := cmd.Run(); err != nil {                                                                                                                                                                
        t.Fatalf("Failed to initialize git repo: %v", err)                                                                                                                                           
    } 
    cmd = exec.Command("git", "commit", "-m", "initial commit", "--allow-empty")                                                                                                                                              
    cmd.Dir = tempDir                                                                                                                                                                                
    if err := cmd.Run(); err != nil {                                                                                                                                                                
        t.Fatalf("Failed to make an initial commit: %v", err)                                                                                                                                            
    } 
    cmd = exec.Command("git", "checkout", "-b", "main-branch")                                                                                                                                              
    cmd.Dir = tempDir                                                                                                                                                                                
    if err := cmd.Run(); err != nil {                                                                                                                                                                
        t.Fatalf("Failed to create 'main-branch' branch: %v", err)                                                                                                                                            
    }
    cmd = exec.Command("git", "commit", "-m", "second commit", "--allow-empty")                                                                                                                                              
    cmd.Dir = tempDir                                                                                                                                                                                
    if err := cmd.Run(); err != nil {                                                                                                                                                                
        t.Fatalf("Failed to make an initial commit: %v", err)                                                                                                                                            
    }                                                                                                                                                                                                 
    return tempDir, func() {                                                                                                                                                                         
        os.RemoveAll(tempDir)                                                                                                                                                                        
    }                                                                                                                                                                                                
}                                                                                                                                                                                                    

func TestRealGetBranchName(t *testing.T) {                                                                                                                                                               
    dir, cleanup := setupTestRepo(t)                                                                                                                                                             
    defer cleanup()                                                                                                                                                                                  
    gitService := NewRealGit(dir)
    branchName, err := gitService.GetBranchName()                                                                                                                                                    
    if err != nil {                                                                                                                                                                                  
        t.Fatalf("Error getting branch name: %v", err)
    }                                                                                                                                                                                                
    if branchName != "main-branch" {                                                                                                                                                                        
        t.Fatalf("Expected branch name 'main-branch', got '%s'", branchName)                                                                                                                                
    }                                                                                                                                                                                                
}                                                                                                                                                                                                    

func TestRealGetBaseBranchName(t *testing.T) {                                                                                                                                                           
    repoDir, cleanup := setupTestRepo(t)                                                                                                                                                             
    defer cleanup()                                                                                                                                                                                  
    gitService := NewRealGit(repoDir)
    baseBranch, err := gitService.GetBaseBranchName()                                                                                                                                                
    if err != nil {                                                                                                                                                                                  
        t.Fatalf("Error getting base branch name: %v", err)                                                                                                                                          
    }                                                                                                                                                                                                
    if baseBranch != "main" {                                                                                                                                                                        
        t.Fatalf("Expected base branch name 'main', got '%s'", baseBranch)                                                                                                                           
    }                                                                                                                                                                                                
}                                                                                                                                                                                                    

func TestRealGetDiff(t *testing.T) {                                                                                                                                                                     
    repoDir, cleanup := setupTestRepo(t)                                                                                                                                                             
    defer cleanup()                                                                                                                                                                                  
    filePath := filepath.Join(repoDir, "file.txt")                                                                                                                                                   
    os.WriteFile(filePath, []byte("Hello, World!"), 0644)                                                                                                                                            
    cmd := exec.Command("git", "add", ".")                                                                                                                                                           
    cmd.Dir = repoDir                                                                                                                                                                                
    cmd.Run()                                                                                                                                                                                        
    cmd = exec.Command("git", "commit", "-m", "Add hello world")                                                                                                                                      
    cmd.Dir = repoDir                                                                                                                                                                                
    cmd.Run()                                                                                                                                                                                        
    os.WriteFile(filePath, []byte("Hello, Git!"), 0644)                                                                                                                                              
    gitService := NewRealGit(repoDir)
    diff, err := gitService.GetDiff()                                                                                                                                                                
    if err != nil {                                                                                                                                                                                  
        t.Fatalf("Error getting diff: %v", err)                                                                                                                                                      
    }                                                                                                                                                                                                
    if diff == "" {                                                                                                                                                                                  
        t.Fatal("Expected non-empty diff")                                                                                                                                                           
    }                                                                                                                                                                                                
}                 
