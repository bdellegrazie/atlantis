package vcs_test

import (
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test github app credentials
func TestWriteGitCreds(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Cleanup(func() {
		// stop the credential-cache after setup
		cmd := exec.Command("git", "credential-cache", "exit")
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("HOME=%s", tmp))
		cmd.Run()
	})

	// Test Credential Caching
	t.Run("CreateAndUpdate", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Second * 10)
		err := vcs.WriteGitCreds("x-access-token", "token", expiresAt, "example.com", logger)
		Ok(t, err)

		expOutput := fmt.Sprintf("protocol=https\nhost=example.com\nusername=x-access-token\npassword=token\npassword_expiry_utc=%d\n", expiresAt.Unix())
		credentialCmd := exec.Command("git", "credential", "fill")
		stdin, err := credentialCmd.StdinPipe()
		Ok(t, err)
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, "url=https://example.com\nusername=x-access-token\n")
		}()
		actOutput, err := credentialCmd.CombinedOutput()
		Ok(t, err)
		Equals(t, expOutput, string(actOutput))

		expiresAt = time.Now().Add(time.Second * 10)
		err = vcs.WriteGitCreds("x-access-token", "token2", expiresAt, "example.com", logger)
		Ok(t, err)

		expOutput = fmt.Sprintf("protocol=https\nhost=example.com\nusername=x-access-token\npassword=token2\npassword_expiry_utc=%d\n", expiresAt.Unix())
		credentialCmd = exec.Command("git", "credential", "fill")
		stdin, err = credentialCmd.StdinPipe()
		Ok(t, err)
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, "url=https://example.com\nusername=x-access-token\n")
		}()
		actOutput, err = credentialCmd.CombinedOutput()
		Ok(t, err)
		Equals(t, expOutput, string(actOutput))
	})

	// Test that git is actually configured to use the credentials
	t.Run("ConfigureGitCredentialhelper", func(t *testing.T) {
		err := vcs.WriteGitCreds("user", "token", time.Now().Add(time.Second*10), "hostname", logger)
		Ok(t, err)

		expOutput := `cache`
		actOutput, err := exec.Command("git", "config", "--global", "credential.helper").Output()
		Ok(t, err)
		Equals(t, expOutput+"\n", string(actOutput))
	})

	// Test that git is configured to use https instead of ssh
	t.Run("ConfigureGitUrlOverride", func(t *testing.T) {
		err := vcs.WriteGitCreds("user", "token", time.Now().Add(time.Second*10), "hostname", logger)
		Ok(t, err)

		expOutput := `ssh://git@hostname`
		actOutput, err := exec.Command("git", "config", "--global", "url.https://user@hostname.insteadof").Output()
		Ok(t, err)
		Equals(t, expOutput+"\n", string(actOutput))
	})
}
