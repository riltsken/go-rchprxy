package main

import (
  "fmt"
  "log"
  "net/http"
  "./handlers"
)

func main() {
  fmt.Printf("starting on port 5500\n")

  proxy := &http.Server{
    Addr: ":5500",
    Handler: createHandlers(),
  }
  log.Fatal(proxy.ListenAndServe())
}

func createHandlers() (*http.ServeMux) {
  h := http.NewServeMux()
  h.Handle("/proxy_health", reachproxy.NewHandler(&reachproxy.HealthHandler{}))
  h.Handle("/", reachproxy.NewHandler(reachproxy.NewApiHandler()))
  return h
}

