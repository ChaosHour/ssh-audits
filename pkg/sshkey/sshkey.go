package sshkey

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/melbahja/goph"
)

func AddPublicKeyToServer(client *goph.Client) error {
	// Get fingerprints of existing keys first
	existingKeys, err := getExistingKeyFingerprints(client)
	if err != nil {
		return fmt.Errorf("failed to get existing keys: %w", err)
	}

	// Get local keys
	output, err := exec.Command("ssh-add", "-L").Output()
	if err != nil {
		return fmt.Errorf("failed to get public keys: %w", err)
	}

	for _, key := range strings.Split(string(output), "\n") {
		if key = strings.TrimSpace(key); key == "" {
			continue
		}

		// Check if key already exists
		fingerprint := getKeyFingerprint(key)
		if existingKeys[fingerprint] {
			fmt.Printf("Key %s already exists, skipping\n", fingerprint)
			continue
		}

		if err := appendKey(client, key); err != nil {
			return err
		}
	}
	return nil
}

func getKeyFingerprint(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func getExistingKeyFingerprints(client *goph.Client) (map[string]bool, error) {
	fingerprints := make(map[string]bool)

	out, err := client.Run("cat ~/.ssh/authorized_keys")
	if err != nil {
		// If file doesn't exist, return empty map
		if strings.Contains(err.Error(), "No such file") {
			return fingerprints, nil
		}
		return nil, fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			fingerprints[getKeyFingerprint(line)] = true
		}
	}
	return fingerprints, nil
}

func appendKey(client *goph.Client, key string) error {
	// Ensure .ssh directory exists with correct permissions
	if _, err := client.Run("mkdir -p ~/.ssh && chmod 700 ~/.ssh"); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Append the key
	cmd := fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", key)
	if _, err := client.Run(cmd); err != nil {
		return fmt.Errorf("failed to append key: %w", err)
	}

	// Set correct permissions
	if _, err := client.Run("chmod 600 ~/.ssh/authorized_keys"); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// ...existing code...
