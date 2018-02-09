# Redis Storage for Colly

This is a redis based storage backend for [Colly](https://github.com/gocolly/colly) collectors.


## Install

```
go get -u github.com/gocolly/redisstorage
```


## Usage

```go
import (
	"github.com/gocolly/colly"
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

defer storage.Close()

err := c.SetStorage(storage)
if err != nil {
    panic(err)
}
```


## Bugs

Bugs or suggestions? Visit the [issue tracker](https://github.com/gocolly/redisstorage/issues) or join `#colly` on freenode
