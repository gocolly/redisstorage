package filter

import (
	"errors"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/spaolacci/murmur3"
)

const (
	setScript = `for _, offset in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], offset, 1)
end`
	testScript = `for _, offset in ipairs(ARGV) do
	if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
		return false
	end
end
return true`
)

// bitSetProvider
type bitSetProvider interface {
	check([]uint) (bool, error)
	set([]uint) error
}

// BloomFilter A Filter is a bloom filter.
type BloomFilter struct {
	bits   uint
	maps   uint
	bitSet bitSetProvider
}

// NewBloomFilter create a bloomFilter, store is the backed redis, key is the key for the bloom filter,
// bits is how many bits will be used, maps is how many hashes for each addition.
// best practices:
// elements - means how many actual elements
// when maps = 14, formula: 0.7*(bits/maps), bits = 20*elements, the error rate is 0.000067 < 1e-4
// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
func NewBloomFilter(store *redis.Client, key string) *BloomFilter {
	var bits uint = 2 << 31
	var maps uint = 14
	return &BloomFilter{
		bits:   bits,
		maps:   maps,
		bitSet: newRedisBitSet(store, key, bits),
	}
}

// Add adds data into f.
func (f *BloomFilter) Add(data []byte) error {
	locations := f.getLocations(data)
	return f.bitSet.set(locations)
}

// Exists checks if data is in f.
func (f *BloomFilter) Exists(data []byte) (bool, error) {
	locations := f.getLocations(data)
	isSet, err := f.bitSet.check(locations)
	if err != nil {
		return false, err
	}

	return isSet, nil
}

func (f *BloomFilter) getLocations(data []byte) []uint {
	locations := make([]uint, f.maps)
	for i := uint(0); i < f.maps; i++ {
		hashValue := murmur3.Sum64(append(data, byte(i)))
		locations[i] = uint(hashValue % uint64(f.bits))
	}
	return locations
}

type redisBitSet struct {
	store *redis.Client
	key   string
	bits  uint
}

func newRedisBitSet(store *redis.Client, key string, bits uint) *redisBitSet {
	return &redisBitSet{
		store: store,
		key:   key,
		bits:  bits,
	}
}

func (r *redisBitSet) buildOffsetArgs(offsets []uint) ([]string, error) {
	var args []string

	for _, offset := range offsets {
		if offset >= r.bits {
			return nil, errors.New(`redisbloomfiler:Too large offset`)
		}

		args = append(args, strconv.FormatUint(uint64(offset), 10))
	}

	return args, nil
}

func (r *redisBitSet) check(offsets []uint) (bool, error) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}

	resp, err := r.store.Eval(testScript, []string{r.key}, args).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return exists == 1, nil
}

func (r *redisBitSet) del() error {
	_, err := r.store.Del(r.key).Result()
	return err
}

func (r *redisBitSet) set(offsets []uint) error {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}

	_, err = r.store.Eval(setScript, []string{r.key}, args).Result()
	if err == redis.Nil {
		return nil
	}

	return err
}
