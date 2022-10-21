package redisstorage

import (
	"encoding/binary"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/gocolly/redisstorage/filter"
)

// RedisBloomFilterStorage implements the redis bloom filter storage for Colly
type RedisBloomFilterStorage struct {
	// redisstorage.Storage
	*Storage
	// RedisBloomFilter implements Bloom filter based on redis
	RedisBloomFilter *filter.BloomFilter
}

// Init initializes the redis bloom filter storage
func (s *RedisBloomFilterStorage) Init() error {
	if s.Client == nil {
		s.Client = redis.NewClient(&redis.Options{
			Addr:     s.Address,
			Password: s.Password,
			DB:       s.DB,
		})
	}
	s.RedisBloomFilter = filter.NewBloomFilter(s.Client, s.Prefix+"_bloom")
	if _, err := s.Client.Ping().Result(); err != nil {
		return fmt.Errorf("redis connection error: %s", err.Error())
	}
	return nil
}

// Visited implements colly/storage.Visited(), base on redis bloom filter
func (s *RedisBloomFilterStorage) Visited(requestID uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, requestID)
	return s.RedisBloomFilter.Add(b)
}

// IsVisited implements colly/storage.IsVisited(), base on redis bloom filter
func (s *RedisBloomFilterStorage) IsVisited(requestID uint64) (bool, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, requestID)
	return s.RedisBloomFilter.Exists(b)
}
