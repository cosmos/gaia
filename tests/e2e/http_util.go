package e2e

import (
	"fmt"
	"io"
	"net/http"
)

func httpGet(endpoint string) ([]byte, error) { //nolint:unused // this is called during e2e tests
	resp, err := http.Get(endpoint) //nolint:gosec // this is only used during tests
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
