Hello World Example
===================

#### Build:

```
docker build -t hello_world_ex .
```

#### Run:

Container:

```
$ docker run --rm --name hello_world -it hello_world_ex
2020/03/07 18:15:28 Starting process: /app/hello_world
Hello World!
Hello World!
Hello World!
Hello World!
2020/03/07 18:15:39 Stopping process...
2020/03/07 18:15:41 Starting process: /app/hello_world
Hello World!
Hello World!
Hello World!
Hello World!
2020/03/07 18:15:45 Stopping process...
```

Exec:

```
$ docker exec -it hello_world /bin/sh
/go #
/go # volleyctl stop
/go # volleyctl start
/go # volleyctl shutdown
/go # %
```
