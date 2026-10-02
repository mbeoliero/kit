package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mbeoliero/kit/log"
	"github.com/mbeoliero/kit/utils/jsonx"
	"resty.dev/v3"
)

// Bodies are logged at debug level and cut to this many bytes.
const maxLoggedBody = 2048

// sharedClient pools connections across every Client; per-Client headers are applied to
// each request instead of the shared client.
var sharedClient = sync.OnceValue(resty.New)

// StatusError reports a response outside 2xx. The body has still been decoded into the
// caller's response value when it was valid JSON.
type StatusError struct {
	Method     string
	Url        string
	StatusCode int
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("httpx %s %s: status %d", e.Method, e.Url, e.StatusCode)
}

type Client struct {
	disableLog bool
	headers    map[string]string
	authToken  string
}

var GetClient = func() *Client {
	return &Client{}
}

var GetNoLogClient = func() *Client {
	return &Client{disableLog: true}
}

func (i *Client) SetHeader(key, val string) *Client {
	if i.headers == nil {
		i.headers = map[string]string{}
	}
	i.headers[key] = val
	return i
}

func (i *Client) SetAuthToken(token string) *Client {
	i.authToken = token
	return i
}

func (i *Client) request(ctx context.Context) *resty.Request {
	req := sharedClient().R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeaders(maps.Clone(i.headers))
	if i.authToken != "" {
		req.SetAuthToken(i.authToken)
	}
	return req
}

func (i *Client) Post(ctx context.Context, url string, req any, bindResp any) error {
	return i.do(ctx, i.request(ctx).SetBody(req), resty.MethodPost, url, req, bindResp)
}

func (i *Client) Get(ctx context.Context, url string, req any, bindResp any) error {
	params, err := toUrlValues(req)
	if err != nil {
		return fmt.Errorf("httpx encode query: %w", err)
	}
	return i.do(ctx, i.request(ctx).SetQueryParamsFromValues(params), resty.MethodGet, url, req, bindResp)
}

func (i *Client) do(ctx context.Context, httpReq *resty.Request, method, rawUrl string, req, bindResp any) error {
	begin := time.Now()
	res, err := httpReq.Execute(method, rawUrl)
	target := redactUrl(rawUrl)
	if err != nil {
		if !i.disableLog {
			log.CtxError(ctx, "httpx %s %s failed after %v: %v", method, target, time.Since(begin), err)
		}
		return fmt.Errorf("httpx %s %s: %w", method, target, err)
	}

	body := res.Bytes()
	if !i.disableLog {
		log.CtxInfo(ctx, "httpx %s %s status=%d latency=%v", method, target, res.StatusCode(), time.Since(begin))
		log.CtxDebug(ctx, "httpx %s %s req: %s, resp: %s", method, target, truncate(jsonx.MarshalToString(req)), truncate(string(body)))
	}
	if len(body) > 0 && bindResp != nil {
		if err := json.Unmarshal(body, bindResp); err != nil && res.IsStatusSuccess() {
			return fmt.Errorf("httpx %s %s decode response: %w", method, target, err)
		}
	}
	if !res.IsStatusSuccess() {
		return &StatusError{Method: method, Url: target, StatusCode: res.StatusCode(), Body: string(body)}
	}
	return nil
}

// redactUrl drops the query string, which may carry tokens or user data.
func redactUrl(rawUrl string) string {
	path, _, _ := strings.Cut(rawUrl, "?")
	return path
}

func truncate(s string) string {
	if len(s) <= maxLoggedBody {
		return s
	}
	return s[:maxLoggedBody] + "...(truncated)"
}

func toUrlValues(req any) (url.Values, error) {
	ret := make(url.Values)
	if req == nil {
		return ret, nil
	}

	if m, ok := req.(map[string]string); ok {
		for k, v := range m {
			ret.Add(k, v)
		}
		return ret, nil
	}

	m, ok := req.(map[string]any)
	if !ok {
		b, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		// UseNumber keeps int64 values such as IDs exact instead of rounding through float64.
		decoder := json.NewDecoder(bytes.NewReader(b))
		decoder.UseNumber()
		if err = decoder.Decode(&m); err != nil {
			return nil, err
		}
	}

	for k, v := range m {
		addValue(ret, k, v)
	}

	return ret, nil
}

func addValue(values url.Values, key string, v any) {
	if v == nil {
		return
	}

	switch val := v.(type) {
	case []any:
		for _, item := range val {
			values.Add(key, toString(item))
		}
	case []string:
		for _, item := range val {
			values.Add(key, item)
		}
	default:
		values.Add(key, toString(val))
	}
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int32:
		return strconv.FormatInt(int64(s), 10)
	case int64:
		return strconv.FormatInt(s, 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case json.Number:
		return s.String()
	case []byte:
		return string(s)
	case nil:
		return ""
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return ""
	}
}
