package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func (repo *GitRepository) RepoPath(path ...string) string {
	return filepath.Join(append([]string{repo.gitdir}, path...)...)
}

// RepoFile returns the path to a file in the repository
func (repo *GitRepository) RepoFile(mkdir bool, path ...string) string {
	dirPath := filepath.Join(path[:len(path)-1]...)
	if con, err := repo.repoDir(mkdir, dirPath); con && err == nil {
		return repo.RepoPath(filepath.Join(path...))
	}
	return ""
}

func (repo *GitRepository) repoDir(mkdir bool, path ...string) (bool, error) {
	repPath := repo.RepoPath(path...)

	if info, err := os.Stat(repPath); err == nil {
		if info.IsDir() {
			return true, nil
		} else {
			return false, fmt.Errorf("not a valid directory %s", repPath)
		}
	}

	if mkdir {
		if err := os.MkdirAll(repPath, os.ModePerm); err != nil {
			return false, fmt.Errorf("could not create directory %s", repPath)
		}
		fmt.Printf("created directory %s\n", repPath)
		return true, nil
	} else {
		return false, nil
	}
}

func RepoCreate(path string) (*GitRepository, error) {
	log.Printf("initializing repository step1: %s", path) //debug
	repo, err := NewGitRepository(path, false)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, fmt.Errorf("failed to create repository: %v", err)
	}

	if info, err := os.Stat(repo.worktree); err == nil {
		if !info.IsDir() {
			log.Fatalf("%s is not a directory", repo.worktree)

		}
		if dirinfo, err := os.Stat(repo.gitdir); err == nil && dirinfo.IsDir() {
			if entries, err := os.ReadDir(repo.gitdir); err == nil && len(entries) > 0 {
				log.Fatalf("%s is not an empty directory", repo.gitdir)
			}
		}
	} else {
		if err := os.MkdirAll(repo.worktree, os.ModePerm); err != nil {
			return nil, fmt.Errorf("could not create directory %s because: %v", repo.worktree, err)
		}
	}
	if _, err := repo.repoDir(true, "branches"); err != nil {
		return nil, fmt.Errorf("failed to create 'branches' directory: %v", err)
	}
	if _, err := repo.repoDir(true, "objects"); err != nil {
		return nil, fmt.Errorf("failed to create 'objects' directory: %v", err)
	}
	if _, err := repo.repoDir(true, "refs", "tags"); err != nil {
		return nil, fmt.Errorf("failed to create 'refs/tags' directory: %v", err)
	}
	if _, err := repo.repoDir(true, "refs", "heads"); err != nil {
		return nil, fmt.Errorf("failed to create 'refs/heads' directory: %v", err)
	}

	configPath := repo.RepoFile(true, "config")
	repoConfig := repoDefaultConfig()
	configFile, err := os.Create(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file: %v", err)
	}
	defer configFile.Close()
	if _, err := repoConfig.WriteTo(configFile); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}
	log.Printf("initializing repository step_final: %s", path) //debug

	return repo, nil
}

func repoDefaultConfig() *ini.File {
	cfg := ini.Empty()
	coreSection, err := cfg.NewSection("core")
	if err != nil {
		log.Fatalf("failed to create core section in ini file: %v", err)
	}
	coreSection.NewKey("repositoryformatversion", "0")
	coreSection.NewKey("filemode", "false")
	coreSection.NewKey("bare", "false")
	return cfg
}

func RepoFind(path string, required bool) (*GitRepository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	gitPath := filepath.Join(absPath, ".git")
	if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
		return NewGitRepository(absPath, false)
	}

	parentPath := filepath.Dir(absPath)
	if parentPath == absPath {
		if required {
			return nil, fmt.Errorf("no git directory found")
		}
		return nil, nil
	}

	return RepoFind(parentPath, required)
}
