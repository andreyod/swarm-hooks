package keystone

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/jeffail/gabs"
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

func (*KeyStoneAPI) ValidateToken(token string) (bool, string) {
	log.Info("Going to validate token: " + token)
	
	log.Info("Please set up the cache...")
	var tenantId string
//	log.Info("Checking cache...")
//	tenantId := cacheAPI.Get(token)
//	if tenantId != "" {
//		return true, tenantId
//	}
	
	var headers = map[string]string{
		"X-Auth-Token": token,
	}
	token = strings.TrimSpace(token)
	resp := doHTTPreq("GET", configs.GetConf().KeystoneUrl+"tenants", "", headers)
	//	resp := doHTTPreq("GET", "http://127.0.0.1:5000/v2.0/tenants", "", headers)
	defer resp.Body.Close()
	log.Debug("response Status:", resp.Status)
	log.Debug("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response Body:", string(body))
	if 200 != resp.StatusCode {
		return false, "Invalid user token"
	}
	log.Info("Valid user token!")
	jsonParsed, _ := gabs.ParseJSON(body)
	children, _ := jsonParsed.S("tenants").Children()
	tenantId = children[0].Path("id").Data().(string)
	log.Info(tenantId)
	return true, tenantId
}
