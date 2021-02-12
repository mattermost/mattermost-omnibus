package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	mmoModel "github.com/mattermost/mattermost-omnibus/mmomni/model"
	mmModel "github.com/mattermost/mattermost-server/v5/model"
)

// ToDo: show command output as it's produced, maybe if a flag is set
// ToDo: add context to the commands to exit if they run for more than a threshold

func Cmd(cmdStr string, a ...interface{}) *exec.Cmd {
	parts := strings.Split(fmt.Sprintf(cmdStr, a...), " ")
	if len(parts) > 1 {
		return exec.Command(parts[0], parts[1:]...)
	}
	return exec.Command(cmdStr)
}

func (s *OmnibusTestSuite) curl() string {
	s.T().Log("Running repository configuration curl...")

	// check if mattermost repository is configured already to skip curl
	_, err := Cmd("grep deb.packages.mattermost.com /etc/apt/sources.list").CombinedOutput()
	if err == nil {
		s.T().Log("Mattermost repository found, skipping curl configuration")
		return ""
	}

	curlCmd := Cmd("curl -o- https://deb.packages.mattermost.com/repo-setup.sh")
	bashCmd := Cmd("sudo bash")

	r, err := curlCmd.StdoutPipe()
	s.Require().Nil(err)
	bashCmd.Stdin = r

	var out bytes.Buffer
	bashCmd.Stdout = &out

	err = curlCmd.Start()
	s.Require().Nil(err)
	err = bashCmd.Start()
	s.Require().Nil(err)
	err = curlCmd.Wait()
	s.Require().Nil(err)
	err = bashCmd.Wait()
	s.Require().Nil(err)
	r.Close()

	return out.String()
}

func (s *OmnibusTestSuite) run(cmdStr string, env ...string) string {
	if len(env) > 0 {
		s.T().Logf("Running %q with env %v\n", cmdStr, env)
	} else {
		s.T().Logf("Running %q\n", cmdStr)
	}

	cmd := Cmd(cmdStr)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.T().Logf(string(output))
		s.FailNow(fmt.Sprintf("Command %q failed", cmdStr))
	}
	return string(output)
}

func (s *OmnibusTestSuite) setSelections(sel string) {
	s.T().Logf("Selecting %q in debconf\n", sel)
	cmd := Cmd("debconf-set-selections")
	cmd.Stdin = strings.NewReader(sel)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.T().Log(string(output))
		s.FailNow(fmt.Sprintf("Setting selection %q failed", sel))
	}
}

func (s *OmnibusTestSuite) addHost(host, addr string) {
	s.T().Logf("Adding host %q for address %q to /etc/hosts", host, addr)

	_, err := Cmd("grep %s /etc/hosts", host).CombinedOutput()
	if err == nil {
		s.T().Logf("Host %q found in /etc/hosts, skipping", host)
		return
	}

	_, err = Cmd("grep %s /etc/hosts", addr).CombinedOutput()
	if err == nil {
		s.T().Logf("Addr %q found in /etc/hosts, skipping", addr)
		return
	}

	hostsFile, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0644)
	s.Require().NoError(err)
	defer hostsFile.Close()

	_, err = hostsFile.WriteString(fmt.Sprintf("%s\t%s\n", addr, host))
	s.Require().NoError(err)
}

func (s *OmnibusTestSuite) configureInstall(fqdn, email string) {
	s.setSelections(fmt.Sprintf("mattermost-omnibus mattermost-omnibus/domain string %s", fqdn))
	s.setSelections(fmt.Sprintf("mattermost-omnibus mattermost-omnibus/email string %s", email))
}

