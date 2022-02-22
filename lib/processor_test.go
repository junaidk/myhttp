package lib

import (
	"bytes"
	"context"
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

	processor := NewProcessor(client, 1)
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
	hash, err := getMD5Hash(r)
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
		"http://www.google.com":     "response1",
		"http://www.microsoft.com/": "response2",
		"http://www.azure.com/":     "response3",
	}

	processor := NewProcessor(client, 5)

	var b bytes.Buffer
	processor.Run(context.Background(), []string{"http://www.google.com", "http://www.microsoft.com/", "http://www.azure.com/"}, &b)

	expected := []string{
		"http://www.google.com d20a5df8f659a0af0f08de8da34fe8bc",
		"http://www.azure.com/ bd6e19d229662b5e48a358e0ec493921",
		"http://www.microsoft.com/ 6d8aa682668fbd4a324aa5299495cc69",
	}

	resp := b.String()
	for _, exp := range expected {
		if !strings.Contains(resp, exp) {
			t.Errorf("Not found: %s", exp)
		}
	}

}
