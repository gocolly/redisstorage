package redisstorage

import (
	"testing"
)

func TestRedisBloomFilterStorage(t *testing.T) {
	var mock_requestId uint64 = 231986
	s := &RedisBloomFilterStorage{
		Storage:          &Storage{Prefix: `test`},
		RedisBloomFilter: nil,
	}
	if err := s.Init(); err != nil {
		t.Error("failed to initialize client: " + err.Error())
	}

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
