# Get-sync-static-server

[This is WIP] Inspried by git-sync, this is an attempt to use the pure Go implementation of git for sync and serving static content from a container.


## Usecase

To deploy a container on Kubernetes that can serve static content and also keep that said content up to date from a specific git repo. 

## Build

```
make build
```


## Build Container

```
make build container

docker run -p 9009:9009 $(REGISTRY):latest --repo https://github.com/EnterpriseQualityCoding/FizzBuzzEnterpriseEdition
```


## Usage

```
go run cmd/git-sync-static/main.go --repo https://github.com/EnterpriseQualityCoding/FizzBuzzEnterpriseEdition
```