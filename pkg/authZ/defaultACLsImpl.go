package authZ

import (
	//	"bytes"
	//	"io/ioutil"
	//	"fmt"
	"net/http"

	"github.com/docker/swarm/cluster"
	//	"github.com/docker/swarm/cluster/swarm"

	"github.com/docker/swarm/pkg/authZ/states"
	//	"github.com/docker/swarm/cluster/swarm"
	"github.com/docker/swarm/pkg/authZ/headers"
	"github.com/docker/swarm/pkg/authZ/utils"
)

//DefaultACLsImpl - Default implementation of ACLs API
type DefaultACLsImpl struct{}

/*
ValidateRequest - Who wants to do what - allow or not
*/
func (*DefaultACLsImpl) ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request) (states.ApprovalEnum, string) {
	tokenToValidate := r.Header.Get(headers.AuthZTokenHeaderName)

	if tokenToValidate == "" {
		return states.NotApproved, ""
	}
	//TODO - Duplication revise
	switch eventType {
	case states.ContainerCreate:
		return states.Approved, ""
	case states.ContainersList:
		return states.ConditionFilter, ""
	case states.Unauthorized:
		return states.NotApproved, ""
	default:
		//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
		isOwner, id := utils.CheckOwnerShip(cluster, tokenToValidate, r)
		if isOwner {
			return states.Approved, id
		}
	}
	return states.NotApproved, ""
}

//Init - Any required initialization
func (*DefaultACLsImpl) Init() error {
	return nil
}
