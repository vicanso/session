package session

import (
	"errors"
	"math/rand"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/oklog/ulid"
	"github.com/spf13/cast"
	"github.com/vicanso/cookies"
)

const (
	defaultCookieName = "sess"
	// CreatedAt the created time for session
	CreatedAt = "_createdAt"
	// UpdatedAt the updated time for session
	UpdatedAt = "_updatedAt"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// ErrNotFetched not fetch error
	ErrNotFetched = errors.New("Not fetch session")
)

type (
	// M alias
	M map[string]interface{}
	// Store session store interface
	Store interface {
		// Get get the session data
		Get(string) ([]byte, error)
		// Set set the session data
		Set(string, []byte, int) error
		// Destroy remove the session data
		Destroy(string) error
	}
	// Options new session options
	Options struct {
		// cookie key, default is sess
		Key string
		// the max age for session data
		MaxAge int
		// the session store
		Store Store
		// function to generate session id
		GenID         func() string
		CookieOptions *cookies.Options
	}
	// Session session struct
	Session struct {
		opts    *Options
		cookies *cookies.Cookies
		signed  bool
		// the session cookie value
		cookieValue string
		// the data fetch from session
		data M
		// the data has been fetched
		fetched bool
		// the data has been modified
		modified bool
		// the sesion has been commited
		commited bool
	}
)

// Mock mock a session
func Mock(data M) *Session {
	sess := &Session{}
	for k, v := range data {
		switch k {
		case "fetched":
			sess.fetched = v.(bool)
		case "modified":
			sess.modified = v.(bool)
		case "commited":
			sess.commited = v.(bool)
		case "signed":
			sess.signed = v.(bool)
		case "cookieValue":
			sess.cookieValue = v.(string)
		case "data":
			sess.data = v.(M)
		}
	}
	return sess
}

func getInitJSON() []byte {
	m := M{}
	m[CreatedAt] = time.Now().Format(time.RFC3339)
	buf, _ := json.Marshal(&m)
	return buf
}

// getCookieName get the cookie's name
func (sess *Session) getCookieName() string {
	opts := sess.opts
	cookieName := opts.Key
	if cookieName == "" {
		cookieName = defaultCookieName
	}
	return cookieName
}

// getCookieValue get the cookie's value
func (sess *Session) getCookieValue() string {
	cookieName := sess.getCookieName()

	return sess.cookies.Get(cookieName, sess.signed)
}

// Fetch fetch the session data from store
func (sess *Session) Fetch() (m M, err error) {
	if sess.fetched {
		m = sess.data
		return
	}
	opts := sess.opts

	value := sess.getCookieValue()
	var buf []byte
	if value != "" {
		sess.cookieValue = value
		buf, err = opts.Store.Get(sess.cookieValue)
		if err != nil {
			return
		}
	}
	if len(buf) == 0 {
		buf = getInitJSON()
	}
	m = make(M)
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return
	}
	sess.fetched = true
	sess.data = m
	return
}

// Destroy remove the data from store and reset session data
func (sess *Session) Destroy() (err error) {
	opts := sess.opts
	value := sess.getCookieValue()
	if value == "" {
		return
	}
	err = opts.Store.Destroy(value)
	if err != nil {
		return
	}
	buf := getInitJSON()
	m := make(M)
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return
	}
	sess.data = m
	return
}

// Set set data to session
func (sess *Session) Set(key string, value interface{}) (err error) {
	if key == "" {
		return
	}
	if !sess.fetched {
		return ErrNotFetched
	}
	if value == nil {
		delete(sess.data, key)
	} else {
		sess.data[key] = value
	}
	sess.data[UpdatedAt] = time.Now().Format(time.RFC3339)
	sess.modified = true
	return
}

