// Package cli parses command-line flags and dispatches to the requested
// operation.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/ChaosHour/ssh-audits/internal/sftp"
	"github.com/ChaosHour/ssh-audits/internal/sshutil"
)

// Run is the entry point called by main.
func Run(version string) error {
	var (
		opts        sshutil.Options
		sftpFile    string
		showVersion bool
	)

	flag.StringVar(&opts.InventoryFile, "i", "", "Ansible inventory file")
	flag.StringVar(&opts.Host, "host", "", "Host to connect to")
	flag.StringVar(&opts.Group, "g", "", "Group to connect to")
	flag.StringVar(&opts.Command, "c", "", "Command to execute")
	flag.BoolVar(&opts.ListHosts, "hosts", false, "List hosts from the inventory")
	flag.BoolVar(&opts.ListGroups, "groups", false, "List groups from the inventory")
	flag.BoolVar(&opts.ShowVars, "vars", false, "Show host vars from the inventory")
	flag.StringVar(&opts.Limit, "l", "", "Limit execution to specified hosts (comma-separated)")
	flag.StringVar(&opts.LimitGroup, "lg", "", "Limit execution to specified groups (comma-separated)")
	flag.StringVar(&opts.CommandsFile, "f", "commands.txt", "Path to commands file")
	flag.DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Per-command execution timeout")
	flag.IntVar(&opts.Port, "p", 22, "SSH port number")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "Print target hosts and commands without connecting")
	flag.IntVar(&opts.Parallel, "parallel", 1, "Maximum hosts to run concurrently (default is sequential, streaming output live)")
	flag.BoolVar(&opts.Stream, "stream", false, "Stream output live, prefixing each line with [host] (runs all hosts at once; -timeout and -parallel are ignored)")
	flag.StringVar(&sftpFile, "sftp", "", "Local script to upload and execute on the host")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println("ssh-audits", version)
		return nil
	case sftpFile != "":
		if opts.Host == "" {
			return errors.New("-sftp requires -host")
		}
		return runSFTP(sftpFile, opts)
	case opts.Host != "" && opts.InventoryFile == "":
		return sshutil.ConnectToDirect(opts.Host, opts)
	case opts.InventoryFile == "":
		return errors.New("inventory file (-i) is required for non-direct connections")
	default:
		return sshutil.Run(opts)
	}
}

// runSFTP uploads a local script to opts.Host and executes it.
func runSFTP(localPath string, opts sshutil.Options) error {
	if opts.DryRun {
		fmt.Printf("Dry run - would upload and execute %s on %s\n", localPath, opts.Host)
		return nil
	}

	client, err := sshutil.ConnectTarget(opts)
	if err != nil {
		return err
	}
	defer client.Close()

	return sftp.UploadFileAndExecute(client, localPath)
}
