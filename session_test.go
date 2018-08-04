package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
			Name:  defaultCookieKey,
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
			Name:  defaultCookieKey,
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
			Name:  defaultCookieKey,
			Value: cookieValue,
		}
		sigCookie := &http.Cookie{
			Name:  defaultCookieKey + ".sig",
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
			Name:  defaultCookieKey,
			Value: cookieValue,
		}
		sigCookie := &http.Cookie{
			Name:  defaultCookieKey + ".sig",
			Value: kg.Sign(defaultCookieKey + "=" + cookieValue),
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
	})
}
