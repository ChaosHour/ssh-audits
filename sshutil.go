// Parse the Ansible inventory file and return a list of hosts

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/relex/aini"
)

// define the command line flags with subcommands.
//var (
//	file = flag.String("i", "", "Ansible inventory file")
//)

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
				color.Yellow("ssh -i %s -p %s %s@%s \n", strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_port"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "))
			}
		}

	default:
		fmt.Println("Usage: go run main.go [subcommand] [flags]")
		fmt.Println("Subcommands: hosts, groups, vars, ssh, limit")
		fmt.Println("Flags: -i inventory file")

	}
}
