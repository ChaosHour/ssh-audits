package main

import (
	"flag"
	"log"

	"github.com/ChaosHour/ssh-audits/pkg/sftp"
	"github.com/ChaosHour/ssh-audits/pkg/sshutil"
)

var (
	sftpFile = flag.String("sftp", "", "File to SFTP")
)

func main() {
	flag.Parse()

	if *sshutil.File == "" {
		log.Fatal("Inventory file (-i) is required")
	}

	// Handle SFTP operations
	if *sftpFile != "" {
		if err := sftp.ExecuteCommandOnHost(*sftpFile, *sshutil.File, *sshutil.Host); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Use sshutil package for all other operations
	if err := sshutil.Run(); err != nil {
		log.Fatal(err)
	}
}
