package chainsuite

import (
	"context"
	"io"
	"net/http"
)

func CheckEndpoint(ctx context.Context, url string, f func([]byte) error) error {
	resp, err := http.Get(url) //nolint:gosec
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
