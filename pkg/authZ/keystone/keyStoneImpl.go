package keystone

import (
	"bytes"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/pkg/authZ/states"
	//	"github.com/docker/swarm/pkg/authZ"
	"errors"

	"github.com/docker/swarm/pkg/authZ/headers"
	"github.com/docker/swarm/pkg/authZ/utils"
)

type KeyStoneAPI struct{ quotaAPI QuotaAPI }

var cacheAPI *Cache

var configs *Configs

func doHTTPreq(reqType, url, jsonBody string, headers map[string]string) *http.Response {
	var req *http.Request = nil
	var err error = nil
	if "" != jsonBody {
		byteStr := []byte(jsonBody)
		data := bytes.NewBuffer(byteStr)
		req, err = http.NewRequest(reqType, url, data)
	} else {
		req, err = http.NewRequest(reqType, url, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

type AuthenticationResponse struct {
	access TokenResponseData
}

type TokenResponseData struct {
	issued_at string
	expires   string
	id        string
}

func (this *KeyStoneAPI) Init() error {
	cacheAPI = new(Cache)
	cacheAPI.Init()
	configs = new(Configs)
	configs.ReadConfigurationFormfile()
	this.quotaAPI = new(QuotaImpl)
	this.quotaAPI.Init()
	return nil
}

//TODO - May want to sperate concenrns
// 1- Validate Token
// 2- Get ACLs or Lable for your valid token
// 3- Set up cache to save Keystone call
func (this *KeyStoneAPI) ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request, reqBody []byte) (states.ApprovalEnum, string, error) {

	tokenToValidate := r.Header.Get(headers.AuthZTokenHeaderName)
	tokenToValidate = strings.TrimSpace(tokenToValidate)
	tenantIdToValidate := r.Header.Get(headers.AuthZTenantIdHeaderName)

	log.Debugf("Going to validate token:  %v, for tenant Id: %v, ", tokenToValidate, tenantIdToValidate)
	valid := queryKeystone(tenantIdToValidate, tokenToValidate)

	if !valid {
		return states.NotApproved, "", errors.New("Keystone validation failed. Not Authorized!")
	}

	if isAdminTenant(tenantIdToValidate) {
		return states.Admin, "", nil
	}

	//SHORT CIRCUIT KEYSTONE
//	tenantIdToValidate = tokenToValidate
	switch eventType {
	case states.ContainerCreate:
		err := this.quotaAPI.ValidateQuota(cluster, reqBody, tenantIdToValidate)
		if err != nil {
			return states.QuotaLimit, "", err
		}
		return states.Approved, "", nil
	case states.ContainersList:
		return states.ConditionFilter, "", nil
	case states.Unauthorized:
		return states.NotApproved, "", nil
	default:
		//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
		//INFO is make
		isOwner, id := utils.CheckOwnerShip(cluster, tenantIdToValidate, r)
		if isOwner {
			return states.Approved, id, nil
		}
	}
	log.Debug("SHOULD NOT BE HERE....")
	return states.NotApproved, "", errors.New("SHOULD NOT BE HERE. Not Authorized!")
}

func isAdminTenant(tenantIdToValidate string) bool {
	//Kenneth - Determine who is admin using keystone...
	return false
}

//SHORT CIRCUIT KEYSTONE
func queryKeystone(tenantIdToValidate string, tokenToValidate string) bool {
	return true
}

/*
func queryKeystone(tenantIdToValidate string, tokenToValidate string) bool {
	var headers = map[string]string{
		headers.AuthZTokenHeaderName: tokenToValidate,
	}
	resp := doHTTPreq("GET", configs.GetConf().KeystoneUrl+"tenants", "", headers)
	defer resp.Body.Close()
	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))
	if 200 != resp.StatusCode {
		return false
	}

	jsonParsed, _ := gabs.ParseJSON(body)
	children, _ := jsonParsed.S("tenants").Children()

	for i := 0; i < len(children); i++ {
		if children[i].Path("id").Data().(string) == tenantIdToValidate {
			return true
		}
	}
	log.Debug("Tenant not found")
	return false
}
*/
