# `go-srv`

Commandline tool to start a http fileserver in the current directory.

## usage
```bash
$> go-srv # starts on server on $PWD on localhost:6969
$> go-srv --port=3000 # set port to 3000
$> go-srv --quiet # disable logging
```

## flags
- `--port=1234` - specify which port to run on (default: `6969`)
- `--quiet` - disable logging to `stdout`

## build
- clone the repository
- `go install`
- alternatively, you can download a release from the sidebar