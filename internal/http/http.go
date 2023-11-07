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

type Client struct {
	client *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{&http.Client{Timeout: timeout}}
}

func (c *Client) Request(ctx context.Context, req *http.Request) (*Result, error) {
	req = req.WithContext(ctx)
	resultStream := make(chan *Result)
	errStream := make(chan error)
	go func() {
		resp, err := c.client.Do(req)
		if err != nil {
			errStream <- err
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errStream <- err
			return
		}
		resultStream <- &Result{Body: string(body)}
	}()
	select {
	case <-ctx.Done():
		return &Result{}, ctx.Err()
	case result := <-resultStream:
		return result, nil
	case err := <-errStream:
		return &Result{}, err
	}
}
