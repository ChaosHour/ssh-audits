// Parse the Ansible inventory file and return a list of hosts

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/relex/aini"

	"github.com/melbahja/goph"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// define the command line flags with subcommands.
//var (
//	file = flag.String("i", "", "Ansible inventory file")
//)

var (
	err    error
	auth   goph.Auth
	client *goph.Client
	addr   string
	//user   string
	port uint
	key  string
)

func utils() {
	flag.Parse()

	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// use a switch statement to handle the subcommands and flags passed to the program
	switch flag.Arg(0) {
	case "hosts":
		for _, host := range inv.Hosts {
			color.Green("%s", host.Name)
			// Uing the output of the case limit subcommand, we can use the output to ssh into the host using the goph library
			auth, err := goph.Key(strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
			if err != nil {
				log.Fatal(err)
			}

			client, err := goph.New(strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "), auth)
			if err != nil {
				log.Fatal(err)
			}

			// Defer closing the network connection.
			defer client.Close()

			// print success
			fmt.Println(green("[+]"), "Connected to", host)

			// Execute your command.
			for _, command := range commands {
				fmt.Println(green("[+]"), "Executing", command)
				out, err := client.Run(command)
				if err != nil {
					//	log.Fatal(err)
					continue
				}

				// close the connection
				defer client.Close()

				// Get your output as []byte.
				fmt.Println(string(out))
			}

		}
	case "groups":
		for _, group := range inv.Groups {
			color.Green("%s", group.Name)
		}
	case "vars":
		for _, host := range inv.Hosts {
			color.Green("%s", host.Name)
			for k, v := range host.Vars {
				color.Yellow("%s : %s	", k, v)
			}
		}
	case "ssh":
		for _, host := range inv.Hosts {
			//color.Green("%s", host.Name)
			color.Yellow("ssh -i %s -p %s %s@%s \n", strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_port"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "))
		}
	// case ssh with limit to a specific host
	case "limit":
		for _, host := range inv.Hosts {
			if host.Name == flag.Arg(1) {
				//color.Yellow("ssh -i %s -p %s %s@%s \n", strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_port"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "))
				// Uing the output of the case limit subcommand, we can use the output to ssh into the host using the goph library
				auth, err := goph.Key(strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
				if err != nil {
					log.Fatal(err)
				}

				client, err := goph.New(strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "), auth)
				if err != nil {
					log.Fatal(err)
				}

				// Defer closing the network connection.
				defer client.Close()

				// print success
				fmt.Println(green("[+]"), "Connected to", host)

				// Execute your command.
				for _, command := range commands {
					fmt.Println(green("[+]"), "Executing", command)
					out, err := client.Run(command)
					if err != nil {
						//	log.Fatal(err)
						continue
					}

					// close the connection
					defer client.Close()

					// Get your output as []byte.
					fmt.Println(string(out))
				}

			}
		}
	default:
		color.Green("Usage: go run main.go [subcommand] [flags]")
		color.Green("Subcommands: hosts, groups, vars, ssh, limit")
		// add color to the output for each subcommand
		color.Red("Subcommands: hosts[run against all hosts], limit[run against a specific host], ssh[print ssh command to]")
		// add an example of how to use the program
		color.Yellow("Example: go run . -i inventory/hosts hosts")
		color.Yellow("Example: go run . -i inventory/hosts limit primary")
		color.Yellow("Example: go run . -i inventory/hosts ssh")
		//fmt.Println("Subcommands: hosts)[run against all hosts], limit[run against a specific host], ssh[print ssh command to]")
		color.Green("Flags: -i inventory file")
		color.Green("Default to using the hosts.txt: go run .")

	}
}
