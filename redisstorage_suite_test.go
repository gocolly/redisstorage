package redisstorage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRedisStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RedisStorage Suite")
}
