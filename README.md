# ssh-audits

### Usage
```GO
MacBook-Pro:ssh-audits klarsen$ go run main.go
Connecting as user:  klarsen
[+] Connected to 10.8.0.11
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

#####  
The main import used 
github.com/melbahja/goph 

```GO
This was compiled for FreeBSD, but you can compile for the OS you wish.

FreeBSD:
env GOOS=freebsd GOARCH=amd64 go build .
```