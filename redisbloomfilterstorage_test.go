package redisstorage

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestRedisBloomFilterStorage(t *testing.T) {
	// mock rquestID to Verify bloomfilter
	mock_requestId := rand.Uint64()
	fmt.Println(mock_requestId)
	s := &RedisBloomFilterStorage{
		Storage: &Storage{Prefix: `test`},
	}
	if err := s.Init(); err != nil {
		t.Error("failed to initialize client: " + err.Error())
	}
	defer s.Clear()

	t.Run(testing.CoverMode(), func(t *testing.T) {
		if err := s.Visited(mock_requestId); err != nil {
			t.Error("fail set redis: " + err.Error())
		}
	})

	t.Run(testing.CoverMode(), func(t *testing.T) {
		if isV, err := s.IsVisited(mock_requestId); err != nil {
			t.Error("failed to initialize client: " + err.Error())
		} else {
			if !isV {
				t.Error("invalid bloom filter ")
				return
			}
		}
	})
}
