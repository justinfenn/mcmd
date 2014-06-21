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

Connection information is specified in host files.  By default, mcmd will attempt to authenticate using an ssh agent.  If an agent with the required key(s) is running, no auth configuration is needed.  You can also specify a private key for mcmd to authenticate with (currently only passphraseless keys are supported).  Mcmd will ask for a password if neither of the above conditions are met.

Use ssh agent or password auth:

```yaml
user: my-username
hosts:
  - host1:22
  - host2:22
```

Specify private key:

```yaml
user: my-username
privatekey: $HOME/.ssh/my_key_rsa
hosts:
  - host1:22
  - host2:22
  - host3:22
  - host4:22
```
