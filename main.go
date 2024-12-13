package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

// type GitRepository interface {
// }
type GitRepository struct {
	worktree string
	gitdir   string
	config   *ini.File
}

func NewGitRepository(path string, force bool) (*GitRepository, error) {
	repo := &GitRepository{
		worktree: path,
		gitdir:   filepath.Join(path, ".git"),
	}
	if !force {
		stat, err := os.Stat(repo.gitdir)
		if err != nil {
			log.Printf("error : %v", err) // debug
		} else if !stat.IsDir() {
			return nil, fmt.Errorf("not a Git repository %s", path)
		}

	}
	// read config file
	confPath := filepath.Join(repo.gitdir, "config")
	if _, err := os.Stat(confPath); err == nil {
		// fmt.Errorf("error at config: %v", err)
		repo.config, err = ini.Load(confPath)
		if err != nil {
			return nil, fmt.Errorf("could not load configuration file %s, error:\n %v", confPath, err)
		}
		if !force {
			vers, err := strconv.Atoi(repo.config.Section("core").Key("repositoryformatversion").String())
			if err != nil {
				return nil, err
			}
			if vers != 0 {
				return nil, fmt.Errorf("unsupported repositoryformatversion %d", vers)
			}
		}
	} /* else if !force {
		return nil, fmt.Errorf("configuration file missing")
	 }*/

	return repo, nil
}

func main() {
	InitCmd := flag.NewFlagSet("init", flag.ExitOnError)
	path := InitCmd.String("path", ".", "path to the repository")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("expected 'init' subcommand")
		os.Exit(1)
	}
	switch flag.Args()[0] {
	case "init":
		// log.Fatalf("init %s", flag.Args()[1:])
		InitCmd.Parse(flag.Args()[1:])
		CmdInit(*path)

	}
}

func CmdInit(args ...string) {
	// log.Fatalf("init test: %s", args[0])
	RepoCreate(args[0])
}
