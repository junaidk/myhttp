package lib

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Processor struct {
	HttpClient    HTTPClient
	ParallelCount int
	wg            *sync.WaitGroup
}

type Result struct {
	Url     string
	MD5Hash string
}

func (res *Result) String() string {
	return fmt.Sprintf("%s %s\n", res.Url, res.MD5Hash)
}

func NewProcessor(httpClient HTTPClient, parallelCount int) *Processor {
	return &Processor{
		HttpClient:    httpClient,
		ParallelCount: parallelCount,
		wg:            &sync.WaitGroup{},
	}
}

// Run starts processing the given urls in parallel and controll the program flow
func (p *Processor) Run(ctx context.Context, urls []string, output io.Writer) {

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	for w := 1; w <= p.ParallelCount; w++ {
		p.wg.Add(1)
		go p.worker(ctx, jobs, results)
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	go func() {
		p.wg.Wait()
		close(results)
	}()

	strReader := strings.NewReader("")
	for res := range results {
		strReader.Reset(res.String())
		io.Copy(output, strReader)
	}

}

// worker is a function that will be executed in a goroutine and implements the core logic of the program
func (p *Processor) worker(ctx context.Context, jobs <-chan string, results chan<- Result) {
	defer p.wg.Done()
	for j := range jobs {
		url, err := validateAndModifyUrl(j)
		if err != nil {
			continue
		}
		md5sum, err := p.getMD5sumFromRequest(ctx, url)

		if err == nil {
			results <- Result{Url: url, MD5Hash: md5sum}
		}
	}
}

// getMD5sumFromRequest returns the md5sum of the response body
func (p *Processor) getMD5sumFromRequest(ctx context.Context, url string) (md5sum string, err error) {
	resp, err := p.httpRequest(ctx, url)
	if err != nil {
		return
	}
	defer resp.Close()

	md5sum, err = getMD5Hash(resp)
	if err != nil {
		return
	}
	return
}

// httpRequest send http request and get body as io.ReadCloser
func (p *Processor) httpRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.HttpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// getMD5Hash get the md5 hash from io.reader
func getMD5Hash(data io.Reader) (string, error) {
	hash := md5.New()
	_, err := io.Copy(hash, data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// validateAndModifyUrl will validate the url and add scheme to it if needed
func validateAndModifyUrl(urlStr string) (string, error) {
	if urlStr == "" {
		return "", fmt.Errorf("url is empty")
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid url: %s", urlStr)
	}
	if err == nil && u.Scheme == "" {
		return "http://" + urlStr, nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid url: %s", urlStr)
	}
	return urlStr, nil
}