// installOmnibus configures and install the Omnibus package. args, if
// provided optionally capture the version of Omnibus to install. If
// the flags for local packages have been provided, and no specific
// version is passed to the helper, it will install the local packages
// instead of the remote ones
func (s *OmnibusTestSuite) installOmnibus(fqdn, email string, args ...string) {
	var version string
	if len(args) > 0 {
		version = args[0]
	}

	s.curl()
	s.configureInstall(fqdn, email)

	var installCmd string
	if s.IsLocal() && version == "" {
		s.T().Logf("Installing Omnibus for domain %q, email %q and ssl %v. Paths %q and %q", fqdn, email, s.sslEnabled, s.localMattermost, s.localOmnibus)
		mattermostPath, err := filepath.Abs(s.localMattermost)
		s.Require().NoError(err)
		omnibusPath, err := filepath.Abs(s.localOmnibus)
		s.Require().NoError(err)
		installCmd = fmt.Sprintf("apt install -y %s %s", mattermostPath, omnibusPath)
	} else if version != "" {
		s.T().Logf("Installing Omnibus for domain %q, email %q, ssl %v and version %q", fqdn, email, s.sslEnabled, version)
		installCmd = fmt.Sprintf("apt-get install -y mattermost=%s mattermost-omnibus=%s", version, version)
	} else {
		s.T().Logf("Installing Omnibus for domain %q, email %q and ssl %v", fqdn, email, s.sslEnabled)
		installCmd = "apt-get install -y mattermost-omnibus"
	}

	if s.sslEnabled {
		s.run(installCmd)
		s.addHost(s.fqdn, "127.0.0.1")
	} else {
		s.run(installCmd, "MMO_HTTPS=false")
	}

	s.fileExists(CONFIG_PATH)
	s.fileExists(MMOMNI_PATH)
	s.fileExists(MMCTL_PATH)
	s.directoryExists(DATA_DIR)
	s.directoryExists(LOGS_DIR)
	s.T().Log("Validating the configuration")
	config := s.config()
	s.Require().Equal(fqdn, *config.FQDN)
	s.Require().Equal(email, *config.Email)
	s.Require().Equal(s.sslEnabled, *config.HTTPS)

	s.checkURL(s.URL(), 200, "<title>Mattermost</title>")
}

func (s *OmnibusTestSuite) removeOmnibus() {
	s.T().Log("Removing Omnibus")
	s.run("apt-get remove -y mattermost mattermost-omnibus")
	s.pathNotExists(MMOMNI_PATH)
	s.pathNotExists(MMCTL_PATH)
	// config should still exist as we only removed the package
	// instead of purging it
	s.fileExists(CONFIG_PATH)
}

func (s *OmnibusTestSuite) purgeOmnibus() {
	s.removeOmnibus()
	s.T().Log("Purging Omnibus")
	s.run("apt-get purge -y mattermost mattermost-omnibus")
	s.pathNotExists(CONFIG_PATH)
	s.pathNotExists(DATA_DIR)
	s.pathNotExists(LOGS_DIR)
}

func (s *OmnibusTestSuite) reconfigure() {
	s.T().Log("Reconfiguring Omnibus")
	s.run("mmomni reconfigure")
}

func (s *OmnibusTestSuite) backup(args ...string) string {
	var path string
	if len(args) > 0 {
		path = args[0]
	} else {
		path = filepath.Join(os.TempDir(), fmt.Sprintf("mmomni-backup_%s.tgz", time.Now().Format("200601021504")))
	}

	s.T().Logf("Taking backup at path %q", path)
	s.run(fmt.Sprintf("mmomni backup --output %s", path))

	return path
}

func (s *OmnibusTestSuite) restore(path string) {
	s.T().Logf("Restoring backup from %q", path)
	s.run(fmt.Sprintf("mmomni restore %s", path))
}

// Resets the state of the server under test, uninstalling mattermost
// and removing the database if it exists
func (s *OmnibusTestSuite) reset() error {
	s.T().Log("Resetting server state...")

	Cmd("systemctl stop mattermost").Run()
	Cmd("sudo -u postgres dropdb mattermost").Run()
	Cmd("sudo -u postgres dropuser mattermost").Run()
	Cmd("apt-get purge -y mattermost mattermost-omnibus").Run()

	return nil
}

