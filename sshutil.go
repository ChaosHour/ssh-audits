// Parse the Ansible inventory file and return a list of hosts
// go run . -i inventory/hosts -g mysql -l 'tag_web' -c 'ls -lrt'

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ChaosHour/ssh-audits/pkg/sftp"
	"github.com/fatih/color"
	"github.com/melbahja/goph"

	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

/*
// Work on this later:
func addNewHost() {
	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}
*/

func connectToGroup() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// find the group in the inventory
	group, ok := inv.Groups[*group]
	if !ok {
		fmt.Printf("Error: group %s not found in inventory", *group)
		os.Exit(1)
	}

	// create a new color object
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// execute the commands on the hosts in the group
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	for _, host := range group.Hosts {
		// connect to the host using the IP address specified by the ansible_host variable
		client, err := goph.New(strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "), auth)
		if err != nil {
			fmt.Printf("%s Error: failed to connect to host %s:\n %v\n", red("[-]"), host.Name, err)
			continue
		}

		// print the "Connected to host" message
		fmt.Printf("%s Connected to host %s:\n", green("[+]"), host.Name)

		// execute the commands if -c flag is set
		if *command != "" {
			out, err := client.Run(*command)
			if err != nil {
				fmt.Printf("%s Error: failed to execute command on host %s:\n %v\n", red("[-]"), host.Name, err)
				continue
			}

			// print the output
			fmt.Printf("%s\n", out)
		} else {
			// execute the commands from the commands.txt file
			for _, command := range commands {
				// print the command being executed
				fmt.Printf("Executing command: %s\n", command)

				out, err := client.Run(command)
				if err != nil {
					fmt.Printf("%s Error: failed to execute command on host %s:\n %v\n", red("[-]"), host.Name, err)
					continue
				}

				// print the output
				fmt.Printf("%s\n", out)
			}
		}

		// close the connection
		client.Close()
	}
}

func executeCommand() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// find the group in the inventory
	group, ok := inv.Groups[*group]
	if !ok {
		fmt.Printf("Error: group %s not found in inventory", *group)
		os.Exit(1)
	}

	// execute the commands on the hosts in the group
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	for _, host := range group.Hosts {
		// connect to the host using the IP address specified by the ansible_host variable
		client, err := goph.New(strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "), auth)
		if err != nil {
			fmt.Printf("%s Error: failed to connect to host %s:\n %v\n", red("[-]"), host.Name, err)
			continue
		}

		// execute the command if -c flag is set
		if *command != "" {
			out, err := client.Run(*command)
			if err != nil {
				fmt.Printf("%s Error: failed to execute command on host %s:\n %v\n", red("[-]"), host.Name, err)
				continue
			}

			// print the output
			fmt.Printf("%s Connected to host %s:\n%s\n", green("[+]"), host.Name, out)
		}

		// close the connection
		client.Close()
	}
}

func limitToHosts() {
	// TODO: implement limitToHosts function
}

func limitToGroups() {
	// TODO: implement limitToGroups function
}

func connectAsUser() {
	// TODO: implement connectAsUser function
}

func excludeHosts() {
	// TODO: implement excludeHosts function
}

func excludeGroups() {
	// TODO: implement excludeGroups function
}

func Vars() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// print the variables for each host
	for _, host := range inv.Hosts {
		color.Green("%s", host.Name)
		for k, v := range host.Vars {
			color.Yellow("%s : %s	", k, v)
		}
	}
}

func utils() {
	flag.Parse()

	switch {

	case *sftpFile != "":
		sftp.ExecuteCommandOnHost(*sftpFile, *file, *host)
	case *host != "":
		connectToHost()
	case *group != "":
		connectToGroup()
	case *command != "":
		executeCommand()
	case *limit != "":
		limitToHosts()
	case *limitGroup != "":
		limitToGroups()
	case *sshUser != "":
		connectAsUser()
	case *listHosts:
		listHosts2()
	case *listGroups:
		listGroups2()
	case *excludeHost != "":
		excludeHosts()
	case *excludeGroup != "":
		excludeGroups()
	case *showVars:
		Vars()
	default:
		fmt.Println("Usage: ssh-audits [-i inventory-file] [-h host] [-g group] [-c command] [-l limit] [-lg limit-group] [-u ssh-user] [-hosts] [-groups] [-e exclude-host] [-eg exclude-group] [-vars] [-sftp sftp-file]")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
