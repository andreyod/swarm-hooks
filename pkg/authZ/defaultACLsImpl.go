package authZ

import (
	//	"bytes"
	//	"io/ioutil"
	//	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	//	"github.com/docker/swarm/cluster/swarm"
	"strings"

	"github.com/gorilla/mux"
)

//DefaultACLsImpl - Default implementation of ACLs API
type DefaultACLsImpl struct{}

var authZTokenHeaderName = "X-Auth-Token"
var tenancyLabel = "com.swarm.tenant.0"

/*
ValidateRequest - Who wants to do what - allow or not
*/
func (*DefaultACLsImpl) ValidateRequest(cluster cluster.Cluster, eventType eventEnum, w http.ResponseWriter, r *http.Request) (approvalEnum, string) {
	tokenToValidate := r.Header.Get(authZTokenHeaderName)

	if tokenToValidate == "" {
		return notApproved, ""
	}
	//TODO - Duplication revise
	switch eventType {
	case containerCreate:
		return approved, ""
	case containersList:
		return conditionFilter, ""
	case unauthorized:
		return notApproved, ""
	default:
		//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
		isOwner, id := checkOwnerShip(cluster, tokenToValidate, r)
		if isOwner {
			return approved, id
		}
	}
	return notApproved, ""
}

//TODO - Pass by ref ?
func checkOwnerShip(cluster cluster.Cluster, tenantName string, r *http.Request) (bool, string) {
	containers := cluster.Containers()
	log.Debug("got name: ", mux.Vars(r)["name"])
	tenantSet := make(map[string]bool)
	for _, container := range containers {
		if "/"+mux.Vars(r)["name"]+tenantName == container.Info.Name {
			log.Debug("Match By name!")
			return true, container.Info.Id
		} else if mux.Vars(r)["name"] == container.Info.Id {
			log.Debug("Match By full ID! Checking Ownership...")
			log.Debug("Tenant name: ", tenantName)
			log.Debug("Tenant Lable: ", container.Labels[tenancyLabel])
			if container.Labels[tenancyLabel] == tenantName {
				return true, container.Info.Id
			}
			return false, ""

		}
		if container.Labels[tenancyLabel] == tenantName {
			tenantSet[container.Id] = true
		}
	}

	//Handle short ID
	ambiguityCounter := 0
	var returnID string
	for k := range tenantSet {
		if strings.HasPrefix(cluster.Container(k).Info.Id, mux.Vars(r)["name"]) {
			ambiguityCounter++
			returnID = cluster.Container(k).Info.Id
		}
		if ambiguityCounter == 1 {
			log.Debug("Matched by short ID")
			return true, returnID
		}
		if ambiguityCounter > 1 {
			log.Debug("Ambiguiy by short ID")
			//TODO - ambiguity
		}
		if ambiguityCounter == 0 {
			log.Debug("No match by short ID")
			//TODO - no such container
		}
	}
	return false, ""
}

//Init - Any required initialization
func (*DefaultACLsImpl) Init() error {
	return nil
}
