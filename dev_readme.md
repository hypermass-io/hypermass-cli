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



