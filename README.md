# ssh-audits

## Usage

```GO

To install the package:
go get -u github.com/ChaosHour/ssh-audits

or download the source code:
git clone git@github.com:ChaosHour/ssh-audits.git

What is the best way to run the package?
Replace the ip or fqdn from the hosts.txt and commands from the commands.txt file and run it.

You can use multiple hosts and multiple commands. No spaces in either file.

Did I really need to create this? 
No, not really, but it was fun and I wanted to learn how to do it.

```

```GO
MacBook-Pro:ssh-audits klarsen$ go run main.go
Connecting as user:  klarsen
[+] Connected to 10.x.x.x
[+] Executing pwd; hostname
/Users/klarsen
Mac-Book-Pro2.local

[+] Executing df -HlP
Filesystem     512-blocks       Used Available Capacity  Mounted on
/dev/disk1s5s1 1953595632   30652240 630075072     5%    /
/dev/disk1s4   1953595632         40 630075072     1%    /System/Volumes/VM
/dev/disk1s2   1953595632     684184 630075072     1%    /System/Volumes/Preboot
/dev/disk1s6   1953595632       2248 630075072     1%    /System/Volumes/Update
/dev/disk1s1   1953595632 1289631168 630075072    68%    /System/Volumes/Data

[+] Executing for i in en{0..4}; do echo ${i}; ifconfig ${i} | egrep 'media|status'; done
en0
        media: autoselect
        status: inactive
en1
        media: autoselect <full-duplex>
        status: inactive
en2
        media: autoselect <full-duplex>
        status: inactive
en3
        media: autoselect <full-duplex>
        status: inactive
en4
        media: autoselect <full-duplex>
        status: inactive
```

## The main import used for ssh is the ssh package

[GOPH](https://github.com/melbahja/goph) - github.com/melbahja/goph

I used the ssh package to connect to the remote host and execute commands.

```GO
SSH Agent was used to connect to the remote hosts. Example provided by the ssh package.

☛ Start Connection With SSH Agent (Unix systems only):
auth, err := goph.UseAgent()
if err != nil {
       // handle error
}

client, err := goph.New("root", "192.1.1.3", auth)
```

```GO

To compile this for FreeBSD:

FreeBSD:
env GOOS=freebsd GOARCH=amd64 go build .

On Mac:
env GOOS=darwin GOARCH=amd64 go build .

```

### Thank you! [Github Copilot](https://copilot.github.com/)
