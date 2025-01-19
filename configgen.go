package httpm

import (
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/BurntSushi/toml"
	"github.com/kmio11/go-httpc"
)

var _ httpc.Middleware = (*ConfigGenerator)(nil)

type ConfigGenerator struct {
	w io.Writer
}

func NewConfigGenerator(w io.Writer) *ConfigGenerator {
	return &ConfigGenerator{w: w}
}

func (r *ConfigGenerator) RoundTripper(next http.RoundTripper) http.RoundTripper {
	return httpc.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(req)
		if err != nil {
			return resp, err
		}

		err = r.appendRule(r.w, req, resp)
		if err != nil {
			panic(err)
		}

		return resp, err
	})
}

func (r *ConfigGenerator) appendRule(w io.Writer, req *http.Request, resp *http.Response) error {
	method := req.Method
	url := req.URL.String()
	dumpedResp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	condition := Condition{
		Methods: []string{method},
		URL:     url,
	}

	action := NewActionMock(string(dumpedResp))

	rule := RuleConfig{
		Condition: condition,
		Action:    action,
	}

	cfg := Config{
		Default: nil,
		Rules:   []RuleConfig{rule},
	}

	e := toml.NewEncoder(w)
	if err := e.Encode(cfg); err != nil {
		return err
	}
	return nil
}
