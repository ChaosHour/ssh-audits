# ssh-audits

## Usage

```GO

To install the package depending on your GO version:
go get -u github.com/ChaosHour/ssh-audits

go install github.com/ChaosHour/ssh-audits@latest

or download the source code:
git clone git@github.com:ChaosHour/ssh-audits.git

What is the best way to run the package?
Replace the ip or fqdn from the hosts.txt and commands from the commands.txt file and run it.

You can use multiple hosts and multiple commands. No spaces in either file.

Did I really need to create this? 
No, not really, but it was fun and I wanted to learn how to do it.

```

```GO
(data-sync) klarsen@Mac-Book-Pro2 ssh-audits % ./ssh-audits -i inventory/hosts 
Usage: ./ssh-audits -i inventory/hosts [subcommand] [flags]
Subcommands: hosts, groups, vars, ssh, limit
Subcommands: hosts[run against all hosts], limit[run against a specific host], ssh[print ssh command to]
Example: ./ssh-audits -i inventory/hosts hosts
Example: ./ssh-audits -i inventory/hosts limit primary
Example: ./ssh-audits -i inventory/hosts ssh
Flags: -i inventory file
Default to using the hosts.txt: ./ssh-audits


Run against a specific host:
klarsen@Mac-Book-Pro2 ssh-audits % ./ssh-audits -i inventory/hosts limit primary
[+] Connected to primary
[+] Executing pwd; hostname
/home/klarsen
primary

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  1.1M  101M   1% /run
/dev/sda1        42G  3.9G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.152




Run against all hosts:
klarsen@Mac-Book-Pro2 ssh-audits % ./ssh-audits -i inventory/hosts hosts
primary
[+] Connected to primary
[+] Executing pwd; hostname
/home/klarsen
primary

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  1.1M  101M   1% /run
/dev/sda1        42G  3.9G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.152

replica
[+] Connected to replica
[+] Executing pwd; hostname
/home/klarsen
replica

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  988k  101M   1% /run
/dev/sda1        42G  3.9G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.153

etlreplica
[+] Connected to etlreplica
[+] Executing pwd; hostname
/home/klarsen
etlreplica

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  1.1M  101M   1% /run
/dev/sda1        42G  3.9G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.154

proxysql
[+] Connected to proxysql
[+] Executing pwd; hostname
/home/klarsen
proxysql

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  988k  101M   1% /run
/dev/sda1        42G  3.8G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.150

orchestrator
[+] Connected to orchestrator
[+] Executing pwd; hostname
/home/klarsen
orchestrator

[+] Executing df -HlP
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           102M  1.1M  101M   1% /run
/dev/sda1        42G  3.9G   38G  10% /
tmpfs           509M     0  509M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  383G  618G  39% /vagrant
tmpfs           102M  4.1k  102M   1% /run/user/1002

[+] Executing cat /proc/cpuinfo | egrep -i 'model name|cpu cores|cache size'
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2
model name	: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
cache size	: 12288 KB
cpu cores	: 2

[+] Executing ip a s enp0s8 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2
10.8.0.151
```

```GO
On Mac:
env GOOS=darwin GOARCH=amd64 go build .
```


## What is New!

I'm re-working the code and adding more features that I think will be useful.





## Connecting to a group of hosts and running chained commands.

```GO
ssh-audits on  fix_ansible_inventory [!] via 🐹 v1.21.3 
❯ go run . -i inventory/hosts -g mysql -c 'pwd; df -HlP'    

[+] Connected to host replica:
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.2G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002

[+] Connected to host etlreplica:
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.2G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002

[+] Connected to host primary:
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.1G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002

```

## Connecting to a host and running chained commands.

```GO
ssh-audits on  fix_ansible_inventory [!] via 🐹 v1.21.3 took 8s 
❯ go run . -i inventory/hosts -h primary -c 'pwd; df -HlP'

[+] Connected to primary
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.1G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002


ssh-audits on  fix_ansible_inventory [!] via 🐹 v1.21.3 
❯ go run . -i inventory/hosts -h replica -c 'pwd; df -HlP'

[+] Connected to replica
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.2G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002


ssh-audits on  fix_ansible_inventory [!] via 🐹 v1.21.3 
❯ go run . -i inventory/hosts -h etlreplica -c 'pwd; df -HlP'

[+] Connected to etlreplica
/home/klarsen
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           101M  1.1M  100M   2% /run
/dev/sda1        42G  4.2G   38G  10% /
tmpfs           502M     0  502M   0% /dev/shm
tmpfs           5.3M     0  5.3M   0% /run/lock
vagrant         1.1T  480G  521G  48% /vagrant
tmpfs           101M  4.1k  101M   1% /run/user/1002

```


### Thank you! [Github Copilot](https://copilot.github.com/)
