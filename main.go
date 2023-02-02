package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"github.com/mikkeloscar/sshconfig"
	_ "golang.org/x/crypto/ssh"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// define colors
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// read the hosts file in map
func readHosts() map[string]string {
	// open file
	file, err := os.Open("hosts.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// defer file close
	defer file.Close()
	// read file
	hosts := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hosts[scanner.Text()] = scanner.Text()
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

// main function
func main() {
	// get user home directory
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dirname)

	// parse ssh config file for user
	hosts, err := sshconfig.Parse(dirname + "/.ssh/config")
	if err != nil {
		fmt.Println(err)
	}

	// use ssh agent
	auth, err := goph.UseAgent()
	if err != nil {
		log.Fatal(err)
	}

	validHosts := readHosts()

	//  Start Connection With SSH Agent using goph
	for _, host := range hosts {
		fmt.Printf("%+v\n", host)

		shouldConnect := false
		if _, ok := validHosts[host.HostName]; !ok {
			for _, h := range host.Host {
				if _, ok := validHosts[h]; ok {
					shouldConnect = true
					break
				}
			}
		}
		if !shouldConnect {
			fmt.Println(red("[!]"), "Skipping...")
			continue
		}
		fmt.Println(green(fmt.Sprintf("Connecting to Host=%v as %v@%v", host.Host, host.User, host.HostName)))

		client, err := goph.New(host.User, host.HostName, auth)
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
