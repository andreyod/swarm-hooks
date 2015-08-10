package authentication

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type Hooks struct{}

var provisionAPI ProvisionAPI

func eventParse(w http.ResponseWriter, r *http.Request, next http.Handler) {
	log.Debug(r.RequestURI)
	configs := new(Configs)
	configs.ReadConfigurationFormfile()

	strUri := r.RequestURI
	log.Debug("---------------")
	log.Debug(strUri)
	log.Debug("---------------")
	switch strUri {
	case "/containers/create":
		defer r.Body.Close()
		reqBody, _ := ioutil.ReadAll(r.Body)
		log.Debug("Old body: ")
		log.Debug(string(reqBody))
		tenancyLabel := configs.GetConf().TenancyLabel
		newBody := bytes.Replace(reqBody, []byte("{"), []byte("{\"Labels\": {\""+tenancyLabel+"\":\""+r.Header.Get("Label")+"\",\"anotherTenantName\": \"PUTMEHERE\"},"), 1)
		log.Debug("New body: ")
		log.Debug(string(newBody))
		newReq := cloneAndModifyRequest(r, bytes.NewReader(newBody))
		//		newReq, e1 := http.NewRequest("POST", r.URL.String(), bytes.NewReader(newBody))
		//		if e1!=nil{
		//			log.Error(e1)
		//		}
		next.ServeHTTP(w, newReq)
		//	case "/containers/*/start":
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
			w.Write([]byte("Not Authorized!!!!"))
		}
		log.Info("Executing Post...")
	})
}

func validatetoken(w http.ResponseWriter, r *http.Request) bool {
	tokenToValidate := r.Header.Get("User-token")
	isValid, label := provisionAPI.ValidateToken(tokenToValidate)
	if !isValid {
		log.Info("Not authenticated. Check user toekn.")
		return false
	}
	//TODO - For now tenant id - Later some hash which works with more complex ACL
	r.Header.Add("Label", label)
	r.Header.Del("User-token")
	return true
}

func (*Hooks) Init() {
	//TODO - should use a map for all the Pre . Post function like in primary.go
	provisionAPI = new(KeyStoneAPI)
	errorInit := provisionAPI.Init()
	if nil != errorInit {
		log.Error("Got error while provisioning auth api")
		log.Error(errorInit)
	}
	log.Info("Init provision engine OK")
}
