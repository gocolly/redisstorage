package main

import (
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/gocolly/redisstorage"
)

func main() {
	c := colly.NewCollector()
	//c.AllowURLRevisit = true
	redisbloomfilterstorage := &redisstorage.RedisBloomFilterStorage{
		Storage: &redisstorage.Storage{
			Address:  "127.0.0.1:6379",
			Password: "",
			DB:       0,
			Prefix:   "bl",
		},
	}
	if err := redisbloomfilterstorage.Init(); err != nil {
		panic(err)
	}

	if err := c.SetStorage(redisbloomfilterstorage); err != err {
		panic(err)
	}
	// close redis client
	defer redisbloomfilterstorage.Client.Close()

	defer redisbloomfilterstorage.Clear()

	q, err := queue.New(10, redisbloomfilterstorage)
	if err != nil {
		return
	}

	c.OnResponse(func(response *colly.Response) {
		fmt.Println(string(response.Body))
	})

	for i := 0; i < 10<<10; i++ {
		q.AddURL(fmt.Sprintf("%s?x=%v", `https://httpbin.org/delay/1`, i))
	}

	q.Run(c)
}
