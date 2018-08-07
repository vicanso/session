package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vicanso/keygrip"

	"github.com/go-redis/redis"
)

func TestSession(t *testing.T) {
	store := NewRedisStore(nil, &redis.Options{
		Addr: "localhost:6379",
	})
	keys := []string{
		"tree.xie",
		"vicanso",
	}
	t.Run("fetch session when no cookies", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		w := httptest.NewRecorder()

		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		data, err := sess.Fetch()
		if err != nil {
			t.Fatalf("no session cookie, get session fail, %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("no session cookie should get init data")
		}
	})

	t.Run("fetch session when cookie exists and not signed", func(t *testing.T) {
		cookieValue := generateID("")
		cookie := &http.Cookie{
			Name:  defaultCookieName,
			Value: cookieValue,
		}
		myName := "tree.xie"
		buf, _ := json.Marshal(map[string]interface{}{
			"name": myName,
		})
		store.Set(cookieValue, buf, 60)
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store: store,
		})
		data, err := sess.Fetch()
		if err != nil {
			t.Fatalf("get session fail, %v", err)
		}
		if data["name"].(string) != myName {
			fmt.Println("get session data fail")
		}
		// fetch again
		data, err = sess.Fetch()
		if err != nil {
			t.Fatalf("get session fail, %v", err)
		}
		if data["name"].(string) != myName {
			fmt.Println("get session data again fail")
		}
	})

	t.Run("fetch empty session", func(t *testing.T) {
		cookieValue := generateID("")
		cookie := &http.Cookie{
			Name:  defaultCookieName,
			Value: cookieValue,
		}
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store: store,
		})
		data, err := sess.Fetch()
		if err != nil {
			t.Fatalf("get empty session fail, %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("get empty session should be init data")
		}
	})

	t.Run("fetch session when cookie exists and signed incorrect", func(t *testing.T) {
		cookieValue := generateID("")
		cookie := &http.Cookie{
			Name:  defaultCookieName,
			Value: cookieValue,
		}
		sigCookie := &http.Cookie{
			Name:  defaultCookieName + ".sig",
			Value: "abcd",
		}
		myName := "tree.xie"
		buf, _ := json.Marshal(map[string]interface{}{
			"name": myName,
		})
		store.Set(cookieValue, buf, 60)
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		r.AddCookie(cookie)
		r.AddCookie(sigCookie)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		data, err := sess.Fetch()
		if err != nil {
			t.Fatalf("get signed incorrect session fail, %v", err)
		}
		if len(data) != 1 {
			t.Fatalf("signed incorrect should be same as no cookie")
		}
	})

	t.Run("fetch session when cookie exists and signed correct", func(t *testing.T) {
		kg := keygrip.New(keys)
		cookieValue := generateID("")
		cookie := &http.Cookie{
			Name:  defaultCookieName,
			Value: cookieValue,
		}
		sigCookie := &http.Cookie{
			Name:  defaultCookieName + ".sig",
			Value: kg.Sign(defaultCookieName + "=" + cookieValue),
		}
		myName := "tree.xie"
		buf, _ := json.Marshal(map[string]interface{}{
			"name": myName,
		})
		store.Set(cookieValue, buf, 60)
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		r.AddCookie(cookie)
		r.AddCookie(sigCookie)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		data, err := sess.Fetch()
		if err != nil {
			t.Fatalf("get session data fail, %v", err)
		}
		if data["name"].(string) != myName {
			t.Fatalf("get session data fail")
		}
		if sess.Get("name").(string) != myName {
			t.Fatalf("get session data fail")
		}
	})

	t.Run("set session data", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		err := sess.Set("", nil)
		if err != nil {
			t.Fatalf("set with empty key should not return error")
		}
		key := "name"
		value := "tree.xie"
		err = sess.Set(key, value)
		if err != ErrNotFetched {
			t.Fatalf("the session should fetch before set")
		}
		sess.Fetch()
		err = sess.Set(key, value)
		if err != nil {
			t.Fatalf("set session fail, %v", err)
		}
		data, _ := sess.Fetch()
		if data[key] != value {
			t.Fatalf("get data from session fail, after set")
		}
	})

	t.Run("get created/updated at", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		if sess.GetCreatedAt() != "" {
			t.Fatalf("not fetch session's createdAt should return empty")
		}
		sess.Fetch()
		if sess.GetCreatedAt() == "" {
			t.Fatalf("fetch session's createdAt should return date string")
		}

		if sess.GetUpdatedAt() != "" {
			t.Fatalf("not modified session's updatedAt should return empty")
		}
		sess.Set("name", "tree.xie")

		if sess.GetUpdatedAt() == "" {
			t.Fatalf("modified session's updatedAt should return date string")
		}
	})

	t.Run("commit not modified session", func(t *testing.T) {
		sess := New(nil, nil, &Options{
			Store: store,
		})
		if sess.Commit() != nil {
			t.Fatalf("sync not modified session should noop")
		}
	})

	t.Run("commit session first created", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		_, err := sess.Fetch()
		if err != nil {
			t.Fatalf("fetch sesion fail, %v", err)
		}
		sess.Set("name", "tree.xie")
		err = sess.Commit()
		if err != nil {
			t.Fatalf("commit session fail, %v", err)
		}
		values := w.HeaderMap["Set-Cookie"]
		if len(values) != 2 {
			t.Fatalf("first created session should set two cookies")
		}
		sessionID := strings.Split(values[0], "=")[1]
		buf, err := store.Get(sessionID)
		if err != nil {
			t.Fatalf("get session from store fail, %v", err)
		}
		if len(buf) == 0 {
			t.Fatalf("get session from store should not be nil")
		}
	})

	t.Run("session get(type) function", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		w := httptest.NewRecorder()
		sess := New(r, w, &Options{
			Store:      store,
			CookieKeys: keys,
		})
		// before fetch
		if sess.GetBool("exists") || sess.GetString("name") != "" || sess.GetInt("age") != 0 || sess.GetFloat64("count") != 0 || sess.GetStringSlice("category") != nil {
			t.Fatalf("get data before fetch fail")
		}
		_, err := sess.Fetch()
		if err != nil {
			t.Fatalf("fetch sesion fail, %v", err)
		}
		sess.data = M{
			"exists": true,
			"name":   "tree.xie",
			"age":    30,
			"count":  10.1,
			"category": []string{
				"a",
				"b",
			},
		}
		if !sess.GetBool("exists") {
			t.Fatalf("get bool data fail")
		}

		if sess.GetString("name") != "tree.xie" {
			t.Fatalf("get string data fail")
		}

		if sess.GetInt("age") != 30 {
			t.Fatalf("get int data fail")
		}

		if sess.GetFloat64("count") != 10.1 {
			t.Fatalf("get float64 data fail")
		}

		if strings.Join(sess.GetStringSlice("category"), ",") != "a,b" {
			t.Fatalf("get string slice fail")
		}
	})
}
