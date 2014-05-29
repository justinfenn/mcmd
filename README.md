# mcmd

Mcmd is a simple utility to run commands on predefined sets of hosts.  It's a lot like `ssh user@host <command>`, but it works on multiple targets at the same time.

## Usage

Download and install:

```bash
go get github.com/justinfenn/mcmd
```

Define a host file and run:

```bash
mcmd hosts.yaml uptime

mcmd hosts.yaml tail /var/log/my-app/out.log
```

Mcmd will first try to treat the host file parameter as a path to the file.  If no file is found at the path, it will look in `$XDG_CONFIG_HOME/mcmd` or `$HOME/.config/mcmd`.  Don't specify the extension in the parameter if the host file lives in one of the config directories.

```bash
mcmd prod-servers df -h
```

## Host files

Connection information is specified in host files.  In addition to the list of hosts, you can also set the username and authentication method.

Use password auth:

```yaml
user: my-username
auth:
  password: true
hosts:
  - host1:22
  - host2:22
```

Specify private key:

```yaml
user: my-username
auth:
  privatekey: $HOME/.ssh/my_key_rsa
hosts:
  - host1:22
  - host2:22
  - host3:22
  - host4:22
```