// args optionally captures the contents to check as part of the
// response of the URL request
func (s *OmnibusTestSuite) checkURL(url string, expectedCode int, args ...string) {
	var expectedContents string
	if len(args) > 0 {
		expectedContents = args[0]
	}

	if expectedContents != "" {
		s.T().Logf("Checking URL %q for status code %d and content %q", url, expectedCode, expectedContents)
	} else {
		s.T().Logf("Checking URL %q for status code %d", url, expectedCode)
	}

	resp, err := http.Get(url)
	s.Require().Nil(err)
	defer resp.Body.Close()

	s.Require().Equal(expectedCode, resp.StatusCode)

	if expectedContents != "" {
		body, err := ioutil.ReadAll(resp.Body)
		s.Require().NoError(err, "error reading the response for %q", url)
		s.Require().Contains(string(body), expectedContents)
	}
}

func (s *OmnibusTestSuite) pathExists(path string, isDir bool) {
	entity := "file"
	if isDir {
		entity = "dir"
	}

	s.T().Logf("Cheking that %s %q exists", entity, path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		s.FailNow(fmt.Sprintf("Path %q doesn't exist", path))
	} else if err != nil {
		s.FailNow(fmt.Sprintf("Error when checking path %q: %s", path, err))
	}

	if isDir && !info.IsDir() {
		s.FailNow(fmt.Sprintf("Path %q is a file", path))
	} else if !isDir && info.IsDir() {
		s.FailNow(fmt.Sprintf("Path %q is a directory", path))
	}
}

func (s *OmnibusTestSuite) fileExists(path string) {
	s.pathExists(path, false)
}

func (s *OmnibusTestSuite) directoryExists(path string) {
	s.pathExists(path, true)
}

func (s *OmnibusTestSuite) pathNotExists(path string) {
	s.T().Logf("Checking that path %q does not exist", path)
	_, err := os.Stat(path)
	if err == nil {
		s.FailNow(fmt.Sprintf("Path %q exists", path))
	}
	if !os.IsNotExist(err) {
		s.FailNow(fmt.Sprintf("Error checking path %q: %s", path, err))
	}
}

func (s *OmnibusTestSuite) config(args ...string) *mmoModel.Config {
	configPath := CONFIG_PATH
	if len(args) > 0 {
		configPath = args[0]
	}
	config, err := mmoModel.ReadConfig(configPath)
	s.Require().Nil(err)

	return config
}

func (s *OmnibusTestSuite) anonymousClient() *mmModel.Client4 {
	return mmModel.NewAPIv4Client(s.URL())
}

func (s *OmnibusTestSuite) checkAPIVersion(version string) {
	s.T().Logf("Checking API version %q at %q", version, s.URL())
	_, resp := s.anonymousClient().GetPing()
	s.Require().Nil(resp.Error)
	s.Require().Contains(resp.ServerVersion, version)
}

func (s *OmnibusTestSuite) createInitialUser(email, username, password string) {
	s.T().Logf("Creating initial user %q at %q", username, s.URL())
	user := &mmModel.User{
		Username: username,
		Password: password,
		Email:    email,
	}
	_, resp := s.anonymousClient().CreateUser(user)
	s.Require().Nil(resp.Error)
}

func (s *OmnibusTestSuite) client(username, password string) *mmModel.Client4 {
	client := s.anonymousClient()
	_, resp := client.Login(username, password)
	s.Require().Nil(resp.Error)

	return client
}

func (s *OmnibusTestSuite) doFileContain(file, text string) bool {
	fileBytes, err := ioutil.ReadFile(file)
	s.Require().NoError(err)

	return strings.Contains(string(fileBytes), text)
}

func (s *OmnibusTestSuite) fileNotContains(file, text string) {
	s.Require().False(s.doFileContain(file, text))
}

func (s *OmnibusTestSuite) fileContains(file, text string) {
	s.Require().True(s.doFileContain(file, text))
}
