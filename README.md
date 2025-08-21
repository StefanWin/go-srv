# `go-srv`

Commandline tool to start a http fileserver in the current directory (`$CWD`).

## usage
```bash
$> go-srv # starts on server on $PWD on localhost:6969
$> go-srv --port=3000 # set port to 3000
$> go-srv --quiet # disable logging
$> go-srv --port=1234 --read-timeout=10
```

## flags
- `--port=1234` - specify which port to run on (default: `6969`)
- `--quiet` - disable logging to `stdout`
- `--read-timout` - specify the http servers read-timeout (default: 15s)

## build
- clone the repository
- `go install`
- alternatively, you can download a release from the sidebar

# releases
pre-built releases are available [here](https://github.com/StefanWin/go-srv/releases)
- pre-built OS targets: `linux`, `windows`, `darwin`
- pre-built architectures: `amd64`, `arm64`