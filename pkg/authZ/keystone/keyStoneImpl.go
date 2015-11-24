package keystone

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"
	"github.com/docker/swarm/pkg/authZ/states"
//	"github.com/docker/swarm/pkg/authZ"
	"github.com/jeffail/gabs"
	"github.com/docker/swarm/pkg/authZ/utils"
	"github.com/docker/swarm/pkg/authZ/headers"
)

type KeyStoneAPI struct{}

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

func (*KeyStoneAPI) Init() error {
	cacheAPI = new(Cache)
	cacheAPI.Init()
	configs = new(Configs)
	configs.ReadConfigurationFormfile()
	return nil
}

//TODO - May want to sperate concenrns
// 1- Validate Token
// 2- Get ACLs or Lable for your valid token



func (*KeyStoneAPI) ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request) (states.ApprovalEnum, string) {

	tokenToValidate := r.Header.Get(headers.AuthZTokenHeaderName)
	tenantIdToValidate := r.Header.Get(headers.AuthZTenantIdHeaderName)
	log.Info("Going to validate token: " + tokenToValidate)
	log.Info("Going to validate tenantId: " + tenantIdToValidate)

	log.Info("Please set up the cache...")

	var headers = map[string]string{
		"X-Auth-Token": tokenToValidate,
	}
	tokenToValidate = strings.TrimSpace(tokenToValidate)
	resp := doHTTPreq("GET", configs.GetConf().KeystoneUrl+"tenants", "", headers)
	defer resp.Body.Close()
	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))
	if 200 != resp.StatusCode {
		return states.NotApproved, "Invalid user token"
	}
	log.Info("Valid user token!")
	jsonParsed, _ := gabs.ParseJSON(body)
	children, _ := jsonParsed.S("tenants").Children()
	//tenantId = children[i].Path("id").Data().(string)
	for i := 0; i < len(children); i++ {
		if children[i].Path("id").Data().(string) == tenantIdToValidate {
			log.Info("tenantId Found: ")
			//TODO - maybe extract code?
			switch eventType {
			case states.ContainerCreate:
				return states.Approved, ""
			case states.ContainersList:
				return states.ConditionFilter, ""
			case states.Unauthorized:
				return states.NotApproved, ""
			default:
				//CONTAINER_INSPECT / CONTAINER_OTHERS / STREAM_OR_HIJACK / PASS_AS_IS
				isOwner, id := utils.CheckOwnerShip(cluster, tenantIdToValidate, r)
				if isOwner {
					return states.Approved, id
				}
			}
			return states.Approved, tenantIdToValidate
		}
	}
	log.Info("tenantId not Found: ")
	//log.Info(tenantId)
	return states.NotApproved, ""
}
