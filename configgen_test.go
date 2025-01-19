package httpm_test

import (
	"bytes"
	"net/http"
	neturl "net/url"
	"testing"

	"github.com/kmio11/gilt"
	"github.com/kmio11/go-httpc"
	"github.com/kmio11/go-httpm"
	"github.com/kmio11/go-httpm/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func request(t *testing.T, client *http.Client, method, base, path string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatal(r)
		}
	}()
	url, err := neturl.JoinPath(base, path)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer resp.Body.Close()
}

func TestConfigGenerator(t *testing.T) {
	testServer, _ := testutils.NewServer()
	t.Cleanup(func() { testServer.Close() })

	golden := gilt.NewBytesGolden(t.Name())

	var buf bytes.Buffer
	transport := httpc.NewTransport().Use(httpm.NewConfigGenerator(&buf))
	client := &http.Client{
		Transport: transport,
	}

	request(t, client, http.MethodGet, testServer.URL, "/0")
	request(t, client, http.MethodPut, testServer.URL, "/1")

	// Replace testServer.URL with example.com because the URL is not stable.
	actual := bytes.ReplaceAll(buf.Bytes(), []byte(testServer.URL), []byte("http://example.com"))

	golden.Assert(t, actual, "test", func(t *testing.T, actual []byte, expected []byte) {
		t.Helper()
		assert.Equal(t, string(expected), string(actual))
	})
}
