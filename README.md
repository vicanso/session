# session

session for golang, it can use signed cookie by keygrip.

## API

#### New(req *http.Request, res http.ResponseWriter, opts *Options)

Create a cookie instance

- `req` http request
- `res` http response writer
- `opts.Key` cookie key, default is `sess`
- `opts.MaxAge` the max age for session data(seconds)
- `opts.Store` the session store
- `opts.GenID` function to generate session id(cookie's value), if not set, it will use `ulid`.
- `opts.CookiePrefix` add the prefix to the cookie value 
- `opts.CookieKeys` key list for keygrip
- `opts.CookiePath` cookie's path
- `opts.CookieDomain` cookie's domain 
- `opts.CookieExpires` cookie's expires
- `opts.CookieMaxAge` cookie's max age
- `opts.CookieSecure` cookie's secure
- `opts.CookieHttpOnly` cookie's http only


```go
store := NewRedisStore(nil, &redis.Options{
  Addr: "localhost:6379",
})
r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
w := httptest.NewRecorder()
sess := New(r, w, &Options{
  Store: store,
  CookieKeys: []string{
    "tree.xie",
  }
})
```

#### Fetch()

Fetch the session data from redis

```go
data, err := sess.Fetch()
```

#### Set(key string, value interface{})

Set the data to session, if `Fetch` isn't called, `ErrNotFetched` will return.

The key `UpdatedAt` will be set to date string of `now` when `Set` call succeed.

```go
err := sess.Set("name", "tree.xie")
```

#### Get(key string)

Get the data for session, if `Fetch` isn't called, it will return `nil`.

```go
i := sess.Get("name")
```

#### GetCreatedAt

Get the created at field from session, if `Fetch` isn't called, it will be `""`.

```go
createdAt := sess.GetCreatedAt()
```

#### GetUpdatedAt

Get the updated at field from session, if `Fetch` isn't called, it will be `""`.

```go
updatedAt := sess.GetUpdatedAt()
```

#### Commit

Commit the data to store when it's be modified.

If the first time create session, it will set the cookie for session.

```go
store := NewRedisStore(nil, &redis.Options{
  Addr: "localhost:6379",
})
r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
w := httptest.NewRecorder()
sess := New(r, w, &Options{
  Store:      store,
  CookieKeys: []string{
    "tree.xie",
  },
})
_, err := sess.Fetch()
if err != nil {
  fmt.Printf("fetch sesion fail, %v", err)
}
sess.Set("name", "tree.xie")
err = sess.Commit()
if err != nil {
  fmt.Printf("commit sesion fail, %v", err)
}
```


## test

go test -race -coverprofile=test.out ./... && go tool cover --html=test.out
