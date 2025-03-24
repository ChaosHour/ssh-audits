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

	// Handle SFTP operations first
	if *sftpFile != "" && *sshutil.Host != "" {
		// If no inventory file is specified, try direct connection
		if *sshutil.File == "" {
			if err := sftp.ExecuteCommandOnHostDirect(*sftpFile, *sshutil.Host); err != nil {
				log.Fatal(err)
			}
			return
		}
		// Use inventory file if specified
		if err := sftp.ExecuteCommandOnHost(*sftpFile, *sshutil.File, *sshutil.Host); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Handle regular SSH operations
	if *sshutil.Host != "" && *sshutil.File == "" {
		if err := sshutil.ConnectToDirect(*sshutil.Host); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *sshutil.File == "" {
		log.Fatal("Inventory file (-i) is required for non-direct connections")
	}

	if err := sshutil.Run(); err != nil {
		log.Fatal(err)
	}
}
