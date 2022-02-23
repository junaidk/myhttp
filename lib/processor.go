package lib

import (
	"context"
	"crypto"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	urllib "net/url"
	"strings"
	"sync"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Processor struct {
	httpClient    HTTPClient
	parallelCount int
	wg            *sync.WaitGroup
	hashAlgo      crypto.Hash
}

type Result struct {
	Url     string
	MD5Hash string
}

func (res *Result) String() string {
	return fmt.Sprintf("%s %s\n", res.Url, res.MD5Hash)
}

func NewProcessor(httpClient HTTPClient, parallelCount int, hashAlgo crypto.Hash) *Processor {
	return &Processor{
		httpClient:    httpClient,
		parallelCount: parallelCount,
		wg:            &sync.WaitGroup{},
		hashAlgo:      hashAlgo,
	}
}

// Run starts processing the given urls in parallel and controll the program flow
func (p *Processor) Run(ctx context.Context, urls []string, output io.Writer) {

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	for w := 1; w <= p.parallelCount; w++ {
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
		md5sum, err := p.getHashFromRequest(ctx, url)

		if err == nil {
			results <- Result{Url: url, MD5Hash: md5sum}
		} //else {
		//  // print error to stderr. We can also use Err channel to send errors to the main thread
		//  // but this is not explicitly required for this implementation
		// 	fmt.Printf("Error: %s\n", err)
		//}
	}
}

// getHashFromRequest returns the md5sum of the response body
func (p *Processor) getHashFromRequest(ctx context.Context, url string) (md5sum string, err error) {
	resp, err := p.httpRequest(ctx, url)
	if err != nil {
		return
	}
	defer resp.Close()

	md5sum, err = p.getHash(resp)
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
	resp, err := p.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// getHash get the hash from io.reader
func (p *Processor) getHash(data io.Reader) (string, error) {
	hash := p.hashAlgo.New()

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
	url, err := urllib.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid url: %s", urlStr)
	}
	if err == nil && url.Scheme == "" {
		return "http://" + urlStr, nil
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		return "", fmt.Errorf("invalid url: %s", urlStr)
	}
	return urlStr, nil
}
