package custom

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func RunHTTP(ctx context.Context, method, url string, headers map[string]string, body []byte) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return Result{Output: string(b), Status: resp.StatusCode}, nil
}
