package redisstorage_test

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jimsmart/redisstorage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage CookieJar", func() {

	newStore := func() *redisstorage.Storage {
		// TODO(js) For ease of testing, these settings should probably come from environment variables.
		s := &redisstorage.Storage{
			Address:  "127.0.0.1:32768",
			Password: "",
			DB:       0,
			Prefix:   randomName(),
		}
		return s
	}

	It("should set and get cookies", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())

		Expect(got).To(HaveLen(2))
		sgot := toStrings(got)
		Expect(sgot).To(ContainElement("cookie1_name=cookie1_value; Path=/; Domain=example.org"))
		Expect(sgot).To(ContainElement(cookies[0].String()))
		Expect(sgot).To(ContainElement(cookies[1].String()))
	})

	It("should update existing cookies", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)

		// Change existing.
		update := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value_new",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, update)).To(BeNil())
		s.SetCookies(url, update)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))
		sgot := toStrings(got)
		Expect(sgot).To(ContainElement(update[0].String()))
	})

	It("should add new cookies to existing cookies", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)

		// Add another.
		more := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie3_name",
				Value:  "cookie3_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, more)).To(BeNil())
		s.SetCookies(url, more)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(3))
		sgot := toStrings(got)
		Expect(sgot).To(ContainElement(more[0].String()))
	})

	It("should drop expired cookies", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))

		// Expire a cookie.
		expired := []*http.Cookie{
			&http.Cookie{
				Name:    "cookie1_name",
				Path:    "/",
				Domain:  ".example.org",
				Expires: time.Now(),
			},
		}
		// Expect(s.SetCookies(url, expired)).To(BeNil())
		s.SetCookies(url, expired)
		// got, err = s.Cookies(url)
		got = s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[0].String()).To(Equal("cookie2_name=cookie2_value; Path=/; Domain=example.org"))
	})

	It("should drop secure cookies if not over https", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies - one is marked secure.
		url, _ := url.Parse("https://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_value",
				Path:   "/",
				Domain: ".example.org",
				Secure: true,
			},
			&http.Cookie{
				Name:   "cookie2_name",
				Value:  "cookie2_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(2))
		// Get for http.
		url, _ = url.Parse("http://example.org")
		// got, err = s.Cookies(url)
		got = s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		Expect(got[0].String()).To(Equal("cookie2_name=cookie2_value; Path=/; Domain=example.org"))
	})

	It("should not get cookies for an unknown domain", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// Get non-existing.
		url, _ := url.Parse("http://no-such-domain.org")
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(0))
	})

	It("should handle cookies containing a newline", func() {
		s := newStore()
		Expect(s.Init()).To(BeNil())
		defer s.Destroy()

		// SetCookies.
		url, _ := url.Parse("http://example.org")
		cookies := []*http.Cookie{
			&http.Cookie{
				Name:   "cookie1_name",
				Value:  "cookie1_\n_value",
				Path:   "/",
				Domain: ".example.org",
			},
		}
		// Expect(s.SetCookies(url, cookies)).To(BeNil())
		s.SetCookies(url, cookies)
		// Get existing.
		// got, err := s.Cookies(url)
		got := s.Cookies(url)
		// Expect(err).To(BeNil())
		Expect(got).To(HaveLen(1))
		// It turns out that this is ok, net/http handles this,
		// and emits a warning  ('dropping invalid bytes') to the console when it does so.
		Expect(got[0].String()).To(Equal("cookie1_name=cookie1__value; Path=/; Domain=example.org"))
		Expect(got[0].String()).To(Equal(cookies[0].String()))
	})

})

func toStrings(cookies []*http.Cookie) []string {
	s := make([]string, len(cookies))
	for i, c := range cookies {
		s[i] = c.String()
	}
	return s
}

func randomName() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	h := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(h, b)
	return string(h)
}
