package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

var dummyParameters string
var dummyError error
var dummyNextCalled bool

type dummyHandler struct{}

func (d *dummyHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	dummyNextCalled = true
}

func dummyValidate(jwt string) (string, error) {
	dummyParameters = jwt

	return "", dummyError
}

func setupAuthMiddleware() (*httptest.ResponseRecorder, *http.Request, *JWTAuthMiddleware) {
	logger := hclog.Default()
	dummyParameters = ""
	dummyError = nil
	dummyNextCalled = false

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	h := &JWTAuthMiddleware{logger, &dummyHandler{}, dummyValidate}

	return rw, r, h
}

func TestNoJWTReturns401(t *testing.T) {
	rw, r, h := setupAuthMiddleware()

	h.Handle(rw, r)

	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestAuthHeaderNotJWTReturns401(t *testing.T) {
	rw, r, h := setupAuthMiddleware()
	r.Header.Add("Authorization", "basic abc:123")

	h.Handle(rw, r)

	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestInvalidJWTReturns401(t *testing.T) {
	rw, r, h := setupAuthMiddleware()
	dummyError = fmt.Errorf("Invalid JWT")
	r.Header.Add("Authorization", "jwt abc:123")

	h.Handle(rw, r)

	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestValidJWTCallsNext(t *testing.T) {
	rw, r, h := setupAuthMiddleware()
	r.Header.Add("Authorization", "jwt abc:123")

	h.Handle(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.True(t, dummyNextCalled)
}