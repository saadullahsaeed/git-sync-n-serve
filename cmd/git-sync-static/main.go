package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	git "gopkg.in/src-d/go-git.v4"
)

const (
	defaultTimeout = 30

	defaultWait = 60
)

func main() {
	flag.String("repo", "", "Repository URL")
	flag.String("dest", "/tmp/git", "Destination directory")
	flag.String("branch", "master", "Branch to sync")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	repo := viper.GetString("repo")
	dest := viper.GetString("dest")

	errChan := make(chan error, 1)

	parts := strings.Split(repo, "/")
	if len(parts) == 0 {
		os.Exit(-1)
	}

	destPath := fmt.Sprintf("%s/%s", dest, parts[len(parts)-1])
	gs := &GitSync{RepositoryURL: repo, Destination: dest, Path: destPath}
	go func() {
		errChan <- gs.Start()
	}()

	//start the server
	go func(destPath string) {
		fs := http.FileServer(http.Dir(destPath))
		http.Handle("/", fs)

		log.Println("Listening...")
		http.ListenAndServe(":3000", nil)
	}(destPath)

	<-errChan
}

// GitSync struct represents something
type GitSync struct {
	RepositoryURL string
	Destination   string
	Path          string
	Branch        string
}

// Start ...
func (gs *GitSync) Start() error {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(defaultTimeout))

		_, err := os.Stat(gs.Path)
		switch {
		case os.IsNotExist(err):
			err := cloneRepo(ctx, gs.RepositoryURL, gs.Path)
			if err != nil {
				cancel()
				return err
			}
		case err != nil:
			cancel()
			return fmt.Errorf("error checking if repo exists %q: %v", gs.RepositoryURL, err)
		default:
			fmt.Println(pull(ctx, gs.Path))
		}

		cancel()
		time.Sleep(time.Second * time.Duration(defaultWait))
	}
}

func pull(ctx context.Context, path string) error {
	fmt.Println("pull")
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return err
	}

	// Print the latest commit that was just pulled
	ref, err := r.Head()
	if err != nil {
		return err
	}
	fmt.Println(ref)
	return nil
}

func cloneRepo(ctx context.Context, repo string, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	})
	fmt.Println(err)
	return err
}
