// Package sftp uploads local scripts to a remote host and executes them.
package sftp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/melbahja/goph"
)

// UploadFileAndExecute uploads a local script to a private temp directory on
// the remote host, runs it, prints its output, and cleans up.
func UploadFileAndExecute(client *goph.Client, localPath string) error {
	if _, err := os.Stat(localPath); err != nil {
		return fmt.Errorf("error accessing local file: %w", err)
	}

	out, err := client.Run("mktemp -d /tmp/ssh-audits.XXXXXX")
	if err != nil {
		return fmt.Errorf("creating remote temp directory: %w", err)
	}
	remoteDir := strings.TrimSpace(string(out))
	if !strings.HasPrefix(remoteDir, "/tmp/") {
		return fmt.Errorf("unexpected mktemp output: %q", out)
	}
	defer func() {
		if _, err := client.Run("rm -rf " + shellQuote(remoteDir)); err != nil {
			fmt.Printf("Warning: failed to clean up %s: %v\n", remoteDir, err)
		}
	}()

	remotePath := remoteDir + "/" + filepath.Base(localPath)
	fmt.Printf("Uploading %s to %s\n", localPath, remotePath)
	if err := client.Upload(localPath, remotePath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	quoted := shellQuote(remotePath)
	if _, err := client.Run("chmod +x " + quoted); err != nil {
		return fmt.Errorf("making file executable: %w", err)
	}

	fmt.Printf("Executing %s\n", remotePath)
	output, err := client.Run(quoted)
	if err != nil {
		return fmt.Errorf("executing file: %w\nOutput: %s", err, output)
	}
	fmt.Println(strings.TrimSpace(string(output)))
	return nil
}

// shellQuote wraps s in single quotes so it is safe to embed in a remote
// shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
