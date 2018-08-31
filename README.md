# session

[![Build Status](https://img.shields.io/travis/vicanso/session.svg?label=linux+build)](https://travis-ci.org/vicanso/session)

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

#### Refresh()

Refresh the session, it just change the updatedAt of session. The function should be called after `Fetch`.

```go
err := sess.Refresh()
```

#### Set(key string, value interface{})

Set the data to session, if `Fetch` isn't called, `ErrNotFetched` will return.

The key `UpdatedAt` will be set to date string of `now` when `Set` call succeed.

```go
err := sess.Set("name", "tree.xie")
```

#### SetMap(value map[string]interface{})

```go
err := sess.Set(map[string]interface{}{
  "a": 1,
  "b": "2",
  "c": true,
})
```

#### Get(key string)

Get the data from session, if `Fetch` isn't called, it will return `nil`.

```go
i := sess.Get("name")
```

#### GetBool(key string)

Get the bool data from session.

```go
exitst := sess.GetBool("exists")
```

#### GetString(key string)

Get the string data from session.

```go
name := sess.GetBool("name")
```

#### GetInt(key string) 

Get the int data from session.

```go
age := sess.GetInt("age")
```

#### GetFloat64(key string)

Get the float64 data from session.

```go
count := sess.GetFloat64("count")
```

#### GetStringSlice(key string)

Get the string slice from session.

```go
category := sess.GetStringSlice("category")
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

#### Destroy

Remove the data from store and reset session data

```go
err := sess.Destroy()
```

## test

go test -race -coverprofile=test.out ./... && go tool cover --html=test.out
