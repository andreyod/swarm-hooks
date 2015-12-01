package authZ

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"net/http/httptest"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/pkg/authZ/states"
	//	"github.com/docker/swarm/pkg/authZ/keystone"
	"github.com/docker/swarm/pkg/authZ/headers"
	"github.com/docker/swarm/pkg/authZ/utils"
	"github.com/gorilla/mux"
	"github.com/docker/swarm/pkg/authZ/keystone"
	"github.com/docker/swarm/cluster"
	"strconv"
)

//DefaultImp - Default basic label based implementation of ACLs & tenancy enforcment
type DefaultImp struct{
	quotaAPI keystone.QuotaAPI
	}

//Init - Any required initialization
func (this *DefaultImp) Init() error {
	this.quotaAPI = new(keystone.QuotaImpl)
	return nil
}

func (this *DefaultImp) validateQuota(cluster cluster.Cluster, reqBody []byte, tenant string) bool {
	log.Info("Going to validate quota")
	log.Debug("Parsing requiered memory field")
	var fieldType float64
	var memory float64 = utils.ParseField("HostConfig.Memory", fieldType, reqBody).(float64)
	log.Debug("Memory field: ", strconv.FormatFloat(memory, 'f', -1, 64))

	return this.quotaAPI.ValidateQuota(cluster, tenant, memory)
}

//HandleEvent - Implement approved operation - Default labels based implmentation
func (this *DefaultImp) HandleEvent(eventType states.EventEnum, w http.ResponseWriter, r *http.Request, next http.Handler, containerID string, cluster cluster.Cluster) {
	switch eventType {
	case states.ContainerCreate:
		log.Debug("In create...")
		defer r.Body.Close()
		reqBody, _ := ioutil.ReadAll(r.Body)

		if !this.validateQuota(cluster, reqBody, r.Header.Get(headers.AuthZTenantIdHeaderName)){
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Tenant quota limit reached!"))
			return
		}

		log.Debug("Old body: " + string(reqBody))

		//TODO - Here we just use the token for the tenant name for now
		newBody := bytes.Replace(reqBody, []byte("{"), []byte("{\"Labels\": {\""+headers.TenancyLabel+"\":\""+r.Header.Get(headers.AuthZTenantIdHeaderName)+"\"},"), 1)
		log.Debug("New body: " + string(newBody))

		var newQuery string
		if "" != r.URL.Query().Get("name") {
			log.Debug("Postfixing name with Label...")
			newQuery = strings.Replace(r.RequestURI, r.URL.Query().Get("name"), r.URL.Query().Get("name")+r.Header.Get(headers.AuthZTenantIdHeaderName), 1)
			log.Debug(newQuery)
		}

		newReq, e1 := utils.ModifyRequest(r, bytes.NewReader(newBody), newQuery, "")
		if e1 != nil {
			log.Error(e1)
		}
		next.ServeHTTP(w, newReq)

	case states.ContainerInspect:
		log.Debug("In inspect...")
		rec := httptest.NewRecorder()

		r.URL.Path = strings.Replace(r.URL.Path, mux.Vars(r)["name"], containerID, 1)
		mux.Vars(r)["name"] = containerID
		next.ServeHTTP(rec, r)

		/*POST Swarm*/
		w.WriteHeader(rec.Code)
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		newBody := utils.CleanUpLabeling(r, rec)
		w.Write(newBody)

	case states.ContainersList:
		log.Debug("In list...")
		var v = url.Values{}
		mapS := map[string][]string{"label": {headers.TenancyLabel + "=" + r.Header.Get(headers.AuthZTenantIdHeaderName)}}
		filterJSON, _ := json.Marshal(mapS)
		v.Set("filters", string(filterJSON))
		var newQuery string
		if strings.Contains(r.URL.RequestURI(), "?") {
			newQuery = r.URL.RequestURI() + "&" + v.Encode()
		} else {
			newQuery = r.URL.RequestURI() + "?" + v.Encode()
		}
		log.Debug("New Query: ", newQuery)

		newReq, e1 := utils.ModifyRequest(r, nil, newQuery, containerID)
		if e1 != nil {
			log.Error(e1)
		}
		rec := httptest.NewRecorder()

		//TODO - May decide to overrideSwarms handlers.getContainersJSON - this is Where to do it.
		next.ServeHTTP(rec, newReq)

		/*POST Swarm*/
		w.WriteHeader(rec.Code)
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		newBody := utils.CleanUpLabeling(r, rec)

		w.Write(newBody)

	case states.ContainerOthers:
		log.Debug("In others...")
		r.URL.Path = strings.Replace(r.URL.Path, mux.Vars(r)["name"], containerID, 1)
		mux.Vars(r)["name"] = containerID

		next.ServeHTTP(w, r)

		//TODO - hijack and others are the same because we handle no post and no stream manipulation and no handler override yet
	case states.StreamOrHijack:
		log.Debug("In stream/hijack...")
		r.URL.Path = strings.Replace(r.URL.Path, mux.Vars(r)["name"], containerID, 1)
		mux.Vars(r)["name"] = containerID
		next.ServeHTTP(w, r)

	case states.PassAsIs:
		log.Debug("Forwarding the request AS IS...")
		next.ServeHTTP(w, r)
	case states.Unauthorized:
		log.Debug("In UNAUTHORIZED...")
	default:
		log.Debug("In default...")
	}
}
