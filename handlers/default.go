package reachproxy

import (
  "log"
  "net/http"
)

type DefaultHandler struct {
  action http.Handler
}

func NewHandler(action http.Handler) *DefaultHandler {
  return &DefaultHandler{action}
}

func (defaultHandler *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%v: %v headers-%v", r.Method, r.URL.Path, r.Header)
  defaultHandler.action.ServeHTTP(w, r)
}