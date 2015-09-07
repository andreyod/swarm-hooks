package authentication

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	//	"net/http/httputil"
	"encoding/json"
	"strings"
	//	"reflect"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	//	"github.com/jeffail/gabs"
)

type Hooks struct{}

var provisionAPI ProvisionAPI
var configs *Configs

type EVENT_ENUM int

const (
	NOT_SUPPORTED EVENT_ENUM = iota
	CONTAINER_CREATE
	CONTAINERS_DETAIL
	CONTAINER_OTHERS
	UNAUTHORIZED
	LOGS
)

func eventParse(originalW http.ResponseWriter, w http.ResponseWriter, r *http.Request, next http.Handler) EVENT_ENUM {

	log.Debug("Got the uri...")
	log.Debug(r.RequestURI)
	log.Debug("---------------------------")
	if strings.Contains(r.RequestURI, "/containers") && (strings.Contains(r.RequestURI, "attach") || strings.Contains(r.RequestURI, "exec")) {
		w.Write([]byte("Not supported!"))
		return NOT_SUPPORTED
	} else if strings.Contains(r.RequestURI, "/containers") && (strings.Contains(r.RequestURI, "create")) {
		defer r.Body.Close()
		reqBody, _ := ioutil.ReadAll(r.Body)
		log.Debug("Old body: " + string(reqBody))
		newBody := bytes.Replace(reqBody, []byte("{"), []byte("{\"Labels\": {\""+configs.GetConf().TenancyLabel+"\":\""+r.Header.Get("Label")+"\"},"), 1)
		log.Debug("New body: " + string(newBody))
		newReq, e1 := cloneAndModifyRequest(r, bytes.NewReader(newBody), "")
		if e1 != nil {
			log.Error(e1)
		}
		next.ServeHTTP(w, newReq)
		return CONTAINER_CREATE
	} else if strings.Contains(r.RequestURI, "/containers/json") {
		var v = url.Values{}
		mapS := map[string][]string{"label": []string{configs.GetConf().TenancyLabel + "=" + r.Header.Get("Label")}}
		filterJSON, _ := json.Marshal(mapS)
		v.Set("filters", string(filterJSON))
		var newQuery string
		if strings.Contains(r.URL.RequestURI(), "?") {
			newQuery = r.URL.RequestURI() + "&" + v.Encode()
		} else {
			newQuery = r.URL.RequestURI() + "?" + v.Encode()
		}
		log.Debug("...New path...")
		log.Debug(newQuery)
		log.Debug("...New path...")
		newReq, e1 := cloneAndModifyRequest(r, nil, newQuery)
		if e1 != nil {
			log.Error(e1)
		}
		next.ServeHTTP(w, newReq)
		return CONTAINERS_DETAIL
	} else if strings.Contains(r.RequestURI, "/containers") {
		name := mux.Vars(r)["name"]
		log.Debug("Got this as name/Id...")
		log.Debug(name)
		log.Debug("Got this as name/Id...")

		//TODO - use better client and use client better:-)
		req, e1 := http.NewRequest("GET", "http://"+r.Host+"/containers/json?all=1", nil)
		req.Header.Set(configs.GetConf().AuthTokenHeader, r.Header.Get(configs.GetConf().AuthTokenHeader))

		client := &http.Client{}
		ownwerShipResp, e1 := client.Do(req)

		if e1 != nil {
			log.Error("Error checking container ownership...", e1)
			w.Write([]byte("Not  sure you are the owner..."))
		} else {
			defer ownwerShipResp.Body.Close()
			contents, err := ioutil.ReadAll(ownwerShipResp.Body)
			if err != nil {
				log.Error("Error checking container ownership...", err)
				w.Write([]byte("\n Not sure you are the owner... \n"))
			} else {
				log.Debug("OwnerShip body....")
				log.Debug(string(contents))
				log.Debug("OwnerShip body....")
				b := []byte(name)
				if bytes.Contains(contents, b) {
					if strings.Contains(r.RequestURI, "logs") {
						next.ServeHTTP(originalW, r)
						return LOGS
					} else {
						next.ServeHTTP(w, r)
					}

				} else {

					w.Write([]byte("\n You are not the owner of that container! \n"))
					return UNAUTHORIZED

				}
			}
		}

		if strings.HasSuffix(r.RequestURI, "/json") {
			return CONTAINERS_DETAIL
		}

		return CONTAINER_OTHERS
	} else {
		w.Write([]byte("\n Not supported! \n"))
		return NOT_SUPPORTED
	}
}

