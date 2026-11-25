package chainsuite

import (
	"context"
	"io"
	"net/http"
)

func CheckEndpoint(ctx context.Context, url string, f func([]byte) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) //nolint:gosec
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return f(bts)
}
