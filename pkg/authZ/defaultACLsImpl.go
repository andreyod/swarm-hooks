package authZ

import (
	//	"bytes"
	//	"io/ioutil"
	//	"fmt"
	"net/http"

	"github.com/docker/swarm/cluster"
	//	"github.com/docker/swarm/cluster/swarm"

)

//DefaultACLsImpl - Default implementation of ACLs API
type DefaultACLsImpl struct{}

var authZTokenHeaderName = "X-Auth-Token"
var tenancyLabel = "com.swarm.tenant.0"

/*
ValidateRequest - Who wants to do what - allow or not
*/
func (*DefaultACLsImpl) ValidateRequest(cluster cluster.Cluster, eventType EventEnum, w http.ResponseWriter, r *http.Request) (ApprovalEnum, string) {
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

//Init - Any required initialization
func (*DefaultACLsImpl) Init() error {
	return nil
}
