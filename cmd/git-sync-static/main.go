package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	syncserve "github.com/saadullahsaeed/git-sync-static-server/lib"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	flag.String("repo", "", "Repository URL")
	flag.String("dest", "/tmp/git", "Destination directory")
	flag.String("branch", "master", "Branch to sync")
	flag.String("port", "3000", "HTTP Port")
	flag.Int("wait", 60, "Number of seconds to wait before each sync")
	flag.String("ssh-key-path", "", "Path of the SSH key for Auth (if using SSH)")
	flag.Bool("ssh-known-hosts", true, "Toggle SSH known_hosts verification")

	//Server flags
	flag.String("root-dir", "/", "root directory to serve")

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

	logger := log.WithContext(context.Background())
	destPath := fmt.Sprintf("%s/%s", dest, parts[len(parts)-1])
	gs := &syncserve.GitSync{
		RepositoryURL:          repo,
		Destination:            dest,
		Path:                   destPath,
		Branch:                 viper.GetString("branch"),
		KeyPath:                viper.GetString("ssh-key-path"),
		KnownHostsVerification: viper.GetBool("ssh-known-hosts"),
		Logger:                 logger,
	}

	go func() {
		errChan <- gs.Start()
	}()

	//start the server
	root := fmt.Sprintf("%s/%s", destPath, strings.TrimPrefix(viper.GetString("root-dir"), "/"))
	go func(destPath string) {
		fs := http.FileServer(syncserve.NewNeuteredFileSystem(http.Dir(destPath), logger))
		http.Handle("/", fs)

		port := fmt.Sprintf(":%s", viper.GetString("port"))
		log.WithField("port", port).Info("Starting server")
		http.ListenAndServe(port, nil)
	}(root)

	fmt.Println(<-errChan)
}
