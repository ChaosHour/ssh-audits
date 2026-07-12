// Package hostkeys provides the SSH host-key verification callback shared by
// every connection path.
package hostkeys

import (
	"fmt"
	"net"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// Verify checks the host against ~/.ssh/known_hosts. Known hosts with a
// matching key are accepted, a key mismatch is refused (possible MITM), and
// unknown hosts are added after printing their fingerprint.
func Verify(hostname string, remote net.Addr, key ssh.PublicKey) error {
	found, err := goph.CheckKnownHost(hostname, remote, key, "")
	if found && err != nil {
		return err
	}
	if found {
		return nil
	}

	fmt.Printf("[+] Adding %s to known_hosts (%s)\n", hostname, ssh.FingerprintSHA256(key))
	return goph.AddKnownHost(hostname, remote, key, "")
}
