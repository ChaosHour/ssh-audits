package sshkey

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/melbahja/goph"
)

func AddPublicKeyToServer(client *goph.Client) error {
	// Get the public keys from the SSH agent
	output, err := exec.Command("ssh-add", "-L").Output()
	if err != nil {
		return fmt.Errorf("could not get public keys from SSH agent: %w", err)
	}

	// Split the output into lines
	lines := strings.Split(string(output), "\n")

	// Ensure .ssh directory exists
	_, err = client.Run("mkdir -p ~/.ssh")
	if err != nil {
		return fmt.Errorf("could not create .ssh directory: %w", err)
	}

	// Append each public key to the authorized_keys file
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		command := fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", line)
		_, err = client.Run(command)
		if err != nil {
			return fmt.Errorf("could not append public key to authorized_keys: %w", err)
		}
	}

	return nil
}
