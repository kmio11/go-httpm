package httpm

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type Mock struct {
	match    Matcher
	response Responder
}

type Matcher interface {
	Match(*http.Request) bool
}

type MatcherFunc func(*http.Request) bool

func (f MatcherFunc) Match(req *http.Request) bool {
	return f(req)
}

type Responder interface {
	Response(req *http.Request) (*http.Response, bool)
}

type ResponderFunc func(req *http.Request) (*http.Response, bool)

// Response returns the response and a boolean indicating whether the response is mocked.
// If the response is not mocked, the boolean will be false, and the response will be nil.
// The caller should check the boolean value before using the response. If the boolean is false,
// the caller should proceed with the request as normal.
func (f ResponderFunc) Response(req *http.Request) (*http.Response, bool) {
	return f(req)
}

type RequestMatcher struct {
	URL     string
	Methods []string
}

func (m *Mock) Handle(req *http.Request) (resp *http.Response, matched bool, mocked bool, err error) {
	if m.match.Match(req) {
		resp, mocked := m.response.Response(req)
		return resp, true, mocked, nil
	}
	return nil, false, false, nil
}

func NewMockFromConfig(rule RuleConfig) *Mock {
	return &Mock{
		match:    newMatcherFromConfig(rule.Condition),
		response: newResponderFromAction(rule.Action),
	}
}

// NewDefaultAction creates a new Mock instance with a default matcher that always returns true
// and the provided Responder. This is useful for setting up a default action that will match
// any HTTP request.
func NewDefaultAction(resp Responder) *Mock {
	return &Mock{
		match:    MatcherFunc(func(*http.Request) bool { return true }),
		response: resp,
	}
}

func newMatcherFromConfig(condition Condition) *RequestMatcher {
	return &RequestMatcher{
		Methods: condition.Methods,
		URL:     condition.URL,
	}
}

func newResponderFromAction(action *Action) ResponderFunc {
	if action == nil {
		action = NewActionPanic()
	}
	return ResponderFunc(func(req *http.Request) (*http.Response, bool) {
		switch action.Type {
		case ActionTypePass:
			return nil, false
		case ActionTypePanic:
			method := req.Method
			url := req.URL.String()
			panic(fmt.Sprintf("[mock] request caused a panic. method:%s url:%s", method, url))
		case ActionTypeMock:
			resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader([]byte(action.Response))), nil)
			if err != nil {
				panic(err)
			}
			return resp, true
		default:
			panic(fmt.Sprintf("[mock] unknown action type: %s", action.Type))
		}
	})
}

func (m *RequestMatcher) Match(req *http.Request) bool {
	// Check URL match
	if m.URL != "" && !matchPattern(m.URL, req.URL.String()) {
		return false
	}

	// Check Methods match
	if len(m.Methods) > 0 {
		methodMatch := false
		for _, method := range m.Methods {
			if method == "*" || method == req.Method {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	return true
}

func matchPattern(pattern, str string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == str
	}

	if !strings.HasPrefix(str, parts[0]) {
		return false
	}

	str = str[len(parts[0]):]
	for i := 1; i < len(parts)-1; i++ {
		if idx := strings.Index(str, parts[i]); idx == -1 {
			return false
		} else {
			str = str[idx+len(parts[i]):]
		}
	}

	return strings.HasSuffix(str, parts[len(parts)-1])
}
