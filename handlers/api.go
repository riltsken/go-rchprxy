package reachproxy

import (
  "encoding/json"
  "bytes"
  "fmt"
  "net/http"
  "regexp"
  "log"
  "io/ioutil"
  "strings"
)

type ServiceCatalogJson struct {
  Access struct {
    ServiceCatalog []ServiceCatalogItem
  }
}

type ServiceCatalogItem struct {
  Name string
  Type string
  Endpoints []ServiceCatalogEndpoint
}

type ServiceCatalogEndpoint struct {
  Region string
  PublicUrl string
}

type ApiHandler struct {
  requests *http.Client
  proxyRegex *regexp.Regexp
}

func NewApiHandler() *ApiHandler {
  return &ApiHandler{
    requests: &http.Client{},
    proxyRegex: regexp.MustCompile("/([A-Za-z:-]+,[A-Za-z]+,?[A-Za-z]{0,3})/(.*)")}
}

func (apiHandler *ApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  apiUrl := apiHandler.getUrl(r)

  apiRequest, _ := http.NewRequest(r.Method, apiUrl, nil)
  apiRequest.Header.Add("X-Auth-Token", r.Header["X-Auth-Token"][0])
  apiRequest.Header.Add("Accept", "application/json")

  log.Printf("Sending request: %s", apiRequest)

  apiResponse, _ := apiHandler.requests.Do(apiRequest)

  log.Printf("Receiving response: %s", apiResponse)

  if apiResponse != nil {
    body, _ := ioutil.ReadAll(apiResponse.Body)
    apiResponse.Body.Close()
    w.Write([]byte(body))
  } else {
    w.WriteHeader(http.StatusServiceUnavailable)
    w.Write([]byte("{\"message\": \"Request Failed\"}"))
  }

}

func (apiHandler *ApiHandler) getUrl(r *http.Request) (string) {
  matches := apiHandler.proxyRegex.FindStringSubmatch(r.URL.Path)
  catalogId, path := matches[1], matches[2]

  endpoint := apiHandler.getEndpointFromCatalog(r, catalogId)

  return endpoint + "/" + path
}

func (apiHandler *ApiHandler) getEndpointFromCatalog(r *http.Request, catalogId string) (string) {
  requestBody := fmt.Sprintf("{\"auth\": {\"tenantId\": \"%s\", \"token\": {\"id\": \"%s\"}}}",
    r.Header["X-Auth-User"][0], r.Header["X-Auth-Token"][0])
  requestBodyBuffer := bytes.NewBufferString(requestBody)
  serviceCatalogRequest, _ := http.NewRequest("POST", "https://identity.api.rackspacecloud.com/v2.0/tokens", requestBodyBuffer)
  serviceCatalogRequest.Header.Add("Content-Type", "application/json")

  response, _ := apiHandler.requests.Do(serviceCatalogRequest)
  body, _ := ioutil.ReadAll(response.Body)
  response.Body.Close()

  responseJson := &ServiceCatalogJson{}
  json.Unmarshal(body, &responseJson)

  customerItem := ServiceCatalogItem{
    Name: "customer",
    Type: "customer",
    Endpoints: []ServiceCatalogEndpoint{{Region: "", PublicUrl: "https://some.api.com/v1"}}}

  responseJson.Access.ServiceCatalog = append(responseJson.Access.ServiceCatalog, customerItem)

  splitId := strings.Split(catalogId, ",")
  requestedType, requestedName := splitId[0], splitId[1]

  serviceEndpoint := ""
  for _, catalog := range responseJson.Access.ServiceCatalog {
    if catalog.Name == requestedName && catalog.Type == requestedType {
      if len(splitId) == 2 { // only 1 region
        serviceEndpoint = catalog.Endpoints[0].PublicUrl
        break
      }

      for _, endpoint := range catalog.Endpoints {
        if endpoint.Region == splitId[2] { // find the specific region endpoint
          serviceEndpoint = endpoint.PublicUrl
          break
        }
      }
    }
  }

  return serviceEndpoint
}
