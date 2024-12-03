package sshutil

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"

	"github.com/relex/aini"
	_ "golang.org/x/crypto/ssh/agent"
	_ "golang.org/x/crypto/ssh/knownhosts"
)

// Exported variables for flags
var (
	File           = flag.String("i", "", "Ansible inventory file")
	Host           = flag.String("h", "", "Host to connect to")
	Group          = flag.String("g", "", "Group to connect to")
	Command        = flag.String("c", "", "Command to execute")
	ListHostsFlag  = flag.Bool("hosts", false, "List hosts")
	ListGroupsFlag = flag.Bool("groups", false, "List groups")
	ShowVars       = flag.Bool("vars", false, "Show host vars from inventory file")
	Limit          = flag.String("l", "", "Limit execution to specified hosts (comma-separated")
	LimitGroup     = flag.String("lg", "", "Limit execution to specified groups (comma-separated)")
	CommandsFile   = flag.String("f", "commands.txt", "Path to commands file")
	Timeout        = flag.Duration("timeout", 30*time.Second, "Command execution timeout")
)

// Colors
var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

// ListHosts lists all hosts from inventory
func ListHosts(inv *aini.InventoryData) error {
	for _, host := range inv.Hosts {
		fmt.Println(host.Name)
	}
	return nil
}

// ListGroups lists all groups from inventory
func ListGroups(inv *aini.InventoryData) error {
	for _, group := range inv.Groups {
		fmt.Println(group.Name)
	}
	return nil
}

// filterHosts filters hosts based on limit or group settings
func filterHosts(inv *aini.InventoryData) map[string]*aini.Host {
	filteredHosts := make(map[string]*aini.Host)

	if *Limit != "" {
		limitHosts := strings.Split(*Limit, ",")
		for _, h := range limitHosts {
			if host, ok := inv.Hosts[strings.TrimSpace(h)]; ok {
				filteredHosts[h] = host
			}
		}
		return filteredHosts
	}

	if *LimitGroup != "" {
		limitGroups := strings.Split(*LimitGroup, ",")
		for _, g := range limitGroups {
			if group, ok := inv.Groups[strings.TrimSpace(g)]; ok {
				for name, host := range group.Hosts {
					filteredHosts[name] = host
				}
			}
		}
		return filteredHosts
	}

	return inv.Hosts
}

// Run is the main entry point for SSH operations
func Run() error {
	inv, err := aini.ParseFile(*File)
	if err != nil {
		return fmt.Errorf("error parsing inventory: %w", err)
	}

	switch {
	case *ListHostsFlag:
		return ListHosts(inv)
	case *ListGroupsFlag:
		return ListGroups(inv)
	case *ShowVars:
		return ShowInventoryVars(inv)
	case *Host != "":
		return ConnectToHost(inv)
	case *Group != "" || *Limit != "" || *LimitGroup != "":
		return ConnectToGroup(inv)
	default:
		return fmt.Errorf("no valid operation specified")
	}
}

// Export functions (capitalize first letter)
func AddNewHost(hostname string, remote net.Addr, key ssh.PublicKey) error {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Add the new host to known hosts file
	err = goph.AddKnownHost(hostname, remote, key, filepath.Join(home, ".ssh", "known_hosts"))
	if err != nil {
		return fmt.Errorf("failed to add known host: %w", err)
	}

	fmt.Printf("%s Added %s to known hosts\n", green("[+]"), hostname)
	return nil
}

func ConnectToGroup(inv *aini.InventoryData) error {
	hosts := filterHosts(inv)

	auth, err := goph.UseAgent()
	if err != nil {
		return fmt.Errorf("error using SSH agent: %w", err)
	}

	for _, host := range hosts {
		// Create client config with callback
		config := &goph.Config{
			Auth: auth,
			User: strings.TrimPrefix(host.Vars["ansible_user"], "Vars: "),
			Addr: strings.TrimPrefix(host.Vars["ansible_host"], "Vars: "),
			Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return AddNewHost(hostname, remote, key)
			},
		}

		// Create client with config
		client, err := goph.NewConn(config)
		if err != nil {
			fmt.Printf("%s Error: failed to connect to host %s:\n %v\n", red("[-]"), host.Name, err)
			continue
		}

		func() {
			defer client.Close()
			ExecuteCommands(client, host.Name)
		}()
	}

	return nil
}

func ConnectToHost(inv *aini.InventoryData) error {
	h, ok := inv.Hosts[*Host]
	if !ok {
		return fmt.Errorf("host %s not found in inventory", *Host)
	}

	auth, err := goph.UseAgent()
	if err != nil {
		return fmt.Errorf("failed to use SSH agent: %w", err)
	}

	config := &goph.Config{
		Auth: auth,
		User: strings.TrimPrefix(h.Vars["ansible_user"], "Vars: "),
		Addr: strings.TrimPrefix(h.Vars["ansible_host"], "Vars: "),
		Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return AddNewHost(hostname, remote, key)
		},
	}

	client, err := goph.NewConn(config)
	if err != nil {
		return fmt.Errorf("failed to connect to host %s: %w", h.Name, err)
	}
	defer client.Close()

	ExecuteCommands(client, h.Name)
	return nil
}

func ExecuteCommands(client *goph.Client, hostName string) {
	ctx, cancel := context.WithTimeout(context.Background(), *Timeout)
	defer cancel()

	commands := GetCommands()
	for _, cmd := range commands {
		select {
		case <-ctx.Done():
			fmt.Printf("%s Timeout or cancellation on %s\n", red("[-]"), hostName)
			return
		default:
			fmt.Printf("%s Executing command on %s: %s\n", green("[+]"), hostName, cmd)
			out, err := client.Run(cmd)
			if err != nil {
				fmt.Printf("%s Error: %v\n", red("[-]"), err)
				continue
			}
			fmt.Printf("%s\n", out)
		}
	}
}

func GetCommands() []string {
	if *Command != "" {
		return []string{*Command}
	}
	return readCommands()
}

func ShowInventoryVars(inv *aini.InventoryData) error {
	// print the variables for each host
	for _, host := range inv.Hosts {
		color.Green("%s", host.Name)
		for k, v := range host.Vars {
			color.Yellow("%s : %s	", k, v)
		}
	}

	return nil
}

func readCommands() []string {
	file, err := os.Open(*CommandsFile)
	if err != nil {
		log.Printf("Error opening commands file: %v", err)
		return nil
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	return commands
}
