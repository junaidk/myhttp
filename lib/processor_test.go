package lib

import (
	"bytes"
	"context"
	"crypto"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type MockClient struct {
	ReqMap map[string]string
}

func (m *MockClient) DoFunc(req *http.Request) (*http.Response, error) {
	body, ok := m.ReqMap[req.URL.String()]
	if !ok {
		body = "Not Found"
	}
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}, nil
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestHttpRequest(t *testing.T) {
	client := &MockClient{}
	client.ReqMap = map[string]string{
		"http://www.google.com": "OK",
	}

	processor := NewProcessor(client, 1, crypto.MD5)
	resp, err := processor.httpRequest(context.Background(), "http://www.google.com")

	if err != nil {
		t.Errorf("Error: %s", err)
	}

	out, err := ioutil.ReadAll(resp)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if string(out) != "OK" {
		t.Errorf("Error: %s", err)
	}

}

func TestGetMD5Hash(t *testing.T) {
	r := strings.NewReader("my request")

	p := NewProcessor(nil, 1, crypto.MD5)

	hash, err := p.getHash(r)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if hash != "0a44cf32bcd5f63fc5e047e25f991f97" {
		t.Errorf("Error: %s", hash)
	}
}

func TestValidateAndModifyUrl(t *testing.T) {

	testCases := []struct {
		url      string
		expected string
		err      bool
	}{
		{"http://www.google.com", "http://www.google.com", false},
		{"https://google.com", "https://google.com", false},
		{"www.google.com", "http://www.google.com", false},
		{"google.com", "http://google.com", false},
		{"", "", true},
	}

	for _, testCase := range testCases {
		url, err := validateAndModifyUrl(testCase.url)
		if testCase.err {
			if err == nil {
				t.Error("Non nil error")
			}
		}
		if url != testCase.expected {
			t.Errorf("Not equal: %s, %s, %s", testCase.url, url, testCase.expected)
		}
	}

}

func TestProcessor_Run(t *testing.T) {

	client := &MockClient{}
	client.ReqMap = map[string]string{
		"":                          "",
		"http://www.google.com":     "response1",
		"http://www.microsoft.com/": "response2",
		"http://www.url1.com/":      "response3",
		"https://www.url2.com/":     "response4",
		"http://www.url3.com/":      "response5",
		"http://www.url4.com/":      "response6",
		"www.url6.com/":             "response7",
	}

	processor := NewProcessor(client, 5, crypto.MD5)

	var input []string
	for k := range client.ReqMap {
		input = append(input, k)
	}

	var output bytes.Buffer
	processor.Run(context.Background(), input, &output)

	expected := []string{
		"http://www.url6.com/ 9d1ead73e678fa2f51a70a933b0bf017",
		"http://www.microsoft.com/ 6d8aa682668fbd4a324aa5299495cc69",
		"http://www.url1.com/ bd6e19d229662b5e48a358e0ec493921",
		"https://www.url2.com/ 3f432a4e962dbb22edc256518abfcaf0",
		"http://www.url4.com/ d6905a9e1498e031293b22e71b6dd795",
		"http://www.url3.com/ 0b1a1bade2c3c3e463a96b3f17f6a490",
		"http://www.google.com d20a5df8f659a0af0f08de8da34fe8bc",
	}
	_ = expected

	resp := output.String()
	for _, exp := range expected {
		if !strings.Contains(resp, exp) {
			t.Errorf("Not found: %s", exp)
		}
	}
}
