package authentication

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Hooks struct{}

var provisionAPI ProvisionAPI
var configs *Configs

func eventParse(w http.ResponseWriter, r *http.Request, next http.Handler) {
	log.Debug("***************")
	log.Debug(r.RequestURI)

	log.Debug("***************")
	if strings.HasPrefix(r.RequestURI, "/containers") && (strings.Contains(r.RequestURI, "attach") || strings.Contains(r.RequestURI, "exec")) {
		w.Write([]byte("Not supported!"))
	} else if strings.HasPrefix(r.RequestURI, "/containers") && (strings.Contains(r.RequestURI, "create")) {
		defer r.Body.Close()
		reqBody, _ := ioutil.ReadAll(r.Body)
		log.Debug("Old body: " + string(reqBody))
		tenancyLabel := configs.GetConf().TenancyLabel
		newBody := bytes.Replace(reqBody, []byte("{"), []byte("{\"Labels\": {\""+tenancyLabel+"\":\""+r.Header.Get("Label")+"\",\"anotherTenantName\": \"Optional\"},"), 1)
		log.Debug("New body: " + string(newBody))
		newReq := cloneAndModifyRequest(r, bytes.NewReader(newBody))
		next.ServeHTTP(w, newReq)
	} else if strings.HasPrefix(r.RequestURI, "/containers/json") {
		mapS := map[string][]string{"label": []string{configs.GetConf().TenancyLabel + "=" + r.Header.Get("Label")}}
//		mapA := map[string]string{ "label":configs.GetConf().TenancyLabel + "=" + r.Header.Get("Label") }
		mapJ, _ := json.Marshal(mapS)

		log.Debug(string(mapJ))
		log.Debug(string(mapJ))
		log.Debug(string(mapJ))

		newReq := cloneAndModifyRequest(r, bytes.NewReader(mapJ))
		newReq.Header.Set("Content-type", "application/json")
		next.ServeHTTP(w, newReq)

		next.ServeHTTP(w, r)
	} else if strings.HasPrefix(r.RequestURI, "/containers") {
		next.ServeHTTP(w, r)
	} else {
		w.Write([]byte("Not supported!"))
	}
}

func cloneAndModifyRequest(r *http.Request, body io.Reader) *http.Request {
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
	}
	r2.Body = rc
	return r2
}

func (*Hooks) PrePostAuthWrapper(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Executing Pre...")
		if validatetoken(w, r) {
			log.Info("OK token is fine")
			eventParse(w, r, next)
		} else {
			w.Write([]byte("Not Authorized!"))
		}
		log.Info("Executing Post...")
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
	r.Header.Del(configs.GetConf().AuthTokenHeader)
	return true
}

func (*Hooks) Init() {
	//TODO - should use a map for all the Pre . Post function like in primary.go
	provisionAPI = new(KeyStoneAPI)
	errorInit := provisionAPI.Init()
	configs = new(Configs)
	configs.ReadConfigurationFormfile()
	if nil != errorInit {
		log.Error("Got error while provisioning auth api")
		log.Error(errorInit)
	}
	log.Info("Init provision engine OK")
}
