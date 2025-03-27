package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func HTTPGet(endpoint string) ([]byte, error) {
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

func ReadJSON(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, errors.New("failed to read Body")
	}

	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, errors.New("failed to unmarshal response body")
	}

	return data, nil
}
