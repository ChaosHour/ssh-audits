# ssh-audits

Run commands on one host or a group of hosts over SSH, driven by an
Ansible-style inventory file.

## What is it?

So, in the days when I was a Sysadmin I would write a bunch of shell scripts
and loop through servers from a list and do stuff.

This is kind of like that now, but with Go and SSH.

Why not just use Ansible or JetPorch from the same creator as Ansible, but
it's in Rust? Good question, I like Ansible and JetPorch, but I wanted to
learn how to do this in Go.

## Install

```sh
go install github.com/ChaosHour/ssh-audits/cmd/ssh-audits@latest

# or from source:
git clone git@github.com:ChaosHour/ssh-audits.git
cd ssh-audits
make build   # binary lands in ./bin/ssh-audits
```

Authentication uses your local SSH agent (`ssh-add -l` should list a key).
Unknown host keys are added to `~/.ssh/known_hosts` after printing their
fingerprint; a changed host key is refused.

## Flags

| Flag | Description |
|---|---|
| `-i <file>` | Ansible inventory file |
| `-host <name>` | Single host to connect to (inventory name, or a direct address without `-i`) |
| `-g <group>` | Group to connect to |
| `-l <hosts>` | Limit execution to specific hosts (comma-separated) |
| `-lg <groups>` | Limit execution to specific groups (comma-separated) |
| `-c <command>` | Command to execute (chain with `;`) |
| `-f <file>` | Commands file, one command per line (default `commands.txt`; `#` comments and blank lines are skipped) |
| `-sftp <script>` | Upload a local script to the host and execute it (requires `-host`) |
| `-hosts` / `-groups` / `-vars` | List hosts, groups, or host vars from the inventory |
| `-p <port>` | SSH port (default 22; a per-host `ansible_port` in the inventory overrides it) |
| `-timeout <dur>` | Per-command timeout (default 30s) |
| `-parallel <n>` | Maximum hosts to run concurrently (default 1: sequential, streaming output live) |
| `-stream` | Stream output live, prefixing each line with `[host]`; runs all hosts at once (for `tail -f`-style commands; `-timeout` and `-parallel` are ignored) |
| `-dry-run` | Print the target hosts and commands without connecting |
| `-version` | Print version and exit |

> Note: `-h` used to be the host flag; it now prints this help. Use `-host`.

If a host or group name doesn't exist in the inventory, ssh-audits errors out
instead of running anywhere. When any command or connection fails, the exit
code is non-zero.

By default hosts run one at a time, in alphabetical order, streaming output
live. Pass `-parallel <n>` to run up to n hosts concurrently — each host's
output is then buffered and printed as one block when that host finishes, in
completion order:

```console
$ ssh-audits -i inventory/hosts -g mysql -c 'uptime' -parallel 10
```

## Streaming: follow logs on many hosts

`-stream` switches to line-by-line output: every host runs at once and each
output line is printed the moment it arrives, prefixed with the host name.
This is what you want for long-running commands like `tail -f`:

```console
$ ssh-audits -i inventory/hosts -g mysql -c 'sudo tail -f /var/log/mysql/error.log' -stream
[primary]    2026-07-11T22:14:03 [Warning] Aborted connection 1042 to db: 'app'
[replica]    2026-07-11T22:14:05 [Note] Replica I/O thread: connected to source
[etlreplica] 2026-07-11T22:14:09 [Note] InnoDB: Buffer pool(s) load completed
```

Host prefixes are padded to line up, lines never interleave mid-line, and
Ctrl-C stops all hosts cleanly. In stream mode `-timeout` and `-parallel`
are ignored — commands run until they exit or you interrupt them.

## Inventory

Standard Ansible INI format (see `inventory/hosts` for a fuller example).
`ansible_host`, `ansible_user`, and `ansible_port` are honored per host —
`ansible_port` makes port-mapped Docker containers work as inventory hosts:

```ini
[mysql]
primary  ansible_host=192.168.64.10 ansible_user=ubuntu
replica  ansible_host=192.168.64.11 ansible_user=ubuntu

[docker]
mysql1   ansible_host=127.0.0.1 ansible_user=root ansible_port=2201
mysql2   ansible_host=127.0.0.1 ansible_user=root ansible_port=2202
```

## Connecting to a group of hosts and running chained commands

```console
$ ssh-audits -i inventory/hosts -g mysql -c 'pwd; df -HlP'

[+] Executing command on primary: pwd; df -HlP
/home/vagrant
Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        42G  4.1G   38G  10% /
...
[+] Executing command on replica: pwd; df -HlP
...
```

Not sure what a run will touch? Preview it first:

```console
$ ssh-audits -i inventory/hosts -g mysql -dry-run
[*] Dry run - would execute on 3 host(s):
  etlreplica
  primary
  replica
Commands:
  uname -a
  uptime
  df -h /
```

## Connecting to a single host

```console
$ ssh-audits -i inventory/hosts -host primary -c 'pwd; df -HlP'
```

Or without an inventory (connects as your local user):

```console
$ ssh-audits -host db1.example.com -c 'uptime'
```

## Limiting a run

```console
$ ssh-audits -i inventory/hosts -l primary,replica -c 'uptime'   # specific hosts
$ ssh-audits -i inventory/hosts -lg mysql,control -c 'uptime'    # specific groups
```

## Running a commands file

Copy `commands.txt.example` to `commands.txt` (or use `-f`):

```console
$ ssh-audits -i inventory/hosts -g mysql -f commands.txt
```

## SFTP: upload a script and execute it

Uploads the script to a private temp directory on the host (created with
`mktemp -d`), runs it, prints the output, and cleans up:

```console
$ ssh-audits -i inventory/hosts -host primary -sftp ./my-thing.sh
Uploading ./my-thing.sh to /tmp/ssh-audits.a1B2c3/my-thing.sh
Executing /tmp/ssh-audits.a1B2c3/my-thing.sh
+----------------+----------------------+----------+
| host_short     | users                | COUNT(*) |
+----------------+----------------------+----------+
| localhost      | event_scheduler,root |        2 |
| total          |                      |        2 |
+----------------+----------------------+----------+
```

## Development

```sh
make build   # build, binary lands in ./bin/ssh-audits
make test    # run unit tests
make vet     # static checks
make clean   # remove ./bin
```

Layout: `cmd/ssh-audits` is the entry point; the logic lives in
`internal/cli` (flags and dispatch), `internal/sshutil` (inventory, connect,
execute), `internal/sftp` (upload and run scripts), and `internal/hostkeys`
(host-key verification).

### Thank you! [Github Copilot](https://copilot.github.com/)
