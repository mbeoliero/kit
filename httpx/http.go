package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/mbeoliero/kit/log"
	"github.com/mbeoliero/kit/utils/jsonx"
	"resty.dev/v3"
)

type Client struct {
	cli *resty.Client
}

var GetClient = func() *Client {
	return &Client{
		cli: resty.New(),
	}
}

func (i *Client) SetHeader(key, val string) *Client {
	i.cli.SetHeader(key, val)
	return i
}

func (i *Client) SetAuthToken(token string) *Client {
	i.cli.SetAuthToken(token)
	return i
}

func (i *Client) Post(ctx context.Context, url string, req any, bindResp any) (err error) {
	httpReq := i.cli.R().
		SetHeader("Content-Type", "application/json")
	httpReq.SetBody(req)

	res, err := httpReq.
		Post(url)
	if err != nil {
		log.CtxError(ctx, "httpx client Post url [%s] req: %v, err: %v", httpReq.URL, jsonx.MarshalToString(req), err)
		return err
	}

	resp := res.String()
	log.CtxInfo(ctx, "httpx client Post url [%s], status_code[%d] req: %v, resp: %v", httpReq.URL, res.StatusCode(), jsonx.MarshalToString(req), resp)

	if len(res.String()) > 0 && bindResp != nil {
		if err = jsoniter.Unmarshal([]byte(res.String()), bindResp); err != nil {
			log.CtxError(ctx, "httpx client Post url [%s] req: %v, err: %v", httpReq.URL, jsonx.MarshalToString(req), err)
			return fmt.Errorf("failed to unmarshal response body: %v", err)
		}
	}

	return res.Err
}

func (i *Client) Get(ctx context.Context, url string, req any, bindResp any) (err error) {
	httpReq := i.cli.R().
		SetHeader("Content-Type", "application/json")
	params, _ := toUrlValues(req)
	httpReq.SetQueryParamsFromValues(params)
	httpReq.SetResult(bindResp)

	res, err := httpReq.
		Get(url)
	if err != nil {
		log.CtxError(ctx, "httpx client Get url [%s] req: %v, err: %v", httpReq.URL, jsonx.MarshalToString(req), err)
		return err
	}

	resp := res.String()
	log.CtxInfo(ctx, "httpx client Get url [%s], status_code[%d] req: %v, resp: %v", httpReq.URL, res.StatusCode(), jsonx.MarshalToString(req), resp)

	if len(res.String()) > 0 && bindResp != nil {
		if err = jsoniter.Unmarshal([]byte(res.String()), bindResp); err != nil {
			log.CtxError(ctx, "httpx client Get url [%s] req: %v, err: %v", httpReq.URL, jsonx.MarshalToString(req), err)
			return fmt.Errorf("failed to unmarshal response body: %v", err)
		}
	}
	return res.Err
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
		b, err := jsoniter.Marshal(req)
		if err != nil {
			return nil, err
		}

		m = make(map[string]any)
		if err = jsoniter.Unmarshal(b, &m); err != nil {
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
