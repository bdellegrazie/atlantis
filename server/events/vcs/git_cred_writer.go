package vcs

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

// WriteGitCreds stores the url, username, token and expiry (if supplied) to the git-credential-cache helper via the git credential protocol
// Used for authenticating with git over HTTPS
func WriteGitCreds(gitUser string, gitToken string, gitTokenExpiry time.Time, gitHostname string, logger logging.SimpleLogging) error {
	credentialCmd := exec.Command("git", "config", "--global", "credential.helper", "cache", "--timeout=86400")
	if out, err := credentialCmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "There was an error running %s: %s", strings.Join(credentialCmd.Args, " "), string(out))
	}
	logger.Info("successfully ran %s", strings.Join(credentialCmd.Args, " "))

	urlCmd := exec.Command("git", "config", "--global", fmt.Sprintf("url.https://%s@%s.insteadOf", gitUser, gitHostname), fmt.Sprintf("ssh://git@%s", gitHostname)) // nolint: gosec
	if out, err := urlCmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "There was an error running %s: %s", strings.Join(urlCmd.Args, " "), string(out))
	}
	logger.Info("successfully ran %s", strings.Join(urlCmd.Args, " "))

	credentialCmd = exec.Command("git", "credential", "approve")
	stdin, err := credentialCmd.StdinPipe()
	if err != nil {
		return errors.Wrapf(err, "There was an error getting stdin of %s", strings.Join(credentialCmd.Args, " "))
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, fmt.Sprintf("url=https://%s\nusername=%s\npassword=%s\n", gitHostname, gitUser, gitToken))
		if !time.Time.IsZero(gitTokenExpiry) {
			io.WriteString(stdin, fmt.Sprintf("password_expiry_utc=%d\n", gitTokenExpiry.Unix()))
		}
	}()
	if out, err := credentialCmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "There was an error running %s: %s", strings.Join(credentialCmd.Args, " "), string(out))
	}
	logger.Info("successfully ran %s", strings.Join(credentialCmd.Args, " "))

	return nil
}
