# Hypermass CLI

## Setup
Switch to the right version of go:

```bash
export PATH="~/.go-sdk/go1.25.0/bin/bin:$PATH"
export GOPATH=~/.go-sdk/go1.25.0/bin/
```

## Dev testing
Run the program from source with this command
```bash
go run main.go
```

## Build
Build the (production) executable with this command;
```bash
go build
```
The executable will be called "hypermass"

of similarly for ('develop' for the development test environment or 'local' for local stub testing)
```bash
go build -tags development .
# or
go build -tags local .
```

## Triggering a release
Manually (change the version as needed);
```bash
RELEASE_VERSION="v0.2.1"
git tag -a $RELEASE_VERSION -m "Release $RELEASE_VERSION"
git push origin $RELEASE_VERSION
```
