Hello World Example
===================

#### Build:

```
docker build -t bash_ex .
```

#### Run:

Container:

```
$ docker run --rm --name bash_ex -it bash_ex
2020/03/08 01:16:40 Starting process: /bin/sh -c /app/example.sh
Hello, world!
Hello, world!
Hello, world!
Hello, world!
2020/03/08 01:16:58 Stopping process...
2020/03/08 01:17:01 Starting process: /bin/sh -c /app/example.sh
Hello, world!
Hello, world!
Hello, world!
Hello, world!
2020/03/08 01:17:04 Stopping process...
```

Exec:

```
$ docker exec -it bash_ex /bin/sh
/go # volleyctl stop
/go # volleyctl start
/go # volleyctl shutdown
/go # %
```
