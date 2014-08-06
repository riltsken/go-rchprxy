package reachproxy

import (
  "net/http"
)

type HealthHandler struct {}

func (healthHandler *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("\nIm up!\n"))
}