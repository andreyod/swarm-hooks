package authZ

import (
	"net/http"

	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/pkg/authZ/states"
)

//ACLsAPI - API for backend ACLs services - for now only tenant seperation - finer grained later
type ACLsAPI interface {

	//The Admin should first provision itself before starting to servce
	Init() error

	//Is valid and the label for the token if it is valid.
	//TODO - expand response according to design
	ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request, reqBody []byte) (states.ApprovalEnum, string, error)
}
