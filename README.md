# fetch-webpage

```shell-session
# local
$ go mod download
$ go run ./cmd/fetch-webpage https://www.google.com

# docker
$ docker build -t fetch-webpage .
$ docker run --rm -v $(pwd):/app fetch-webpage https://www.google.com
```
