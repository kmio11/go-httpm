package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

func NewServer() (testServer *httptest.Server, reset func()) {
	m := sync.RWMutex{}
	cnt := 0
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Lock()
		defer m.Unlock()
		// Set a fixed date to make the response deterministic
		w.Header().Set("Date", time.Date(2025, time.January, 19, 4, 20, 9, 0, time.UTC).Format(http.TimeFormat))
		fmt.Fprintf(w, "Hello, %d", cnt)
		cnt++
	})

	reset = func() {
		m.Lock()
		defer m.Unlock()
		cnt = 0
	}

	testServer = httptest.NewServer(h)
	return testServer, reset
}
