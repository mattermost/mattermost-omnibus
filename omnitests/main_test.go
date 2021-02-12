package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	CONFIG_PATH   = "/etc/mattermost/mmomni.yml"
	MMOMNI_PATH   = "/usr/local/bin/mmomni"
	MMCTL_PATH    = "/usr/local/bin/mmctl"
	DATA_DIR      = "/var/opt/mattermost/data"
	LOGS_DIR      = "/var/log/mattermost"
	DEFAULT_FQDN  = "localhost"
	DEFAULT_EMAIL = "syadmin@example.com"
)

var (
	fqdn            = flag.String("fqdn", "", "The domain name of the server where the test are running. Will run the tests with SSL expectations if provided")
	disableReset    = flag.Bool("disable-reset", false, "Disables the reset process of the server that is run before each test")
	localMattermost = flag.String("local-mattermost", "", "The local Mattermost package to use instead the remote one")
	localOmnibus    = flag.String("local-omnibus", "", "The local Omnibus package to use instead the remote one")
	testSuite       *OmnibusTestSuite
)

func TestOmnibusTestSuite(t *testing.T) {
	suite.Run(t, testSuite)
}

func TestMain(m *testing.M) {
	flag.Parse()

	if os.Getuid() != 0 {
		fmt.Fprintln(os.Stderr, "Omnibus tests need to run as root")
		os.Exit(1)
	}

	sslEnabled := false
	if *fqdn != "" {
		sslEnabled = true
	}

	testSuite = &OmnibusTestSuite{
		fqdn:            *fqdn,
		sslEnabled:      sslEnabled,
		disableReset:    *disableReset,
		localMattermost: *localMattermost,
		localOmnibus:    *localOmnibus,
	}

	os.Exit(m.Run())
}
