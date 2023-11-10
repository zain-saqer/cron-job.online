package http

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Result struct {
	Body string
}

type Client interface {
	Request(ctx context.Context, method, url string, body io.Reader) (*Result, error)
}

type SimpleClient struct {
	client *http.Client
}

func (c SimpleClient) Request(ctx context.Context, method, url string, body io.Reader) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Result{Body: string(resBody)}, nil
}

func NewClient(timeout time.Duration) *SimpleClient {
	return &SimpleClient{&http.Client{Timeout: timeout}}
}
