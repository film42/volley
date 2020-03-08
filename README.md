Volley
======

Volley is a docker entrypoint process manager that allows stopping and starting of the managed process without restarting the container. It's a prototype to help test the idea "is there a simple way to hot-patch my apps via a docker container (under kubernetes)?"

### Building

```
$ go build ./cmd/volleyd
$ go build ./cmd/volleyctl
```

### Volleyd

Volleyd is the supervisor that starts and stops your process.

#### Usage

```
$ ./volleyd run --help
Usage:
  volleyd run [command] [flags]

Flags:
  -h, --help                help for run
      --pid-file string     File to write the volleyd pid while running (default "/tmp/volleyd.pid")
```

#### Example

Check the [examples](examples/) directory of this repo for help getting started.

Run a process that doesn't need a bash entrypoint

```
$ ./volleyd run my_binary_process
```

Run a process that needs a bash entrypoint:

```
$ ./volleyd run -- /bin/sh -c "while true; do echo 'hello world'; sleep 1; done"
```

### Volleyctl

Volleyctl is a helper utility that can send start, stop, and shutdown commands to volleyd.

#### Usage

```
$ ./volleyctl start|stop|shutdown
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
