package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

func Cache(d time.Duration, handler http.Handler) http.Handler {
	mutex := new(sync.RWMutex)
	cache := make(map[string]*httptest.ResponseRecorder)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := r.URL.Path

		mutex.RLock()
		response, ok := cache[cacheKey]
		mutex.RUnlock()

		if !ok {
			response = httptest.NewRecorder()
			handler.ServeHTTP(response, r)

			if response.Code < 300 {
				mutex.Lock()
				cache[cacheKey] = response
				mutex.Unlock()

				time.AfterFunc(d, func() {
					mutex.Lock()
					delete(cache, cacheKey)
					mutex.Unlock()
				})
			}

		}

		for k, v := range response.Header() {
			for _, v := range v {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(response.Code)
		io.Copy(w, bytes.NewBuffer(response.Body.Bytes()))
	})
}
