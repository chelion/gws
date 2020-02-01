package session
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"time"
	"errors"
	"github.com/chelion/gws/utils"
	"github.com/chelion/gws/fasthttp"
)


type ProviderConfig interface {
	Name() string
}

type Provider interface {
	Init(int64,ProviderConfig) error
	GC()
	NeedGC()bool
	ReadStore(string) (SessionStore, error)
	Regenerate(string, string) (SessionStore, error)
	Destroy(string) error
	Count() int
}

type Session struct {
	provider Provider
	config   *Config
	cookie   *Cookie
}

type SessionStore interface {
	Set(key string, value []byte)
	Get(key string) []byte
	GetAll() map[string][]byte
	Delete(key string)
	GetSessionId() string
	Save()error 
	Flush()
}

var providers = make(map[string]Provider)

func Register(providerName string, provider Provider) {
	if providers[providerName] != nil {
		panic("session register error, provider " + providerName + " already registered!")
	}
	if provider == nil {
		panic("session register error, provider " + providerName + " is nil!")
	}
	providers[providerName] = provider
}

func GetProvider(name string) (Provider, error) {
	provider, ok := providers[name]
	if !ok {
		return nil, errors.New("session: unknown provide "+name+" not exsits!")
	}
	return provider, nil
}

// return new Session
func NewSession(cfg *Config) *Session {

	if cfg.CookieName == "" {
		cfg.CookieName = defaultCookieName
	}
	if cfg.GCLifetime == 0 {
		cfg.GCLifetime = defaultGCLifetime
	}
	if cfg.SessionLifetime == 0 {
		cfg.SessionLifetime = cfg.GCLifetime
	}
	if cfg.SessionIdGeneratorFunc == nil {
		cfg.SessionIdGeneratorFunc = cfg.defaultSessionIdGenerator
	}

	session := &Session{
		config: cfg,
		cookie: NewCookie(),
	}

	return session
}

// set session provider and provider config
func (s *Session) SetProvider(providerName string, config ProviderConfig) error {
	provider, ok := providers[providerName]
	if !ok {
		return errors.New("session set provider error, " + providerName + " not registered!")
	}
	err := provider.Init(s.config.SessionLifetime, config)
	if err != nil {
		return err
	}
	s.provider = provider

	// start gc
	if s.provider.NeedGC() {
		go func() {
			defer func() {
				e := recover()
				if e != nil {
					panic(errors.New(fmt.Sprintf("session gc crash, %v", e)))
				}
			}()
			s.gc()
		}()
	}
	return nil
}

// start session gc process.
func (s *Session) gc() {
	for {
		select {
		case <-time.After(time.Duration(s.config.GCLifetime) * time.Second):
			s.provider.GC()
		}
	}
}

// session start
// 1. get sessionId from fasthttp ctx
// 2. if sessionId is empty, generator sessionId and set response Set-Cookie
// 3. return session provider store
func (s *Session) Start(ctx *fasthttp.RequestCtx) (store SessionStore, err error) {

	if s.provider == nil {
		return store, errors.New("session start error, not set provider")
	}

	sessionId := s.GetSessionId(ctx)
	if len(sessionId) == 0 {
		// new generator session id
		sessionId = s.config.SessionIdGenerator()
		if len(sessionId) == 0 {
			return store, errors.New("session generator sessionId is empty")
		}
	}
	// read provider session store
	store, err = s.provider.ReadStore(sessionId)
	if err != nil {
		return
	}

	// encode cookie value
	encodeCookieValue := s.config.Encode(sessionId)

	// set response cookie
	s.cookie.Set(ctx,
		s.config.CookieName,
		encodeCookieValue,
		s.config.Domain,
		s.config.Expires,
		s.config.Secure)

	if s.config.SessionIdInHttpHeader {
		ctx.Request.Header.Set(s.config.SessionNameInHttpHeader, sessionId)
		ctx.Response.Header.Set(s.config.SessionNameInHttpHeader, sessionId)
	}

	return
}

// get session id
// 1. get session id by reading from cookie
// 2. get session id from query
// 3. get session id from http headers
func (s *Session) GetSessionId(ctx *fasthttp.RequestCtx) string {

	cookieByte := ctx.Request.Header.Cookie(s.config.CookieName)
	if len(cookieByte) > 0 {
		return s.config.Decode(utils.Bytes2String(cookieByte))
	}

	if s.config.SessionIdInURLQuery {
		cookieFormValue := ctx.FormValue(s.config.SessionNameInUrlQuery)
		if len(cookieFormValue) > 0 {
			return s.config.Decode(utils.Bytes2String(cookieFormValue))
		}
	}

	if s.config.SessionIdInHttpHeader {
		cookieHeader := ctx.Request.Header.Peek(s.config.SessionNameInHttpHeader)
		if len(cookieHeader) > 0 {
			return s.config.Decode(utils.Bytes2String(cookieHeader))
		}
	}

	return ""
}

// regenerate a session id for this SessionStore
func (s *Session) Regenerate(ctx *fasthttp.RequestCtx) (store SessionStore, err error) {

	if s.provider == nil {
		return store, errors.New("session regenerate error, not set provider")
	}

	// generator new session id
	sessionId := s.config.SessionIdGenerator()
	if len(sessionId) == 0 {
		return store, errors.New("session generator sessionId is empty")
	}
	// encode cookie value
	encodeCookieValue := s.config.Encode(sessionId)

	oldSessionId := s.GetSessionId(ctx)
	// regenerate provider session store
	if len(oldSessionId) != 0{
		store, err = s.provider.Regenerate(oldSessionId, sessionId)
	} else {
		store, err = s.provider.ReadStore(sessionId)
	}
	if err != nil {
		return
	}

	// reset response cookie
	s.cookie.Set(ctx,
		s.config.CookieName,
		encodeCookieValue,
		s.config.Domain,
		s.config.Expires,
		s.config.Secure)

	// reset http header
	if s.config.SessionIdInHttpHeader {
		ctx.Request.Header.Set(s.config.SessionNameInHttpHeader, sessionId)
		ctx.Response.Header.Set(s.config.SessionNameInHttpHeader, sessionId)
	}

	return
}


func (s *Session) Destroy(ctx *fasthttp.RequestCtx) {

	// delete header if sessionId in http Header
	if s.config.SessionIdInHttpHeader {
		ctx.Request.Header.Del(s.config.SessionNameInHttpHeader)
		ctx.Response.Header.Del(s.config.SessionNameInHttpHeader)
	}

	cookieValue := s.cookie.Get(ctx, s.config.CookieName)
	if len(cookieValue) == 0 {
		return
	}

	sessionId := s.config.Decode(cookieValue)
	s.provider.Destroy(sessionId)

	// delete cookie by cookieName
	s.cookie.Delete(ctx, s.config.CookieName)
}

func (s *Session)GetSessionNum()int{
	return s.provider.Count()
}