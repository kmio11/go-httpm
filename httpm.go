package httpm

import (
	"net/http"

	"github.com/kmio11/go-httpc"
)

type Transport struct {
	*httpc.Transport
}

func NewTransport(mock *MockMiddleware, opts ...httpc.Option) *Transport {
	c := httpc.NewTransport(opts...).
		Use(mock)

	return &Transport{
		Transport: c,
	}
}

func NewTransportFromConfigFile(file string, opts ...httpc.Option) (*Transport, error) {
	mock, err := NewMockMiddlewareFromConfigFile(file)
	if err != nil {
		return nil, err
	}
	return NewTransport(mock, opts...), nil
}

var _ httpc.Middleware = (*MockMiddleware)(nil)

type MockMiddleware struct {
	defaultAction Mock
	mocks         []Mock
}

func NewMockMiddleware(mocks []Mock, opts ...MockMWOption) *MockMiddleware {
	m := &MockMiddleware{
		mocks:         mocks,
		defaultAction: *NewDefaultAction(newResponderFromAction(NewActionPass())),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type MockMWOption func(*MockMiddleware)

func WithDefaultAction(action Mock) MockMWOption {
	return func(m *MockMiddleware) {
		m.defaultAction = action
	}
}

func NewMockMiddlewareFromConfig(config *Config) *MockMiddleware {
	var opts []MockMWOption
	var mocks []Mock
	for _, rule := range config.Rules {
		mocks = append(mocks, *NewMockFromConfig(rule))
	}
	if config.Default != nil && config.Default.Type != "" {
		opts = append(opts, WithDefaultAction(*NewDefaultAction(newResponderFromAction(config.Default))))
	}
	return NewMockMiddleware(mocks, opts...)
}

func NewMockMiddlewareFromConfigFile(file string) (*MockMiddleware, error) {
	config, err := LoadConfigFile(file)
	if err != nil {
		return nil, err
	}
	return NewMockMiddlewareFromConfig(config), nil
}

func (m *MockMiddleware) RoundTripper(next http.RoundTripper) http.RoundTripper {
	return httpc.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		for _, mock := range m.mocks {
			if resp, matched, mocked, err := mock.Handle(req); matched {
				if !mocked {
					return next.RoundTrip(req)
				}
				return resp, err
			}
		}
		resp, matched, mocked, err := m.defaultAction.Handle(req)
		if !matched {
			panic("no match")
		}
		if !mocked {
			return next.RoundTrip(req)
		}
		return resp, err
	})
}
