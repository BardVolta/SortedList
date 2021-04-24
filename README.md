## Introduction

sorted list a high-performance & scalable concurrent sorted list.  

## Features

- Concurrent safe API with high-performance.
- Wait-free Contains and Range operations.
- Sorted items.

## Test
```
go version go1.14.10 darwin/amd64

> go test sortedlist.go sortedlist_test.go
ok      command-line-arguments  0.821s

> go test -race sortedlist.go sortedlist_test.go
ok      command-line-arguments  33.118s
```