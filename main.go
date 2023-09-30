package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/fatih/color"
	"github.com/melbahja/goph"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// define colors
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// define the command line flags with subcommands.
var (
	file = flag.String("i", "", "Ansible inventory file")
)

// define hosts
var hosts = readHosts()

// read hosts from file
func readHosts() []string {
	// open file
	file, err := os.Open("hosts.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// defer file close
	defer file.Close()
	// read file
	hosts := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}
	return hosts
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

func main() {
	flag.Parse()

	// Check if flag -i is set. If it is set, call the util function to read from the Ansible inventory file.
	if *file != "" {
		utils()
		return
	}

	// Use SSH Agent to connect to hosts
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	// connecting as user, print user name
	fmt.Println(green("Connecting as user: "), User)

	//  Start Connection With SSH Agent using goph and looping through hosts in hosts.txt file
	for _, host := range hosts {
		client, err := goph.New(User, host, auth)
		if err != nil {
			fmt.Println(red("[!]"), "Failed to connect to", host)
			//log.Fatal(red(err))
			// if one connection fails try the next host in the list
			continue
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