func cloneAndModifyRequest(r *http.Request, body io.Reader, urlStr string) (*http.Request, error) {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	//Put the modified body
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
		r2.Body = rc
	}
	if urlStr != "" {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		r2.URL = u
	}
	return r2, nil
}

func isAdmin(w http.ResponseWriter, r *http.Request) bool {
	//TODO - This should be talking to a Keystone or vault etc this patch is just for testing
	tokenToValidate := r.Header.Get(configs.GetConf().AuthTokenHeader)
	return tokenToValidate == "admin"
}

func (*Hooks) PrePostAuthWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Executing Pre...")

		if isAdmin(w, r) {
			next.ServeHTTP(w, r)
			return
		}
		var eventType EVENT_ENUM
		rec := httptest.NewRecorder()
		if validatetoken(w, r) {
			log.Info("OK token is fine")
			eventType = eventParse(w, rec, r, next)
			//			eventParse(w, r, next)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not Authorized!"))
			return
		}
		log.Info("Executing Post...")

		if eventType == UNAUTHORIZED {
			w.WriteHeader(http.StatusUnauthorized)
		}

		if eventType == CONTAINER_CREATE && bytes.Contains(rec.Body.Bytes(), []byte("Conflict")) {
			w.WriteHeader(http.StatusConflict)
		}

		//TODO - confsider parse body and look for status=!20X and put that...


		w.WriteHeader(rec.Code)
		// we copy the original headers first
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		if eventType == CONTAINERS_DETAIL {

			log.Debug("++++++++++++++++++++++++++++++++")
			log.Debug(configs.GetConf().TenancyLabel)
			log.Debug(r.Header.Get("Label"))
			log.Debug("++++++++++++++++++++++++++++++++")

			newBody := bytes.Replace(rec.Body.Bytes(), []byte(configs.GetConf().TenancyLabel), []byte(" "), -1)
			newBody = bytes.Replace(newBody, []byte(r.Header.Get("Label")), []byte(" "), -1)
			newBody = bytes.Replace(newBody, []byte(",\" \":\" \""), []byte(""), -1)

			log.Debug("Got this...")
			log.Debug(string(newBody))
			log.Debug("Got this...")
			w.Write(newBody)
		} else {
			w.Write(rec.Body.Bytes())
		}

	})
}

func validatetoken(w http.ResponseWriter, r *http.Request) bool {

	tokenToValidate := r.Header.Get(configs.GetConf().AuthTokenHeader)
	isValid, label := provisionAPI.ValidateToken(tokenToValidate)
	if !isValid {
		log.Info("Not authenticated. Check user token.")
		return false
	}
	//TODO - For now tenant id - Later some hash which works with more complex ACL
	r.Header.Add("Label", label)
	//	r.Header.Del(configs.GetConf().AuthTokenHeader)
	return true
}

func (*Hooks) Init() {
	//TODO - should use a map for all the Pre . Post function like in primary.go
	//	provisionAPI = new(KeyStoneAPI)
	provisionAPI = new(MockAPI)
	//TODO reflection using configuration file tring for the backend type
	errorInit := provisionAPI.Init()
	configs = new(Configs)
	configs.ReadConfigurationFormfile()
	if nil != errorInit {
		log.Error("Got error while provisioning auth api")
		log.Error(errorInit)
	}
	log.Info("Init provision engine OK")
}
