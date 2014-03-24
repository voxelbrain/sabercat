package sabercat

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	DATA_LENGTH = 8 * (1 << 10)
)

type Response struct {
	Header     map[string]string
	ReadAtMost int // Optional
	BodyLength int
}

type TableTest struct {
	Handler      http.Handler
	Responses    []Response
	RequestCount int // Optional
}

var (
	tests = map[string]TableTest{
		"Caching": TableTest{
			Handler:      NewDummyDataHandler(DATA_LENGTH, DATA_LENGTH, 200),
			RequestCount: 1,
			Responses: []Response{
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH,
				},
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH,
				},
			},
		},
		"ClientReadAbort": TableTest{
			Handler: NewDummyDataHandler(DATA_LENGTH, DATA_LENGTH, 200),
			Responses: []Response{
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					ReadAtMost: 4,
					BodyLength: 4,
				},
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH,
				},
			},
		},
		"ServerWriteAbort": TableTest{
			Handler:      NewDummyDataHandler(DATA_LENGTH, DATA_LENGTH/2, 200),
			RequestCount: 2,
			Responses: []Response{
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH / 2,
				},
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH / 2,
				},
			},
		},
		"CacheOnly200": TableTest{
			Handler:      NewDummyDataHandler(DATA_LENGTH, DATA_LENGTH, 201),
			RequestCount: 2,
			Responses: []Response{
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH,
				},
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DATA_LENGTH),
					},
					BodyLength: DATA_LENGTH,
				},
			},
		},
		"AboveCacheLimit": TableTest{
			Handler:      NewDummyDataHandler(DEFAULT_MAX_CACHEABLE_SIZE+1, DEFAULT_MAX_CACHEABLE_SIZE+1, 200),
			RequestCount: 2,
			Responses: []Response{
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DEFAULT_MAX_CACHEABLE_SIZE+1),
					},
					BodyLength: DEFAULT_MAX_CACHEABLE_SIZE + 1,
				},
				{
					Header: map[string]string{
						"Content-Length": fmt.Sprintf("%d", DEFAULT_MAX_CACHEABLE_SIZE+1),
					},
					BodyLength: DEFAULT_MAX_CACHEABLE_SIZE + 1,
				},
			},
		},
	}
)

type RequestCounterHandler struct {
	Handler http.Handler
	Counter int
}

func (rc *RequestCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rc.Counter++
	rc.Handler.ServeHTTP(w, r)
}

func TestCache_table(t *testing.T) {
	for name, test := range tests {
		t.Logf("Running %s...", name)
		rc := &RequestCounterHandler{Handler: test.Handler}
		server := httptest.NewServer(NewCache(1*time.Minute, rc))
		for _, expectedresp := range test.Responses {
			http.DefaultTransport.(*http.Transport).CloseIdleConnections()
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("%s: Could not send request: %s", name, err)
			}

			for header, expectedv := range expectedresp.Header {
				if v := resp.Header.Get(header); v != expectedv {
					t.Fatalf("%s: Header field %s has unexpected value. Expected \"%s\", got \"%s\"", name, header, expectedv, v)
				}
			}

			var buf []byte
			if expectedresp.ReadAtMost != 0 {
				buf = make([]byte, expectedresp.ReadAtMost)
				resp.Body.Read(buf)
			} else {
				buf, _ = ioutil.ReadAll(resp.Body)
			}
			if err != nil {
				t.Fatalf("%s: Reading from body failed: %s", name, err)
			}
			if len(buf) != expectedresp.BodyLength {
				t.Fatalf("%s: Unexpected body length. Expected %d, got %d", name, expectedresp.BodyLength, len(buf))
			}
		}
		server.Close()
		if test.RequestCount != 0 && rc.Counter != test.RequestCount {
			t.Fatalf("%s: Unexpected request count. Expected %d, got %d", name, test.RequestCount, rc.Counter)
		}
	}
}

func NewDummyDataHandler(announcedContentLength, actualContentLength, errorCode int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", announcedContentLength))
		w.WriteHeader(errorCode)
		buf := make([]byte, actualContentLength)
		w.Write(buf)
	})
}
