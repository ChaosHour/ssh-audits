// Package sshutil connects to hosts from an Ansible-style inventory and runs
// commands on them over SSH.
package sshutil

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"github.com/relex/aini"
	"golang.org/x/sync/errgroup"

	"github.com/ChaosHour/ssh-audits/internal/hostkeys"
)

// Options holds the CLI configuration for a run.
type Options struct {
	InventoryFile string
	Host          string
	Group         string
	Command       string
	ListHosts     bool
	ListGroups    bool
	ShowVars      bool
	Limit         string
	LimitGroup    string
	CommandsFile  string
	Timeout       time.Duration
	Port          int
	DryRun        bool
	Parallel      int
	Stream        bool
}

// Colors
var (
	green = color.New(color.FgGreen).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

// Run parses the inventory and dispatches the requested operation.
func Run(opts Options) error {
	inv, err := aini.ParseFile(opts.InventoryFile)
	if err != nil {
		return fmt.Errorf("error parsing inventory: %w", err)
	}

	switch {
	case opts.ListHosts:
		return ListHosts(inv)
	case opts.ListGroups:
		return ListGroups(inv)
	case opts.ShowVars:
		return ShowInventoryVars(inv)
	case opts.Host != "":
		return ConnectToHost(inv, opts)
	case opts.Group != "" || opts.Limit != "" || opts.LimitGroup != "":
		return ConnectToGroup(inv, opts)
	default:
		return errors.New("no valid operation specified")
	}
}

// ListHosts lists all hosts from inventory.
func ListHosts(inv *aini.InventoryData) error {
	for _, name := range slices.Sorted(maps.Keys(inv.Hosts)) {
		fmt.Println(name)
	}
	return nil
}

// ListGroups lists all groups from inventory.
func ListGroups(inv *aini.InventoryData) error {
	for _, name := range slices.Sorted(maps.Keys(inv.Groups)) {
		fmt.Println(name)
	}
	return nil
}

// ShowInventoryVars prints the variables for each host.
func ShowInventoryVars(inv *aini.InventoryData) error {
	for _, name := range slices.Sorted(maps.Keys(inv.Hosts)) {
		color.Green("%s", name)
		for _, k := range slices.Sorted(maps.Keys(inv.Hosts[name].Vars)) {
			color.Yellow("  %s : %s", k, inv.Hosts[name].Vars[k])
		}
	}
	return nil
}

// filterHosts selects the target hosts from -g, -l, and -lg. Unknown host or
// group names are errors so a typo can never widen the target set.
func filterHosts(inv *aini.InventoryData, opts Options) (map[string]*aini.Host, error) {
	if opts.Limit != "" {
		filtered := make(map[string]*aini.Host)
		for _, name := range splitList(opts.Limit) {
			host, ok := inv.Hosts[name]
			if !ok {
				return nil, fmt.Errorf("host %q not found in inventory", name)
			}
			filtered[name] = host
		}
		return filtered, nil
	}

	groups := append(splitList(opts.Group), splitList(opts.LimitGroup)...)
	if len(groups) > 0 {
		filtered := make(map[string]*aini.Host)
		for _, name := range groups {
			group, ok := inv.Groups[name]
			if !ok {
				return nil, fmt.Errorf("group %q not found in inventory", name)
			}
			maps.Copy(filtered, group.Hosts)
		}
		return filtered, nil
	}

	return inv.Hosts, nil
}

// splitList splits a comma-separated flag value, trimming whitespace and
// dropping empty items.
func splitList(s string) []string {
	var out []string
	for item := range strings.SplitSeq(s, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// hostUser returns ansible_user, falling back to the local user.
func hostUser(h *aini.Host) string {
	if u := h.Vars["ansible_user"]; u != "" {
		return u
	}
	return os.Getenv("USER")
}

// hostAddr returns ansible_host, falling back to the inventory hostname.
func hostAddr(h *aini.Host) string {
	if a := h.Vars["ansible_host"]; a != "" {
		return a
	}
	return h.Name
}

// hostPort returns ansible_port, falling back to the -p flag.
func hostPort(h *aini.Host, fallback int) (int, error) {
	p := h.Vars["ansible_port"]
	if p == "" {
		return fallback, nil
	}
	port, err := strconv.Atoi(p)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("host %s: invalid ansible_port %q", h.Name, p)
	}
	return port, nil
}

// dialHost resolves the host's user, address, and port from its inventory
// vars and opens the connection.
func dialHost(h *aini.Host, opts Options) (*goph.Client, error) {
	port, err := hostPort(h, opts.Port)
	if err != nil {
		return nil, err
	}
	return dial(hostUser(h), hostAddr(h), port)
}

// dial opens an SSH connection authenticated via the local SSH agent.
func dial(user, addr string, port int) (*goph.Client, error) {
	auth, err := goph.UseAgent()
	if err != nil {
		return nil, fmt.Errorf("error using SSH agent: %w", err)
	}

	return goph.NewConn(&goph.Config{
		Auth:     auth,
		User:     user,
		Addr:     addr,
		Port:     uint(port),
		Callback: hostkeys.Verify,
	})
}

// ConnectTarget connects to opts.Host, resolving user and address through the
// inventory when one was given.
func ConnectTarget(opts Options) (*goph.Client, error) {
	if opts.InventoryFile == "" {
		return dial(os.Getenv("USER"), opts.Host, opts.Port)
	}

	inv, err := aini.ParseFile(opts.InventoryFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing inventory: %w", err)
	}
	h, ok := inv.Hosts[opts.Host]
	if !ok {
		return nil, fmt.Errorf("host %q not found in inventory", opts.Host)
	}
	return dialHost(h, opts)
}

// ConnectToHost runs the commands on a single inventory host.
func ConnectToHost(inv *aini.InventoryData, opts Options) error {
	h, ok := inv.Hosts[opts.Host]
	if !ok {
		return fmt.Errorf("host %q not found in inventory", opts.Host)
	}
	return runOnHosts(map[string]*aini.Host{h.Name: h}, opts)
}

// ConnectToGroup runs the commands on the hosts selected by -g, -l, or -lg.
func ConnectToGroup(inv *aini.InventoryData, opts Options) error {
	hosts, err := filterHosts(inv, opts)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		return errors.New("no hosts matched the given filters")
	}
	return runOnHosts(hosts, opts)
}

// ConnectToDirect connects to a host without an inventory, as the local
// user. The synthetic inventory host makes hostUser/hostAddr fall back to
// $USER and the hostname.
func ConnectToDirect(hostname string, opts Options) error {
	return runOnHosts(map[string]*aini.Host{hostname: {Name: hostname}}, opts)
}

// runOnHosts executes the commands on each host, at most opts.Parallel hosts
// at a time. Hosts are all attempted even when earlier ones fail; failures
// are joined into the returned error so the process exits non-zero.
func runOnHosts(hosts map[string]*aini.Host, opts Options) error {
	commands, err := GetCommands(opts)
	if err != nil {
		return err
	}

	names := slices.Sorted(maps.Keys(hosts))
	if opts.DryRun {
		printPlan(names, commands)
		return nil
	}

	if opts.Stream {
		return streamOnHosts(hosts, commands, opts)
	}

	// Sequential runs stream output live; parallel runs buffer each host's
	// output and print it as one block when that host finishes, so hosts
	// never interleave mid-command.
	parallel := max(opts.Parallel, 1)
	buffered := parallel > 1

	var (
		g    errgroup.Group
		mu   sync.Mutex
		errs = make([]error, len(names))
	)
	g.SetLimit(parallel)

	for i, name := range names {
		g.Go(func() error {
			var out io.Writer = os.Stdout
			var buf bytes.Buffer
			if buffered {
				out = &buf
			}

			errs[i] = runOnHost(out, hosts[name], name, commands, opts)

			if buffered {
				mu.Lock()
				os.Stdout.Write(buf.Bytes())
				mu.Unlock()
			}
			return nil
		})
	}
	g.Wait()

	return errors.Join(errs...)
}

// runOnHost connects to a single host and runs the commands, writing all
// output to w.
func runOnHost(w io.Writer, h *aini.Host, name string, commands []string, opts Options) error {
	client, err := dialHost(h, opts)
	if err != nil {
		fmt.Fprintf(w, "%s Error: failed to connect to host %s:\n %v\n", red("[-]"), name, err)
		return fmt.Errorf("%s: %w", name, err)
	}
	defer client.Close()

	if err := ExecuteCommands(w, client, name, commands, opts.Timeout); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// ExecuteCommands runs each command on the client with a per-command timeout,
// writing output to w. All commands are attempted; failures are joined into
// one error.
func ExecuteCommands(w io.Writer, client *goph.Client, hostName string, commands []string, timeout time.Duration) error {
	var errs []error
	for _, cmd := range commands {
		fmt.Fprintf(w, "%s Executing command on %s: %s\n", green("[+]"), hostName, cmd)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		out, err := client.RunContext(ctx, cmd)
		cancel()
		if err != nil {
			fmt.Fprintf(w, "%s Error: %v\n", red("[-]"), err)
			errs = append(errs, fmt.Errorf("%q: %w", cmd, err))
			continue
		}
		fmt.Fprintf(w, "%s\n", out)
	}
	return errors.Join(errs...)
}

// GetCommands returns the -c command if given, otherwise the commands file.
func GetCommands(opts Options) ([]string, error) {
	if opts.Command != "" {
		return []string{opts.Command}, nil
	}
	return readCommandsFile(opts.CommandsFile)
}

// readCommandsFile reads one command per line, skipping blank lines and
// #-comments.
func readCommandsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening commands file: %w", err)
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		commands = append(commands, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading commands file: %w", err)
	}
	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands found in %s", path)
	}
	return commands, nil
}

// printPlan shows what a dry run would do.
func printPlan(hosts, commands []string) {
	fmt.Printf("%s Dry run - would execute on %d host(s):\n", green("[*]"), len(hosts))
	for _, h := range hosts {
		fmt.Println("  " + h)
	}
	fmt.Println("Commands:")
	for _, c := range commands {
		fmt.Println("  " + c)
	}
}
