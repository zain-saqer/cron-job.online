package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClient_Request(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, client")
		if err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
	}))
	defer testServer.Close()

	t.Run(`client will timeout`, func(t *testing.T) {
		client := NewClient(5 * time.Millisecond)
		req, err := http.NewRequest(`GET`, testServer.URL, strings.NewReader(``))
		if err != nil {
			t.Error(err)
		}
		ctx := context.Background()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = client.Request(ctx, req)
			if err == nil {
				t.Error(`expect request to timeout, request didn't timeout`)
			}
			var e net.Error
			if !(errors.As(err, &e) && err.(*url.Error).Timeout()) {
				t.Error(`unexpected error:`, err)
			}
		}()
		wg.Wait()
	})

	t.Run(`client will cancel`, func(t *testing.T) {
		client := NewClient(5 * time.Millisecond)
		req, err := http.NewRequest(`GET`, testServer.URL, strings.NewReader(``))
		if err != nil {
			t.Error(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = client.Request(ctx, req)
			if err == nil {
				t.Error(`expect request to timeout, request didn't timeout`)
			}
			if !errors.Is(err, context.Canceled) {
				t.Error(`unexpected error:`, err)
			}
		}()
		select {
		case <-time.After(time.Millisecond):
			cancel()
		}
		wg.Wait()
	})
}
