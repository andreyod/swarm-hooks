package authZ

import (
	//	"bytes"
	//	"io/ioutil"
	//	"fmt"
	"net/http"

	"github.com/docker/swarm/cluster"
	//	"github.com/docker/swarm/cluster/swarm"

	"github.com/docker/swarm/pkg/authZ/keystone"

	"github.com/docker/swarm/cluster"

	//	"github.com/docker/swarm/cluster/swarm"

)

//DefaultACLsImpl - Default implementation of ACLs API
type DefaultACLsImpl struct{}

var authZTokenHeaderName = "X-Auth-Token"
var authZTenantIdHeaderName = "X-Auth-TenantId"
var tenancyLabel = "com.swarm.tenant.0"

var keyStoneAPI keystone.KeyStoneAPI

/*
ValidateRequest - Who wants to do what - allow or not
*/
func (*DefaultACLsImpl) ValidateRequest(cluster cluster.Cluster, eventType EventEnum, w http.ResponseWriter, r *http.Request) (ApprovalEnum, string) {
	tokenToValidate := r.Header.Get(authZTokenHeaderName)
	tenantIdToValidate := r.Header.Get(authZTenantIdHeaderName)
	log.Debug("tenantIdToValidate is " + tenantIdToValidate)

	if tokenToValidate == "" {
		return notApproved, ""
	}

	if tenantIdToValidate == "" {
		return notApproved, ""
	}

	tokenAuthorized, tenantId := keyStoneAPI.ValidateToken(tokenToValidate, tenantIdToValidate)
	if !tokenAuthorized {
		log.Debug("token not authorized or tenantId not associated with token")
		return notApproved, ""
	}
	log.Debug("tenantId is " + tenantId)

	//TODO - Duplication revise
	switch eventType {

	case containerCreate:
		log.Debug("case containerCreate ")

		return approved, ""
	case containersList:
		log.Debug("case containersList ")
		return conditionFilter, ""
	default:
		log.Debug("case default ")
		//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
		isOwner, id := checkOwnerShip(cluster, tenantIdToValidate, r)
		if isOwner {
			return approved, id
		}
	}
	return notApproved, ""
}

//Init - Any required initialization
func (*DefaultACLsImpl) Init() error {
	//This is the keyStone version...
	keyStoneAPI := new(keystone.KeyStoneAPI)
	keyStoneAPI.Init()
	return nil
}