// SetMap set map data to session
func (sess *Session) SetMap(value map[string]interface{}) (err error) {
	if value == nil {
		return
	}
	if !sess.fetched {
		return ErrNotFetched
	}
	for k, v := range value {
		if v == nil {
			delete(sess.data, k)
			continue
		}
		sess.data[k] = v
	}

	sess.data[UpdatedAt] = time.Now().Format(time.RFC3339)
	sess.modified = true
	return
}

// Refresh refresh session (update updatedAt)
func (sess *Session) Refresh() (err error) {
	if !sess.fetched {
		return ErrNotFetched
	}
	sess.data[UpdatedAt] = time.Now().Format(time.RFC3339)
	sess.modified = true
	// 刷新cookie的max age
	if sess.cookieValue != "" {
		sess.addSessionCookie(sess.cookieValue)
	}
	return
}

// Get get data from session's data
func (sess *Session) Get(key string) interface{} {
	if !sess.fetched {
		return nil
	}
	return sess.data[key]
}

// GetBool get bool data from session's data
func (sess *Session) GetBool(key string) bool {
	if !sess.fetched {
		return false
	}
	return cast.ToBool(sess.data[key])
}

// GetString get string data from session's data
func (sess *Session) GetString(key string) string {
	if !sess.fetched {
		return ""
	}
	return cast.ToString(sess.data[key])
}

// GetInt get int data from session's data
func (sess *Session) GetInt(key string) int {
	if !sess.fetched {
		return 0
	}
	return cast.ToInt(sess.data[key])
}

// GetFloat64 get float64 data from session's data
func (sess *Session) GetFloat64(key string) float64 {
	if !sess.fetched {
		return 0
	}
	return cast.ToFloat64(sess.data[key])
}

// GetStringSlice get string slice data from session's data
func (sess *Session) GetStringSlice(key string) []string {
	if !sess.fetched {
		return nil
	}
	return cast.ToStringSlice(sess.data[key])
}

// GetCreatedAt get the created at of session
func (sess *Session) GetCreatedAt() string {
	if !sess.fetched {
		return ""
	}
	v := sess.data[CreatedAt]
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetUpdatedAt get the updated at of session
func (sess *Session) GetUpdatedAt() string {
	if !sess.fetched {
		return ""
	}
	v := sess.data[UpdatedAt]
	if v == nil {
		return ""
	}
	return v.(string)
}

// Commit sync the session to store
func (sess *Session) Commit() (err error) {
	if !sess.modified || sess.commited {
		return
	}
	opts := sess.opts
	// not cookie value, create and set cookie
	if sess.cookieValue == "" {
		sess.RegenerateCookie()
	}
	buf, err := json.Marshal(sess.data)
	if err != nil {
		return
	}
	err = opts.Store.Set(sess.cookieValue, buf, opts.MaxAge)
	if err != nil {
		return
	}
	sess.commited = true
	return
}

// RegenerateCookie regenerate the session's cookie
func (sess *Session) RegenerateCookie() {
	if sess.commited {
		return
	}
	opts := sess.opts
	fn := opts.GenID
	if fn == nil {
		fn = generateID
	}
	// id := fn(opts.CookiePrefix)
	id := fn()
	sess.addSessionCookie(id)
}

func (sess *Session) addSessionCookie(value string) {
	sess.cookieValue = value
	cookieName := sess.getCookieName()
	cookie := sess.cookies.CreateCookie(cookieName, value)
	sess.cookies.Set(cookie, sess.signed)
}

// GetData get the session's data
func (sess *Session) GetData() M {
	return sess.data
}

// generateID gen id
func generateID() string {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// New create a session instance
func New(rw cookies.ReadWriter, opts *Options) *Session {
	if opts == nil || opts.Store == nil {
		panic(errors.New("the options for session should not be nil"))
	}
	sess := &Session{}
	sess.opts = opts
	cookieOptions := opts.CookieOptions
	sess.cookies = cookies.New(rw, cookieOptions)
	if cookieOptions != nil && len(cookieOptions.Keys) != 0 {
		sess.signed = true
	}
	return sess
}
