package redisstorage

import (
	"testing"

	"github.com/gocolly/colly/queue"
)

func TestQueue(t *testing.T) {
	s := &Storage{
		Address:  "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Prefix:   "queue_test",
	}

	if err := s.Init(); err != nil {
		t.Error("failed to initialize client: " + err.Error())
		return
	}
	defer s.Clear()
	urls := []string{"http://example.com/", "http://go-colly.org/", "https://xx.yy/zz"}
	for _, u := range urls {
		if err := s.AddRequest(&queue.Request{Method: "GET", URL: u}); err != nil {
			t.Error("failed to add request: " + err.Error())
			return
		}
	}
	if size, err := s.QueueSize(); size != 3 || err != nil {
		t.Error("invalid queue size")
		return
	}
	for _, u := range urls {
		if r, err := s.GetRequest(); err != nil || r.URL != u {
			t.Error("failed to get request: " + err.Error())
			return
		}
	}
}
