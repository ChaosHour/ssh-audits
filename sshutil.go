// Parse the Ansible inventory file and return a list of hosts
// go run . -i inventory/hosts -g mysql -l 'tag_web' -c 'ls -lrt'

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/melbahja/goph"

	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// define the command line flags with subcommands.
var (
	file         = flag.String("i", "", "Ansible inventory file")
	host         = flag.String("h", "", "Host to connect to")
	excludeHost  = flag.String("e", "", "Host to exclude")
	group        = flag.String("g", "", "Group to connect to")
	excludeGroup = flag.String("eg", "", "Group to exclude")
	command      = flag.String("c", "", "Command to execute")
	limit        = flag.String("l", "", "Limit to hosts")
	limitGroup   = flag.String("lg", "", "Limit to groups")
	sshUser      = flag.String("u", "", "User to connect as")
	listHosts    = flag.Bool("hosts", false, "List hosts")
	listGroups   = flag.Bool("groups", false, "List groups")
	showVars     = flag.Bool("vars", false, "show host vars from inventory file")
)

//var (
// err    error
// auth   goph.Auth
// client *goph.Client
// addr   string
// user   string
// port uint
// key  string
//)

type Group struct {
	Name     string
	Vars     map[string]string
	Hosts    map[string]*Host
	Children map[string]*Group
	Parents  map[string]*Group

	// Has unexported fields.
}

type Host struct {
	Name   string
	Port   int
	Vars   map[string]string
	Groups map[string]*Group

	// Has unexported fields.
}

type InventoryData struct {
	Groups map[string]*Group
	Hosts  map[string]*Host
}

func listHosts2() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// print the list of hosts
	for _, host := range inv.Hosts {
		fmt.Println(host.Name)
	}
}

func listGroups2() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// print the list of groups
	for _, group := range inv.Groups {
		fmt.Println(group.Name)
	}
}

func connectToHost() {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// find the host in the inventory
	h, ok := inv.Hosts[*host]
	if !ok {
		fmt.Printf("Error: host %s not found in inventory", *host)
		os.Exit(1)
	}

	// execute the command on the specified host
	auth, err := goph.Key(strings.TrimPrefix(h.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
	if err != nil {
		log.Fatal(err)
	}

	client, err := goph.New(strings.TrimPrefix(h.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(h.Vars["ansible_host"], "Vars: "), auth)
	if err != nil {
		log.Fatal(err)
	}

	// create a new color object
	green := color.New(color.FgGreen).SprintFunc()

	// print success
	fmt.Printf("%s Connected to %s\n", green("[+]"), *host)

	defer client.Close()

	output, err := client.Run(*command)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

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

	// iterate over the hosts in the group
	// create a new color object
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, host := range group.Hosts {
		// create an SSH client config
		auth, err := goph.Key(strings.TrimPrefix(host.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
		if err != nil {
			log.Fatal(err)
		}

		// connect to the host using the IP address specified by the ansible_host variable
		client, err := goph.New(strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "), host.Vars["ansible_host"], auth)
		if err != nil {
			fmt.Printf("%s Error: failed to connect to host %s:\n %v\n", red("[-]"), host.Name, err)
			continue
		}

		// execute the command
		out, err := client.Run(*command)
		if err != nil {
			fmt.Printf("%s Error: failed to execute command on host %s:\n %v\n", red("[-]"), host.Name, err)
			continue
		}

		// print the output
		fmt.Printf("%s Connected to host %s:\n%s\n", green("[+]"), host.Name, out)

		// close the connection
		client.Close()
	}
}

func executeCommand() {
	// if the command is not specified, read from commands.txt
	if *command == "" {
		// parse the inventory file
		inv, err := aini.ParseFile(*file)
		if err != nil {
			fmt.Printf("Error: %s ", err)
			os.Exit(1)
		}

		// read and execute the commands from the commands.txt file. Use function from main.go
		// find the host in the inventory
		h, ok := inv.Hosts[*host]
		if !ok {
			fmt.Printf("Error: host %s not found in inventory", *host)
			os.Exit(1)
		}

		// execute the commands on the specified host
		auth, err := goph.Key(strings.TrimPrefix(h.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
		if err != nil {
			log.Fatal(err)
		}

		client, err := goph.New(strings.TrimPrefix(h.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(h.Vars["ansible_host"], "Vars: "), auth)
		if err != nil {
			log.Fatal(err)
		}

		defer client.Close()

		for _, command := range commands {
			fmt.Println(green("[+]"), "Executing", command)
			out, err := client.Run(command)
			if err != nil {
				//	log.Fatal(err)
				continue
			}
			fmt.Println(string(out))
		}

	} else {
		// execute the specified command
		executeSingleCommand(*command)
	}
}

func executeSingleCommand(cmd string) {
	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// find the host in the inventory
	h, ok := inv.Hosts[*host]
	if !ok {
		fmt.Printf("Error: host %s not found in inventory", *host)
		os.Exit(1)
	}

	// execute the command on the specified host
	auth, err := goph.Key(strings.TrimPrefix(h.Vars["ansible_ssh_private_key_file"], "Vars: "), "")
	if err != nil {
		log.Fatal(err)
	}

	client, err := goph.New(strings.TrimPrefix(h.Vars["ansible_user"], "Vars: "), strings.TrimPrefix(h.Vars["ansible_host"], "Vars: "), auth)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	output, err := client.Run(cmd)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
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
		fmt.Println("Usage: sshutil [-i inventory-file] [-h host] [-g group] [-c command] [-l limit] [-lg limit-group] [-u ssh-user] [-hosts] [-groups] [-e exclude-host] [-eg exclude-group] [-vars]")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
