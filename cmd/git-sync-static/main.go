package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	defaultTimeout = 30

	defaultWait = (time.Minute * 60)
)

func main() {
	flag.String("repo", "", "Repository URL")
	flag.String("dest", "/tmp/git", "Destination directory")
	flag.String("branch", "master", "Branch to sync")
	flag.String("port", "3000", "HTTP Port")
	flag.String("ssh-key-path", "", "Path of the SSH key for Auth (if using SSH)")

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
	gs := &GitSync{
		RepositoryURL: repo,
		Destination:   dest,
		Path:          destPath,
		KeyPath:       viper.GetString("ssh-key-path"),
	}

	go func() {
		errChan <- gs.Start()
	}()

	//start the server
	go func(destPath string) {
		// fs := http.FileServer(http.Dir(destPath))
		fs := http.FileServer(NewNeuteredFileSystem(http.Dir(destPath)))
		http.Handle("/", fs)

		log.Println("Listening...")
		http.ListenAndServe(fmt.Sprintf(":%s", viper.GetString("port")), nil)
	}(destPath)

	fmt.Println(<-errChan)
}

func NewNeuteredFileSystem(fs http.FileSystem) neuteredFileSystem {
	ns := neuteredFileSystem{fs: fs}
	ns.dotRegex = regexp.MustCompile(`^\.(.*)`)
	return ns
}

type neuteredFileSystem struct {
	fs       http.FileSystem
	dotRegex *regexp.Regexp
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	// we don't want to server .dot directories
	if nfs.dotRegex != nil {
		if nfs.dotRegex.MatchString(path[1:]) {
			return nil, errors.New("not found")
		}
	}

	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := fmt.Sprintf("%s/%s", strings.TrimSuffix(path, "/"), "index.html")
		_, err := nfs.fs.Open(index)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// GitSync struct represents something
type GitSync struct {
	RepositoryURL string
	Destination   string
	Path          string
	Branch        string
	KeyPath       string
}

// Start ...
func (gs *GitSync) Start() error {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(defaultTimeout))
		_, err := os.Stat(gs.Path)
		switch {
		case os.IsNotExist(err):
			err := gs.Clone(ctx, gs.RepositoryURL, gs.Path)
			if err != nil {
				cancel()
				return err
			}
		case err != nil:
			cancel()
			return fmt.Errorf("error checking if repo exists %q: %v", gs.RepositoryURL, err)
		default:
			err := gs.Pull(ctx, gs.Path)
			if err != nil {
				cancel()
				return err
			}
		}

		cancel()
		time.Sleep(time.Second * time.Duration(defaultWait))
	}
}

// Pull ...
func (gs *GitSync) Pull(ctx context.Context, path string) error {
	var auth transport.AuthMethod
	var err error
	if gs.KeyPath != "" {
		auth, err = getAuth(gs.KeyPath)
		if err != nil {
			return err
		}
	}
	return _pull(ctx, path, auth)
}

// Clone ...
func (gs *GitSync) Clone(ctx context.Context, repo, path string) error {
	var auth transport.AuthMethod
	var err error
	if gs.KeyPath != "" {
		auth, err = getAuth(gs.KeyPath)
		if err != nil {
			return err
		}
	}
	return _clone(ctx, repo, path, auth)
}

func _pull(ctx context.Context, path string, auth transport.AuthMethod) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	pullops := &git.PullOptions{RemoteName: "origin"}
	if auth != nil {
		pullops.Auth = auth
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(pullops)
	if err != nil {
		// we do not consider this an error
		if err.Error() == git.NoErrAlreadyUpToDate.Error() {
			return nil
		}
		return err
	}

	// Print the latest commit that was just pulled
	_, err = r.Head()
	if err != nil {
		return err
	}
	return nil
}

func _clone(ctx context.Context, repo string, path string, auth transport.AuthMethod) error {
	cloneops := &git.CloneOptions{
		URL:      repo,
		Progress: progressWriter{},
	}

	if auth != nil {
		cloneops.Auth = auth
	}

	_, err := git.PlainClone(path, false, cloneops)
	return err
}

func getAuth(keyPath string) (transport.AuthMethod, error) {
	pem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, err
	}
	return &ssh2.PublicKeys{User: "git", Signer: signer}, nil
}

type progressWriter struct {
}

func (w progressWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}
