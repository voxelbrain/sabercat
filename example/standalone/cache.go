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

func Cache(d time.Duration, handler http.Handler) http.Handler {
	mutex := new(sync.RWMutex)
	cache := make(map[string]*httptest.ResponseRecorder)
	if d <= 0 {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := r.URL.Path

		mutex.RLock()
		response, isCached := cache[cacheKey]
		mutex.RUnlock()

		if !isCached {
			func() {
				response = httptest.NewRecorder()
				handler.ServeHTTP(response, r)

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

				if response.Code != 200 {
					return
				}
				mutex.Lock()
				cache[cacheKey] = response
				mutex.Unlock()

				time.AfterFunc(d, func() {
					mutex.Lock()
					delete(cache, cacheKey)
					mutex.Unlock()
				})
			}()
		}

		for k, v := range response.Header() {
			for _, v := range v {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(response.Code)
		buf := bytes.NewBuffer(response.Body.Bytes())
		io.Copy(w, buf)
	})
}
