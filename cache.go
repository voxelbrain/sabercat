package sabercat

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"
)

const (
	// Per default items up to 20 Mibibytes in size will be cached
	DEFAULT_MAX_CACHEABLE_SIZE = 20 * (1 << (10 * 2))
)

// Cache is a `net/http.Handler`, which uses `net/http/httptest.ResponseRecorder`
// to record and recall the responses of the wrapped handler from memory
// if the same URL is requested again. The `net/http.Request.URL.Path` is
// used as the caching key.
type Cache struct {
	// Time for which a answer is held im memory. A TTL <= 0
	// makes the Cache pass all requests directly to the handler.
	TTL time.Duration
	// Handler, whose responses are supposed to be cached
	Handler http.Handler
	// Maximum number of bytes, with which an response is eligible
	// for caching (default: 20MiB)
	MaxCacheableSize int64
	mutex            *sync.RWMutex
	cache            map[string]*httptest.ResponseRecorder
}

// NewCache returns a new Cache, which records the responses of the given
// handler for the given duration.
func NewCache(d time.Duration, handler http.Handler) *Cache {
	return &Cache{
		TTL:              d,
		Handler:          handler,
		MaxCacheableSize: DEFAULT_MAX_CACHEABLE_SIZE,
		mutex:            &sync.RWMutex{},
		cache:            make(map[string]*httptest.ResponseRecorder),
	}
}

func (ch *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ch.TTL <= 0 {
		ch.Handler.ServeHTTP(w, r)
		return
	}
	cacheKey := r.URL.Path

	ch.mutex.RLock()
	response, isCached := ch.cache[cacheKey]
	ch.mutex.RUnlock()

	if !isCached {
		response = ch.fillCache(r)
	}

	for k, v := range response.Header() {
		for _, v := range v {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Expires", time.Now().Add(2592000*time.Second).Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "private,max-age=2592000")
	w.Header().Set("Pragma", "no-cache")

	w.WriteHeader(response.Code)
	buf := bytes.NewBuffer(response.Body.Bytes())
	io.Copy(w, buf)
}

func (ch *Cache) fillCache(r *http.Request) (response *httptest.ResponseRecorder) {
	cacheKey := r.URL.Path
	response = httptest.NewRecorder()

	ch.Handler.ServeHTTP(response, r)
	if response.Code != 200 {
		return
	}

	length, err := strconv.ParseInt(response.Header().Get("Content-Length"), 10, 64)
	if length <= 0 || length > ch.MaxCacheableSize || err != nil {
		return
	}

	newbuf := new(bytes.Buffer)
	n, err := newbuf.ReadFrom(response.Body)
	response.Body = newbuf
	if n != length || err != nil {
		return
	}

	ch.mutex.Lock()
	ch.cache[cacheKey] = response
	ch.mutex.Unlock()

	time.AfterFunc(ch.TTL, func() {
		ch.mutex.Lock()
		delete(ch.cache, cacheKey)
		ch.mutex.Unlock()
	})
	return
}
