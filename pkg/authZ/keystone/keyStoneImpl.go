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
	"github.com/docker/swarm/pkg/authZ/headers"
	"github.com/docker/swarm/pkg/authZ/utils"
	"github.com/jeffail/gabs"
	"strconv"
	"fmt"
	"errors"
)

type KeyStoneAPI struct{quotaAPI QuotaAPI}

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

func (this *KeyStoneAPI) ValidateRequest(cluster cluster.Cluster, eventType states.EventEnum, w http.ResponseWriter, r *http.Request, reqBody []byte) (states.ApprovalEnum, string) {

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

			if isAdminTenant(tenantIdToValidate) {
				return states.Admin, ""
			}
			log.Info("tenantId Found: ")
			//TODO - maybe extract code?
			switch eventType {
			case states.ContainerCreate:
				err := this.validateQuota(cluster, reqBody, tenantIdToValidate)
				if err != nil{
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(fmt.Sprintf("%v", err)))
					return states.NotApproved, ""
				}
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

func (this *KeyStoneAPI) validateQuota(cluster cluster.Cluster, reqBody []byte, tenant string) error {
	log.Info("Going to validate quota")
	log.Debug("Parsing requiered memory field")
	var fieldType float64
	res, err := utils.ParseField("HostConfig.Memory", fieldType, reqBody)
	if err != nil{
		log.Error("Failed to parse mandatory memory limit in container config")
		return errors.New("Failed to parse mandatory memory limit from container config")
	}

	memory := res.(float64)
	log.Debug("Memory field: ", strconv.FormatFloat(memory, 'f', -1, 64))

	return this.quotaAPI.ValidateQuota(cluster, tenant, memory)
}

func isAdminTenant(tenantIdToValidate string) bool {
	//Kenneth - Determine who is admin using keystone...
	return false
}
