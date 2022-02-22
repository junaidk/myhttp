## Build

`go build -o bin/myhttp main.go`

## Run

Pass list of URLs as arguments:

`./myhttp http://www.bing.com http://google.com`

Program will print out the URL and MD5 hash of the body for each URL.

To set number of concurrent requests, use `-parallel` flag:

`./myhttp -parallel 4 http://www.bing.com http://google.com`

example output:

```shell
./bin/myhttp -parallel 4 http://www.bing.com http://google.com
http://www.bing.com 2abd0afbdeea99cf3cebbd660d542aba
http://google.com 3c983c7d603b35db2a2861cced333188

```

Invalid URL will be ignored from the output.

## Test

`go test -v -race ./lib`
