Volley
======

Volley is a docker entrypoint process manager that allows stopping and starting of the managed process without restarting the container. It's a prototype to help test the idea "is there a simple way to hot-patch my apps via a docker container (under kubernetes)?"

### Volleyd

Volleyd is the supervisor that starts and stops your process.

#### Usage

```
$ ./cmd/volleyd/volleyd run --help
Usage:
  volleyd run [command] [flags]

Flags:
  -h, --help              help for run
      --pid-file string   File to write the volleyd pid while running (default "/tmp/volleyd.pid")
```

#### Example

```
$ ./cmd/volleyd/volleyd run
```

### Volleyctl

Volleyctl is a helper utility that can send start, stop, and shutdown commands to volleyd.

#### Usage

```
$ ./cmd/volleyctl/volleyctl start|stop|shutdown
```

#### Example

```
# To tell volleyd to start a process (if it's stopped)
$ volleyctl start

# To tell volleyd to stop a process without exiting (if it's started)
$ volleyctl stop

# To tell volleyd to shutdown (stop and exit)
$ volleyctl shutdown
```

### License

MIT
