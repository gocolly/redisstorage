# Redis Storage for Colly

This is a redis based storage backend for [Colly](https://github.com/gocolly/colly) collectors.

[![GoDoc](https://godoc.org/github.com/gocolly/redisstorage?status.svg)](https://godoc.org/github.com/gocolly/redisstorage)


## Redis Storage

### Install

```
go get -u github.com/gocolly/redisstorage
```


### Usage

```go
import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/redisstorage"
)
```


```go
c := colly.NewCollector()

storage := &redisstorage.Storage{
    Address:  "127.0.0.1:6379",
    Password: "",
    DB:       0,
    Prefix:   "job01",
}

if err := c.SetStorage(storage); err != nil{
    panic(err)
}
```

## RedisBloomFilterStorage

```go
import (
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/redisstorage"
)
```


```go
c := colly.NewCollector()

storage := &redisstorage.RedisBloomFilterStorage{
    Storage: &Storage{
        Address:  "127.0.0.1:6379",
        Password: "",
        DB:       0,
        Prefix:   "job01",
    }
}

if err := c.SetStorage(storage); err != {
    panic(err)
}

```

## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/gocolly/redisstorage/issues) or join `#colly` on freenode
