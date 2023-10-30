# ssh-audits

Work is in progress...

## What is it?
So, in the days when I was a Sysadmin I would write a bunch of shell scripts and loop through servers from a list and do stuff.

This is kind of like that now, but with Go and SSH.

Why not just use Ansible or JetPorch from the same creator as Ansible, but it's in Rust?
Good question, I like Ansible and JetPorch, but I wanted to learn how to do this in Go.

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

## Added sftp support

```GO
❯ go run . -i inventory/hosts -h primary -sftp './my-thing.sh'                                      
Uploading file ./my-thing.sh to /tmp/my-thing.sh
File uploaded successfully
Making file /tmp/my-thing.sh executable
File made executable successfully
Executing file /tmp/my-thing.sh




sh-audits on  adding_scp:main [!?] via 🐹 v1.21.3 
❯ go run . -i inventory/hosts -h primary -c 'sudo -H bash -c /tmp/my-thing.sh'  
[+] Connected to primary
+----------------+----------------------+----------+
| host_short     | users                | COUNT(*) |
+----------------+----------------------+----------+
| 192.168.50.153 | repl                 |        1 |
| 192.168.50.154 | repl                 |        1 |
| localhost      | event_scheduler,root |        2 |
| total          |                      |        4 |
+----------------+----------------------+----------+
```


```GO
Install the package:
go install github.com/ChaosHour/ssh-audits@latest

or download the source code:
git clone git@github.com:ChaosHour/ssh-audits.git


```

```GO
On Mac:
env GOOS=darwin GOARCH=amd64 go build .
```



### Thank you! [Github Copilot](https://copilot.github.com/)
