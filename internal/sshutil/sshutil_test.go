package sshutil

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/relex/aini"
)

const testInventory = `[mysql]
primary  ansible_host=10.0.0.1 ansible_user=vagrant
replica  ansible_host=10.0.0.2 ansible_user=vagrant

[control]
proxysql ansible_host=10.0.0.3 ansible_user=vagrant
bare

[docker]
docker1 ansible_host=127.0.0.1 ansible_user=root ansible_port=2201
badport ansible_host=127.0.0.1 ansible_port=nope
`

func parseTestInventory(t *testing.T) *aini.InventoryData {
	t.Helper()
	inv, err := aini.Parse(strings.NewReader(testInventory))
	if err != nil {
		t.Fatalf("parsing test inventory: %v", err)
	}
	return inv
}

func hostNames(hosts map[string]*aini.Host) []string {
	names := make([]string, 0, len(hosts))
	for name := range hosts {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func TestFilterHostsByGroup(t *testing.T) {
	inv := parseTestInventory(t)

	hosts, err := filterHosts(inv, Options{Group: "mysql"})
	if err != nil {
		t.Fatalf("filterHosts: %v", err)
	}
	want := []string{"primary", "replica"}
	if got := hostNames(hosts); !slices.Equal(got, want) {
		t.Errorf("group mysql matched %v, want %v", got, want)
	}
}

func TestFilterHostsUnknownGroupErrors(t *testing.T) {
	inv := parseTestInventory(t)

	if _, err := filterHosts(inv, Options{Group: "nope"}); err == nil {
		t.Fatal("expected error for unknown group, got nil")
	}
}

func TestFilterHostsByLimit(t *testing.T) {
	inv := parseTestInventory(t)

	hosts, err := filterHosts(inv, Options{Limit: "primary, replica"})
	if err != nil {
		t.Fatalf("filterHosts: %v", err)
	}
	want := []string{"primary", "replica"}
	if got := hostNames(hosts); !slices.Equal(got, want) {
		t.Errorf("limit matched %v, want %v", got, want)
	}
}

func TestFilterHostsUnknownHostErrors(t *testing.T) {
	inv := parseTestInventory(t)

	if _, err := filterHosts(inv, Options{Limit: "primary,badhost"}); err == nil {
		t.Fatal("expected error for unknown host, got nil")
	}
}

func TestFilterHostsByLimitGroup(t *testing.T) {
	inv := parseTestInventory(t)

	hosts, err := filterHosts(inv, Options{LimitGroup: "control"})
	if err != nil {
		t.Fatalf("filterHosts: %v", err)
	}
	want := []string{"bare", "proxysql"}
	if got := hostNames(hosts); !slices.Equal(got, want) {
		t.Errorf("limit-group control matched %v, want %v", got, want)
	}
}

func TestFilterHostsNoFilterReturnsAll(t *testing.T) {
	inv := parseTestInventory(t)

	hosts, err := filterHosts(inv, Options{})
	if err != nil {
		t.Fatalf("filterHosts: %v", err)
	}
	if len(hosts) != 6 {
		t.Errorf("no filter matched %d hosts, want 6: %v", len(hosts), hostNames(hosts))
	}
}

func TestFilterHostsLimitTakesPrecedence(t *testing.T) {
	inv := parseTestInventory(t)

	hosts, err := filterHosts(inv, Options{Group: "mysql", Limit: "proxysql"})
	if err != nil {
		t.Fatalf("filterHosts: %v", err)
	}
	want := []string{"proxysql"}
	if got := hostNames(hosts); !slices.Equal(got, want) {
		t.Errorf("limit+group matched %v, want %v", got, want)
	}
}

func TestHostAddr(t *testing.T) {
	inv := parseTestInventory(t)

	if got := hostAddr(inv.Hosts["primary"]); got != "10.0.0.1" {
		t.Errorf("hostAddr(primary) = %q, want 10.0.0.1", got)
	}
	if got := hostAddr(inv.Hosts["bare"]); got != "bare" {
		t.Errorf("hostAddr(bare) = %q, want fallback to hostname", got)
	}
}

func TestHostUser(t *testing.T) {
	inv := parseTestInventory(t)
	t.Setenv("USER", "localuser")

	if got := hostUser(inv.Hosts["primary"]); got != "vagrant" {
		t.Errorf("hostUser(primary) = %q, want vagrant", got)
	}
	if got := hostUser(inv.Hosts["bare"]); got != "localuser" {
		t.Errorf("hostUser(bare) = %q, want fallback to local user", got)
	}
}

func TestHostPort(t *testing.T) {
	inv := parseTestInventory(t)

	if got, err := hostPort(inv.Hosts["docker1"], 22); err != nil || got != 2201 {
		t.Errorf("hostPort(docker1) = %d, %v; want 2201 from ansible_port", got, err)
	}
	if got, err := hostPort(inv.Hosts["primary"], 2222); err != nil || got != 2222 {
		t.Errorf("hostPort(primary) = %d, %v; want fallback 2222", got, err)
	}
	if _, err := hostPort(inv.Hosts["badport"], 22); err == nil {
		t.Error("hostPort(badport) expected error for invalid ansible_port, got nil")
	}
}

func TestSplitList(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{" a , b ,", []string{"a", "b"}},
	}
	for _, tt := range tests {
		if got := splitList(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("splitList(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestReadCommandsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.txt")
	content := "# comment\n\nuptime\n  df -h  \n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	commands, err := readCommandsFile(path)
	if err != nil {
		t.Fatalf("readCommandsFile: %v", err)
	}
	want := []string{"uptime", "df -h"}
	if !slices.Equal(commands, want) {
		t.Errorf("readCommandsFile = %v, want %v", commands, want)
	}
}

func TestReadCommandsFileMissingErrors(t *testing.T) {
	if _, err := readCommandsFile(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReadCommandsFileEmptyErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.txt")
	if err := os.WriteFile(path, []byte("# only comments\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := readCommandsFile(path); err == nil {
		t.Fatal("expected error for file with no commands, got nil")
	}
}

func TestGetCommandsPrefersFlag(t *testing.T) {
	commands, err := GetCommands(Options{Command: "uptime", CommandsFile: "does-not-exist.txt"})
	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}
	if want := []string{"uptime"}; !slices.Equal(commands, want) {
		t.Errorf("GetCommands = %v, want %v", commands, want)
	}
}
