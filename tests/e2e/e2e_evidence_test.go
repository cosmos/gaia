package e2e

import "fmt"

func (s *IntegrationTestSuite) TestEvidence() {
	s.Run("teste query evidences", func() {
		chain := s.chainA
		chainAPI := fmt.Sprintf("http://%s", s.valResources[chain.id][0].GetHostPort("1317/tcp"))
		res, err := queryAllEvidence(chainAPI)
		s.Require().NoError(err)
		s.Require().Equal(numberOfEvidences, len(res.Evidence))
	})
}
