package redisstorage

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	// Client is the redis connection
	Client *redis.Client

	mu sync.RWMutex // Only used for cookie methods.
}

// Init initializes the redis storage
func (s *Storage) Init() error {
	if s.Client == nil {
		s.Client = redis.NewClient(&redis.Options{
			Addr:     s.Address,
			Password: s.Password,
			DB:       s.DB,
		})
	}
	_, err := s.Client.Ping().Result()
	if err != nil {
		return fmt.Errorf("Redis connection error: %s", err.Error())
	}
	return err
}

func (s *Storage) Destroy() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := "*"
	if s.Prefix != "" {
		prefix = s.Prefix
	}
	r := s.Client.Keys(prefix + ":cookie:*")
	keys, err := r.Result()
	if err != nil {
		return err
	}
	r2 := s.Client.Keys(prefix + ":request:*")
	keys2, err := r2.Result()
	if err != nil {
		return err
	}
	keys = append(keys, keys2...)
	return s.Client.Del(keys...).Err()
}

// Visited implements colly/storage.Visited()
func (s *Storage) Visited(requestID uint64) error {
	return s.Client.Set(s.getIDStr(requestID), "1", 0).Err()
}

// IsVisited implements colly/storage.IsVisited()
func (s *Storage) IsVisited(requestID uint64) (bool, error) {
	_, err := s.Client.Get(s.getIDStr(requestID)).Result()
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

	// TODO(js) Cookie methods currently have no way to return an error.

	// We need to use a write lock to prevent a race in the db:
	// if two callers set cookies in a very small window of time,
	// it is possible to drop the new cookies from one caller
	// ('last update wins' == best avoided).
	s.mu.Lock()
	defer s.mu.Unlock()

	cookieStr, err := s.Client.Get(s.getCookieID(u.Host)).Result()
	if err == redis.Nil {
		cookieStr = ""
	} else if err != nil {
		// return nil
		log.Printf("SetCookies() .Get error %s", err)
		return
	}

	// Merge existing cookies, new cookies have precendence.
	cnew := make([]*http.Cookie, len(cookies))
	copy(cnew, cookies)
	existing := unstringify(cookieStr)
	for _, c := range existing {
		if !contains(cnew, c.Name) {
			cnew = append(cnew, c)
		}
	}
	// return s.Client.Set(s.getCookieID(u.Host), stringify(cnew), 0).Err()
	err = s.Client.Set(s.getCookieID(u.Host), stringify(cnew), 0).Err()
	if err != nil {
		// return nil
		log.Printf("SetCookies() .Set error %s", err)
		return
	}
}

// Cookies implements http/CookieJar.Cookies()
func (s *Storage) Cookies(u *url.URL) []*http.Cookie {
	// TODO RFC 6265
	// TODO(js) Cookie methods currently have no way to return an error.

	s.mu.RLock()
	cookiesStr, err := s.Client.Get(s.getCookieID(u.Host)).Result()
	s.mu.RUnlock()
	if err == redis.Nil {
		cookiesStr = ""
	} else if err != nil {
		// return nil, err
		log.Printf("Cookies() .Get error %s", err)
		return nil
	}

	// Parse raw cookies string to []*http.Cookie.
	cookies := unstringify(cookiesStr)

	// Filter.
	now := time.Now()
	cnew := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		// Drop expired cookies.
		if c.RawExpires != "" && c.Expires.Before(now) {
			continue
		}
		// Drop secure cookies if not over https.
		if c.Secure && u.Scheme != "https" {
			continue
		}
		cnew = append(cnew, c)
	}
	// return cnew, nil
	return cnew
}

func (s *Storage) getIDStr(ID uint64) string {
	return fmt.Sprintf("%s:request:%d", s.Prefix, ID)
}

func (s *Storage) getCookieID(c string) string {
	return fmt.Sprintf("%s:cookie:%s", s.Prefix, c)
}

func stringify(cookies []*http.Cookie) string {
	// Stringify cookies.
	cs := make([]string, len(cookies))
	for i, c := range cookies {
		cs[i] = c.String()
	}
	return strings.Join(cs, "\n")
}

func unstringify(s string) []*http.Cookie {
	h := http.Header{}
	for _, c := range strings.Split(s, "\n") {
		h.Add("Set-Cookie", c)
	}
	r := http.Response{Header: h}
	return r.Cookies()
}

func contains(cookies []*http.Cookie, name string) bool {
	for _, c := range cookies {
		if c.Name == name {
			return true
		}
	}
	return false
}
