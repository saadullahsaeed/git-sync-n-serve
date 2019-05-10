package syncserve

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	defaultTimeout = 30

	defaultWait = 15
)

// GitSync struct represents something
type GitSync struct {
	RepositoryURL          string
	Destination            string
	Path                   string
	Branch                 string
	KeyPath                string
	KnownHostsVerification bool
	Logger                 *log.Entry
	EventChannel           chan Event
}

// Start ...
func (gs *GitSync) Start() error {
	var err error
	defer func() {
		if err != nil && !os.IsNotExist(err) {
			gs.Logger.Error(err)
		}
	}()

	gs.Logger.WithFields(log.Fields{
		"repository": gs.RepositoryURL,
		"path":       gs.Path,
		"branch":     gs.Branch,
	}).Info("Starting sync process")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(defaultTimeout))
		_, err = os.Stat(gs.Path)
		switch {
		case os.IsNotExist(err):
			gs.Logger.WithFields(log.Fields{
				"path": gs.Path,
			}).Info("Git clone")
			err = gs.Clone(ctx)
			if err != nil {
				cancel()
				return err
			}
			gs.Logger.Info("Clone completed")
			gs.Logger.Info("event")
			gs.EventChannel <- Event{Repository: gs.RepositoryURL, Action: "Cloned"}
		case err != nil:
			cancel()
			return fmt.Errorf("error checking if repo exists %q: %v", gs.RepositoryURL, err)
		default:
			gs.Logger.WithFields(log.Fields{
				"path": gs.Path,
			}).Info("Git pull")
			err = gs.Pull(ctx, gs.Path)
			if err != nil {
				cancel()
				return err
			}
		}
		cancel()
		gs.Logger.Info("event")
		gs.EventChannel <- Event{Repository: gs.RepositoryURL, Action: "Updated"}
		time.Sleep(time.Minute * time.Duration(defaultWait))
	}
}

// Pull ...
func (gs *GitSync) Pull(ctx context.Context, path string) error {
	var auth transport.AuthMethod
	var err error
	if gs.KeyPath != "" {
		auth, err = getAuth(gs.KeyPath, gs.KnownHostsVerification)
		if err != nil {
			return err
		}
	}
	return _pull(ctx, path, auth)
}

// Clone ...
func (gs *GitSync) Clone(ctx context.Context) error {
	var auth transport.AuthMethod
	var err error
	if gs.KeyPath != "" {
		auth, err = getAuth(gs.KeyPath, gs.KnownHostsVerification)
		if err != nil {
			return err
		}
	}
	return _clone(ctx, gs.RepositoryURL, gs.Branch, gs.Path, auth)
}

func _pull(ctx context.Context, path string, auth transport.AuthMethod) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	pullops := &git.PullOptions{RemoteName: "origin"}
	if auth != nil {
		pullops.Auth = auth
	}

	err = w.Pull(pullops)
	if err != nil {
		// we do not consider this an error
		if err.Error() == git.NoErrAlreadyUpToDate.Error() {
			return nil
		}
		return err
	}

	_, err = r.Head()
	if err != nil {
		return err
	}
	return nil
}

func _clone(ctx context.Context, repo, branch, path string, auth transport.AuthMethod) error {
	cloneops := &git.CloneOptions{
		URL:           repo,
		Progress:      progressWriter{},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	}

	if auth != nil {
		cloneops.Auth = auth
	}

	_, err := git.PlainClone(path, false, cloneops)
	return err
}

func getAuth(keyPath string, knownHostsVerification bool) (transport.AuthMethod, error) {
	pem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, err
	}

	auth := &ssh2.PublicKeys{User: "git", Signer: signer}
	if !knownHostsVerification {
		auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	return auth, nil
}

type progressWriter struct {
}

func (w progressWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
