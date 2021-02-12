package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFQDN(t *testing.T) {
	testCases := []struct {
		name         string
		fqdn         string
		expectedFQDN string
	}{
		{
			name:         "A domain name should not be modified",
			fqdn:         "mattermost.example.com",
			expectedFQDN: "mattermost.example.com",
		},
		{
			name:         "A domain name with http prefix",
			fqdn:         "http://mattermost.example.com",
			expectedFQDN: "mattermost.example.com",
		},
		{
			name:         "A domain name with https prefix and a trailing slash",
			fqdn:         "https://mattermost.example.com/",
			expectedFQDN: "mattermost.example.com",
		},
		{
			name:         "A domain name with an URL path",
			fqdn:         "https://mattermost.example.com/chat",
			expectedFQDN: "mattermost.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedFQDN := ParseFQDN(tc.fqdn)
			require.Equal(t, tc.expectedFQDN, parsedFQDN)
		})
	}
}
