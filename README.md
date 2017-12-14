# A simple todo list app

This simple todo list application is used to demonstrate micro-service development process.
The frontend code is copied from [TodoMVC](http://todomvc.com/) jQuery implementation.

The backend is implemented with a simple [Go](https://golang.org/) application. 
The browser talks to the backend using jQuery AJAX. The backend then stores data in MySQL.

# How to run it

First you need a go development environment. After the code is cloned, you can:

```
$ cd go-todo-demo
$ ./build.sh
```

A docker image with tag `go-todo-demo` is generated.

To run the program manually:

```
$ env MYSQL_USER=xxx \
      MYSQL_PASSWORD=xxx \
      MYSQL_DATABASE=xxx \
      MYSQL_HOST=xxx \
      MYSQL_PORT=3306 \
      MYSQL_TABLE=xxx \
      ./go-todo-demo
```

As you may have noticed, this program needs a few environment variables.

If you have [docker-compose](https://docs.docker.com/compose/) available, 
you can launch `go-todo-demo` and `mysql` by a single command:

```
$ docker-compose up -d
```

Now the program is running. Open a browser and visit http://localhost:3000 to play with it.

# Running it in kubernetes

With help of [kompose](https://github.com/kubernetes/kompose), moving to kubernetes is super easy.