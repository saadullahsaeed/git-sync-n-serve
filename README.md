# Get-sync-n-serve

Inspried by git-sync, this is an attempt to use the pure Go implementation of git for sync and serving static content using a single binary.

This is not battle tested at all. Please use with caution.

## Usecase

To deploy a container on Kubernetes that can serve static content and also keep that said content up to date from a specific git repo. 

## Build

```
make build
```

## Usage

```
go run cmd/git-sync-static/main.go --repo https://github.com/EnterpriseQualityCoding/FizzBuzzEnterpriseEdition --root-dir / --branch master 
```