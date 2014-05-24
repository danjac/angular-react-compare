package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"net/http"
)

var store *sessions.CookieStore

func init() {
	// TBD: grab key from environ
	store = sessions.NewCookieStore([]byte("secret-key"))
}

type RequestContext struct {
	Response http.ResponseWriter
	Request  *http.Request
	Vars     map[string]string
	CSRF     *CSRF
}

func (ctx *RequestContext) GetSession(name string) (*sessions.Session, error) {
	return store.Get(ctx.Request, name)
}

func (ctx *RequestContext) SaveSession(session *sessions.Session) error {
	return session.Save(ctx.Request, ctx.Response)
}

func (ctx *RequestContext) Var(name string) string {
	return ctx.Vars[name]
}

func (ctx *RequestContext) DecodeJSON(value interface{}) error {
	return json.NewDecoder(ctx.Request.Body).Decode(value)
}

func (ctx *RequestContext) RenderJSON(status int, value interface{}) {
	ctx.Response.WriteHeader(status)
	ctx.Response.Header().Add("content-type", "application/json")
	json.NewEncoder(ctx.Response).Encode(value)
}

func (ctx *RequestContext) Render(status int, msg string) {
	ctx.Response.WriteHeader(status)
	ctx.Response.Write([]byte(msg))
}

func (ctx *RequestContext) HandleError(err error) {
	http.Error(ctx.Response, err.Error(), http.StatusInternalServerError)
}

func (ctx *RequestContext) CheckCSRF() bool {
	return ctx.CSRF.Validate()
}

func NewRequestContext(w http.ResponseWriter, r *http.Request) *RequestContext {
	ctx := &RequestContext{Response: w, Request: r, Vars: mux.Vars(r)}
	ctx.CSRF = NewCSRF(w, r)
	return ctx
}

type AppHandler func(ctx *RequestContext)

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := NewRequestContext(w, r)
	if !ctx.CheckCSRF() {
		ctx.Render(http.StatusForbidden, "CSRF token missing")
		return
	}
	fn(NewRequestContext(w, r))
}
