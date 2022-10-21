package main

import (
	"fmt"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/gocolly/redisstorage"
)

func main() {
	c := colly.NewCollector()

	storage := &redisstorage.Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "job01",
	}

	if err := c.SetStorage(storage); err != err {
		panic(err)
	}
	// close redis client
	//defer storage.Client.Close()

	q, err := queue.New(10, storage)
	if err != nil {
		return
	}

	c.OnResponse(func(response *colly.Response) {
		fmt.Println(string(response.Body))
	})

	for i := 0; i < 10; i++ {
		q.AddURL(fmt.Sprintf("%s?x=%v", `https://httpbin.org/delay/1`, i))
	}

	q.Run(c)
}
