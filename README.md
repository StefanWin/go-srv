# `go-srv`

Commandline tool to start a http fileserver in the current directory (`$CWD`).

## usage
```bash
$> go-srv # starts server on $PWD on localhost:6969
$> go-srv --port=3000 # set port to 3000
$> go-srv --quiet # disable logging
$> go-srv --port=1234 --read-timeout=10
$> go-srv --bind=0.0.0.0 --cors --open # serve on all interfaces with CORS and auto-open browser
$> go-srv --dir=./public --headers="X-Frame-Options:DENY" # serve specific directory with custom headers
```

## flags
- `--port=1234` - specify which port to run on (default: `6969`)
- `--bind=localhost` - the network interface to bind to (default: `localhost`, use `0.0.0.0` for all interfaces)
- `--dir=/path/to/serve` - directory to serve (defaults to current working directory)
- `--quiet` - disable logging to `stdout`
- `--cors` - enable CORS headers for cross-origin requests
- `--dotfiles` - also serve dotfiles and dot-directories such as `.git/` and `.env` (hidden by default: they return `404` and are left out of directory listings)
- `--open` - automatically open browser on startup
- `--headers="Key:Value,Key2:Value2"` - custom headers in format `Key1:Value1,Key2:Value2`
- `--read-timeout=15` - specify the http server's read timeout in seconds (default: `15`)
- `--write-timeout=15` - specify the http server's write timeout in seconds (default: `15`)

## build
- clone the repository
- `go install`
- alternatively, you can download a release from the sidebar

# releases
pre-built releases are available [here](https://github.com/StefanWin/go-srv/releases)
- pre-built OS targets: `linux`, `windows`, `darwin`
- pre-built architectures: `amd64`, `arm64`