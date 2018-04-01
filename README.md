# mallocd

here is my memcache killer written in go, mallocd

## demo

```
➜  ~ go run cmd/server/mallocd.go

➜  ~ ptr=$(go run cmd/client/client.go malloc 5)
➜  ~ go run cmd/client/client.go write $ptr 5 hello
➜  ~ go run cmd/client/client.go read $ptr 5
➜  ~ go run cmd/client/client.go free $ptr
```

## design

the design of mallocd is described in our paper: [mallocd: designing a garbage-free nosql data store](https://github.com/igorwwwwwwwwwwwwwwwwwwww/mallocd/blob/master/mallocd.pdf), which has been published in [sigbovik 2018](http://sigbovik.org/2018/).

## performance

you can run a stress test with gc tracing enabled to verify that mallocd is garbage-free.

```
➜  ~ GODEBUG=gctrace=2 go run cmd/server/mallocd.go
➜  ~ go run cmd/stress/stress.go 1000
```

## license

permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "software"), to `kill -9`.
