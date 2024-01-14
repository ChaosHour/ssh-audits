package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	_ "github.com/ChaosHour/ssh-audits/pkg/sftp"
	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// define colors
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

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
	sftpFile     = flag.String("sftp", "", "File to SFTP")
	//addPublicKey = flag.Bool("addkey", false, "Add public key to server")
	listHosts  = flag.Bool("hosts", false, "List hosts")
	listGroups = flag.Bool("groups", false, "List groups")
	showVars   = flag.Bool("vars", false, "show host vars from inventory file")
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
	auth, err := goph.UseAgent()
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

	// if the command is not specified, read from commands.txt or -c flag
	var commands []string
	if *command == "" {
		// read commands from commands.txt
		file, err := os.Open("commands.txt")
		if err != nil {
			fmt.Printf("Error: %s ", err)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			commands = append(commands, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error: %s ", err)
			os.Exit(1)
		}
	} else {
		// use command from -c flag
		commands = []string{*command}
	}

	for _, command := range commands {
		output, err := client.Run(command)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(output))
	}
}

// define commands vairable
var commands = readCommands()

// read from commands.txt file to execute commands
func readCommands() []string {
	// open file
	file, err := os.Open("commands.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// defer file close
	defer file.Close()
	// read file
	commands := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	return commands
}

// check for current user who is running this and return the user name
func CurrentUser() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.Username
}

// User represents a user account.
var User = (os.Getenv("USER"))

/*
func Scp2Remote() {
	// Parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}


	// Execute the scp command
	cmd := exec.Command("scp", *scpFile, fmt.Sprintf("%s@%s:/tmp/", h.Vars["ansible_user"], h.Vars["ansible_host"]))
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("File copied successfully")
}

*/

func main() {
	flag.Parse()

	// Check if flag -i is set. If it is set, call the util function to read from the Ansible inventory file.
	if *file != "" {
		utils()
		return
	}

	/*
		if *sftpFile != "" {
			err := sftp.ExecuteCommandOnHost(*file, *host, *sftpFile)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	*/

	// Use SSH Agent to connect to hosts
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	// connecting as user, print user name
	fmt.Println(green("Connecting as user: "), User)

	// parse the inventory file
	inv, err := aini.ParseFile(*file)
	if err != nil {
		fmt.Printf("Error: %s ", err)
		os.Exit(1)
	}

	// loop through the hosts in the inventory file
	for _, host := range inv.Hosts {
		client, err := goph.New(User, host.Name, auth)
		if err != nil {
			fmt.Println(red("[!]"), "Failed to connect to", host.Name)
			// if one connection fails try the next host in the list
			continue
		}

		// Defer closing the network connection.
		defer client.Close()

		// print success
		fmt.Println(green("[+]"), "Connected to", host.Name)

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
