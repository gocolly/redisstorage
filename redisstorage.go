package redisstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// Storage implements the redis storage backend for Colly
type Storage struct {
	// Address is the redis server address
	Address string
	// Password is the password for the redis server
	Password string
	// DB is the redis database. Default is 0
	DB int
	// Prefix is an optional string in the keys. It can be used
	// to use one redis database for independent scraping tasks.
	Prefix string
	client *redis.Client
}

// Init initializes the redis storage
func (s *Storage) Init() error {
	if s.client == nil {
		s.client = redis.NewClient(&redis.Options{
			Addr:     s.Address,
			Password: s.Password,
			DB:       s.DB,
		})
	}
	_, err := s.client.Ping().Result()
	if err != nil {
		return fmt.Errorf("Redis connection error: %s", err.Error())
	}
	return err
}

// Visited implements colly/storage.Visited()
func (s *Storage) Visited(requestID uint64) error {
	return s.client.Set(s.getIDStr(requestID), "1", 0).Err()
}

// IsVisited implements colly/storage.IsVisited()
func (s *Storage) IsVisited(requestID uint64) (bool, error) {
	_, err := s.client.Get(s.getIDStr(requestID)).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// GetCookieJar implements colly/storage.GetCookieJar()
func (s *Storage) GetCookieJar() http.CookieJar {
	return s
}

// SetCookies implements http/CookieJar.SetCookies()
func (s *Storage) SetCookies(u *url.URL, cookies []*http.Cookie) {
	// TODO RFC 6265
	cookieStrings := make([]string, 0, len(cookies))
	for _, c := range cookies {
		cookieStrings = append(cookieStrings, c.String())
	}
	for _, c := range s.Cookies(u) {
		duplication := false
		for _, c2 := range cookies {
			if c2.Name == c.Name {
				duplication = true
				break
			}
		}
		if !duplication {
			cookieStrings = append(cookieStrings, c.String())
		}
	}
	s.client.Set(s.getCookieID(u.Host), strings.Join(cookieStrings, "\n"), 0).Err()
}

// Cookies implements http/CookieJar.Cookies()
func (s *Storage) Cookies(u *url.URL) []*http.Cookie {
	// TODO RFC 6265
	cookieStr, err := s.client.Get(s.getCookieID(u.Host)).Result()
	if err != nil {
		return nil
	}
	header := http.Header{}
	for _, cs := range strings.Split(cookieStr, "\n") {
		header.Add("Cookie", cs)
	}
	request := http.Request{Header: header}
	rc := request.Cookies()
	cookies := make([]*http.Cookie, 0, len(rc))
	now := time.Now()
	for _, c := range rc {
		if c.RawExpires != "" && !c.Expires.After(now) {
			continue
		}
		cookies = append(cookies, c)
	}
	return cookies
}

// Close implements colly/storage.Close() and closes redis connection
func (s *Storage) Close() error {
	s.client.Close()
	s.client = nil
	return nil
}

func (s *Storage) getIDStr(ID uint64) string {
	return fmt.Sprintf("%s:request:%d", s.Prefix, ID)
}

func (s *Storage) getCookieID(c string) string {
	return fmt.Sprintf("%s:cookie:%s", s.Prefix, c)
}
