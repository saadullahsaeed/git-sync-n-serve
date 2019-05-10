# Get-sync-n-serve

Inspried by git-sync, this is an attempt to use the pure Go implementation of git for sync and serving static content using a single binary. GSnS also supports webhook notifications when an event occurs, these can be use to post on a slack channel for example. 

The webhook notifications support custom payload templates. 

This is not battle tested at all. Please use with caution.

## Usecase

To deploy a container on Kubernetes that can serve static content and also keep that said content up to date from a specific git repo. 

## Build

```
make build
```

## Usage
The following example uses a Slack Incoming URL with a custom template to generate a JSON message for the webhook notification payload.

```
go run cmd/git-sync-static/main.go --repo git@github.com:saadullahsaeed/some-repo.git --ssh-key-path ~/.ssh/path_to_key --root-dir / --branch master --webhook-url https://hooks.slack.com/services/XXXX/XXXX/XXXXXX --webhook-method POST --webhook-payload-template '{"text": "{{ .String }} "}'
```