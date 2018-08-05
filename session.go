package session

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/oklog/ulid"
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
		GenID func(string) string
		// the cookie value's prefix
		CookiePrefix string
		// key list for keygrip
		CookieKeys []string
		// cookie path
		CookiePath string
		// cookie domain
		CookieDomain string
		// cookie expires
		CookieExpires time.Time
		// cookie max age
		CookieMaxAge int
		// cookie secure
		CookieSecure bool
		// cookie http only
		CookieHttpOnly bool
	}
	// Session session struct
	Session struct {
		Request  *http.Request
		Response http.ResponseWriter
		opts     *Options
		cookies  *cookies.Cookies
		signed   bool
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

// Fetch fetch the session data from store
func (sess *Session) Fetch() (m M, err error) {
	if sess.fetched {
		m = sess.data
		return
	}
	opts := sess.opts
	cookieName := sess.getCookieName()

	value := sess.cookies.Get(cookieName, sess.signed)
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

// Get get data from session's data
func (sess *Session) Get(key string) interface{} {
	if !sess.fetched {
		return nil
	}
	return sess.data[key]
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
	id := sess.cookieValue
	if id == "" {
		if opts.GenID != nil {
			id = opts.GenID(opts.CookiePrefix)
		} else {
			id = generateID(opts.CookiePrefix)
		}
		sess.cookieValue = id
		cookieName := sess.getCookieName()
		cookie := sess.cookies.CreateCookie(cookieName, id)
		sess.cookies.Set(cookie, sess.signed)
	}
	buf, err := json.Marshal(sess.data)
	if err != nil {
		return
	}
	err = opts.Store.Set(id, buf, opts.MaxAge)
	if err != nil {
		return
	}
	sess.commited = true
	return
}

// generateID gen id
func generateID(prefix string) string {
	t := time.Now()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return prefix + ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// New create a session instance
func New(req *http.Request, res http.ResponseWriter, opts *Options) *Session {
	if opts == nil || opts.Store == nil {
		panic(errors.New("the options for session should not be nil"))
	}
	sess := &Session{
		Request:  req,
		Response: res,
	}
	sess.opts = opts
	sess.cookies = cookies.New(req, res, &cookies.Options{
		Keys:     opts.CookieKeys,
		Path:     opts.CookiePath,
		Domain:   opts.CookieDomain,
		Expires:  opts.CookieExpires,
		MaxAge:   opts.CookieMaxAge,
		Secure:   opts.CookieSecure,
		HttpOnly: opts.CookieHttpOnly,
	})
	if len(opts.CookieKeys) != 0 {
		sess.signed = true
	}
	return sess
}
