package main

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
	MAX_CACHEABLE_SIZE = 20 * (1 << 20) //20 Mibis
)

type Cache struct {
	TTL     time.Duration
	Handler http.Handler
	mutex   *sync.RWMutex
	cache   map[string]*httptest.ResponseRecorder
}

func NewCache(d time.Duration, handler http.Handler) *Cache {
	return &Cache{
		TTL:     d,
		Handler: handler,
		mutex:   &sync.RWMutex{},
		cache:   make(map[string]*httptest.ResponseRecorder),
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
	if length <= 0 || length > MAX_CACHEABLE_SIZE || err != nil {
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
