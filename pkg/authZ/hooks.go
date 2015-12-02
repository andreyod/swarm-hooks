package authZ

import (
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/pkg/authZ/keystone"
	"github.com/docker/swarm/pkg/authZ/states"
	"io/ioutil"
	"fmt"
)

//Hooks - Entry point to AuthZ mechanisem
type Hooks struct{}

//TODO  - Hooks Infra for overriding swarm
//TODO  - Take bussiness logic out
//TODO  - Refactor packages
//TODO  - Expand API
//TODO -  Images...
//TODO - https://github.com/docker/docker/pull/15953
//TODO - https://github.com/docker/docker/pull/16331

var authZAPI HandleAuthZAPI
var aclsAPI ACLsAPI
//EventEnum - State of event
//type EventEnum int


//ApprovalEnum - State of approval
//type ApprovalEnum int

//PrePostAuthWrapper - Hook point from primary to the authZ mechanisem
func (*Hooks) PrePostAuthWrapper(cluster cluster.Cluster, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventType := eventParse(r)
		defer r.Body.Close()
		reqBody, _ := ioutil.ReadAll(r.Body)
		isAllowed, containerID, err := aclsAPI.ValidateRequest(cluster, eventType, w, r, reqBody)
		if isAllowed == states.Admin {
			next.ServeHTTP(w, r)
			return
		}
		//TODO - all kinds of conditionals
		if eventType == states.PassAsIs || isAllowed == states.Approved || isAllowed == states.ConditionFilter {
			authZAPI.HandleEvent(eventType, w, r, next, containerID, reqBody)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("%v", err)))
		}
	})
}

func eventParse(r *http.Request) states.EventEnum {
	log.Debug("Got the uri...", r.RequestURI)

	if strings.Contains(r.RequestURI, "/containers") && (strings.Contains(r.RequestURI, "create")) {
		return states.ContainerCreate
	}

	if strings.Contains(r.RequestURI, "/containers/json") {
		return states.ContainersList
	}

	if strings.Contains(r.RequestURI, "/containers") &&
		(strings.Contains(r.RequestURI, "logs") || strings.Contains(r.RequestURI, "attach") || strings.Contains(r.RequestURI, "exec")) {
		return states.StreamOrHijack
	}
	if strings.Contains(r.RequestURI, "/containers") && strings.HasSuffix(r.RequestURI, "/json") {
		return states.ContainerInspect
	}
	if strings.Contains(r.RequestURI, "/containers") {
		return states.ContainerOthers
	}

//	if strings.Contains(r.RequestURI, "Will add to here all APIs we explicitly want to block") {
//		return states.NotSupported
//	}

	return states.NotSupported
}

//Init - Initialize the Validation and Handling APIs
func (*Hooks) Init() {
	//TODO - should use a map for all the Pre . Post function like in primary.go

	aclsAPI = new(keystone.KeyStoneAPI)
	aclsAPI.Init()
	authZAPI = new(DefaultImp)
	authZAPI.Init()
	//TODO reflection using configuration file tring for the backend type

	log.Info("Init provision engine OK")
}
