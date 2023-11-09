package http

import (
	"io"
	"net/http"
	"time"
)

type Result struct {
	Body string
}

type Client struct {
	client *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{&http.Client{Timeout: timeout}}
}

func (c *Client) Request(req *http.Request) (*Result, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Result{Body: string(body)}, nil
}
