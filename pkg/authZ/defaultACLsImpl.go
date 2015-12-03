package authZ

import (
	//	"bytes"
	//	"io/ioutil"
	//	"fmt"
	"net/http"

	"github.com/docker/swarm/cluster"
	//	"github.com/docker/swarm/cluster/swarm"
	log "github.com/Sirupsen/logrus"
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
func (*DefaultACLsImpl) ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request, reqBody []byte) (states.ApprovalEnum, *utils.ValidationOutPutDTO) {
	tokenToValidate := r.Header.Get(headers.AuthZTokenHeaderName)

	if tokenToValidate == "" {
		return states.NotApproved, nil
	}
	//TODO - Duplication revise
	switch eventType {
	case states.ContainerCreate:
		valid, dto := utils.CheckLinksOwnerShip(cluster, tokenToValidate, r, reqBody)
		log.Debug(valid)
		log.Debug(dto)
		log.Debug("-----------------")
		return states.Approved, dto
	case states.ContainersList:
		return states.ConditionFilter, nil
	case states.Unauthorized:
		return states.NotApproved, nil
	default:
		//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
		isOwner, dto := utils.CheckOwnerShip(cluster, tokenToValidate, r)
		if isOwner {
			return states.Approved, dto
		}
	}
	return states.NotApproved, nil
}

//Init - Any required initialization
func (*DefaultACLsImpl) Init() error {
	return nil
}
	