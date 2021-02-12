package main

import (
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

type OmnibusTestSuite struct {
	suite.Suite

	fqdn            string
	sslEnabled      bool
	disableReset    bool
	localMattermost string
	localOmnibus    string
}

func (s *OmnibusTestSuite) SetupAll() {
	s.T().Log("Ensuring requisites are met")
	Cmd("apt-get install -y apt-utils")

	if s.IsLocal() {
		s.T().Log("Ensuring local package paths exist")
		mattermostPath, err := filepath.Abs(s.localMattermost)
		s.Require().NoError(err)
		s.fileExists(mattermostPath)
		omnibusPath, err := filepath.Abs(s.localOmnibus)
		s.Require().NoError(err)
		s.fileExists(omnibusPath)
	}
}

func (s *OmnibusTestSuite) SetupTest() {
	if !s.disableReset {
		if err := s.reset(); err != nil {
			s.Require().NoError(err, "Error resetting server state")
		}
	}
}

// Returns a URL prefixed by http or https depending on the
// s.sslEnabled boolean. Defaults the domain name to localhost if
// s.fqdn is empty
func (s *OmnibusTestSuite) URL() string {
	fqdn := s.fqdn
	if fqdn == "" {
		fqdn = DEFAULT_FQDN
	}

	if s.sslEnabled {
		return "https://" + fqdn
	}
	return "http://" + fqdn
}

func (s *OmnibusTestSuite) IsLocal() bool {
	return s.localMattermost != "" && s.localOmnibus != ""
}

// Marks a test as runnable only in a HTTPS enabled server
func (s *OmnibusTestSuite) httpsOnly() {
	if !s.sslEnabled {
		s.T().Skip("HTTPS only test, skipping...")
	}
}

// Marks a test as runnable only in a HTTPS disabled server
func (s *OmnibusTestSuite) httpOnly() {
	if s.sslEnabled {
		s.T().Skip("HTTP only test, skipping...")
	}
}
