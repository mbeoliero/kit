package httpx

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type query struct {
	UserId int64    `json:"user_id"`
	Tags   []string `json:"tags"`
}

func TestGetKeepsInt64QueryValues(t *testing.T) {
	var got http.Header
	var userId string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		userId = r.URL.Query().Get("user_id")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	var resp struct{ Ok bool }
	err := GetNoLogClient().SetHeader("X-App", "kit").SetAuthToken("t").
		Get(t.Context(), srv.URL, query{UserId: 1234567890123456789, Tags: []string{"a"}}, &resp)
	if err != nil || !resp.Ok {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if userId != "1234567890123456789" {
		t.Fatalf("user_id = %s", userId)
	}
	if got.Get("X-App") != "kit" || got.Get("Authorization") != "Bearer t" {
		t.Fatalf("headers = %v", got)
	}
}

func TestNonSuccessStatusReturnsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"code":40001}`)
	}))
	defer srv.Close()

	var resp struct{ Code int }
	err := GetNoLogClient().Post(t.Context(), srv.URL+"?token=secret", map[string]any{"a": 1}, &resp)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v", err)
	}
	if statusErr.Url != srv.URL {
		t.Fatalf("error url keeps query: %s", statusErr.Url)
	}
	if resp.Code != 40001 {
		t.Fatalf("error body was not decoded: %+v", resp)
	}
}

func TestHeadersStayPerClient(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("X-App"))
	}))
	defer srv.Close()

	_ = GetNoLogClient().SetHeader("X-App", "first").Get(t.Context(), srv.URL, nil, nil)
	_ = GetNoLogClient().Get(t.Context(), srv.URL, nil, nil)
	if len(seen) != 2 || seen[0] != "first" || seen[1] != "" {
		t.Fatalf("seen = %q", seen)
	}
}
