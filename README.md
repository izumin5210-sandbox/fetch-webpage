# fetch-webpage

## Usage

```
Usage:
  fetch-webpage [flags] url...

Flags:
      --debug            prints more logs
  -h, --help             help for fetch-webpage
      --metadata         prints metadata
      --out-dir string   output directory
  -p, --parallel int     parallelism (default 2)
  -v, --verbose          prints logs
```

## Build

```shell-session
# local
$ go mod download
$ go run ./cmd/fetch-webpage https://www.google.com

# docker
$ docker build -t fetch-webpage .
$ docker run --rm -v $(pwd):/app fetch-webpage https://www.google.com
```
