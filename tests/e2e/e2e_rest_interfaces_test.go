package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Helper function to read the response body and unmarshal it into a map
func readJSON(resp *http.Response) (map[string]interface{}, error) {
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read Body")
	}

	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body")
	}

	return data, nil
}

// Integration test to check if the following rest interfaces have been wired correctly:
//
//   - /node_info
//   - /syncing
//
// Assuming that check on these are not necessary, if above are working:
//
//   - /blocks/latest
//   - /validatorsets/latest
//   - /blocks/{height}
//   - /validatorsets/{height}
func (s *IntegrationTestSuite) testRestInterfaces() {
	s.Run("test rest interfaces", func() {
		var (
			testOk        = true
			valIdx        = 0
			c             = s.chainA
			endpointURL   = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
			testEndpoints = []struct {
				Path         string
				ExpectedFail bool
			}{
				{"/node_info", false},
				{"/syncing", false},
				{"/missing_endpoint", true},
			}
		)

		for _, endpoint := range testEndpoints {

			// Call the required endpoint
			resp, err := http.Get(endpointURL + endpoint.Path)
			if err != nil {
				testOk = false
				s.T().Logf("failed to get endpoint: %s, %v", endpointURL+endpoint.Path, err)
				continue
			}

			// Decode the JSON resopnse
			jsonBody, errJSON := readJSON(resp)
			if errJSON != nil {
				testOk = false
				s.T().Logf("failed to read body of endpoint: %s, %v", endpointURL+endpoint.Path, errJSON)
				continue
			}

			if endpoint.ExpectedFail == false && jsonBody["message"] == "Not Implemented" {
				testOk = false
				s.T().Logf("Encountered a Not implemented endpoint: %s", endpointURL+endpoint.Path)
				continue
			}
		}

		s.Require().Equal(true, testOk)
	})
}
