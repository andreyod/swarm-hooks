package authentication

import (
	"bytes"
	//	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	//	"time"

	"github.com/jeffail/gabs"
)

//TODO - Consider moving to gophercloud or other Go client...
//Documentation is out dated etc.

//TODO - Take constant and configuration from outside (consul, etcd, zookeeper,...)

type KeyStoneAPI struct{}

var cacheAPI *Cache

//TODO - May be better to return just the body
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

	//	defer resp.Body.Close()
	//	log.Println("response Status:", resp.Status)
	//	log.Println("response Headers:", resp.Header)
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	log.Println("response Body:", string(body))
	return resp
}

func (*KeyStoneAPI) Init() error {
	//This is just for conviniance when workig with ephemeral keystone
	//TODO - Remove this when moving to keystone which has data
	//	createUserJoe()
	//	createUserMoe()
	cacheAPI = new(Cache)
	cacheAPI.Init()

	return nil
}

type AuthenticationResponse struct {
	access TokenResponseData
}

type TokenResponseData struct {
	issued_at string
	expires   string
	id        string
}

//TODO - May want to sperate concenrns
// 1- Validate Token
// 2- Get ACLs or Lable for your valid token

func (*KeyStoneAPI) ValidateToken(token string) (bool, string) {
	log.Info("Going to validate token: " + token)
	log.Info("Checking cache...")

	tenantId := cacheAPI.Get(token)
	if tenantId != "" {
		return true, tenantId
	}
	configs := new(Configs)
	configs.ReadConfigurationFormfile()

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

	//	something := jsonParsed.Path("tenants.id")

	children, _ := jsonParsed.S("tenants").Children()

	tenantId = children[0].Path("id").Data().(string)

	log.Info(tenantId)

	//TODO - Not using cache since we now dont have the expiration time one the token

	//	userID, _ := jsonParsed.Path("access.user.id").Data().(string)
	//	log.Println(userID)

	//	respGetUser := doHTTPreq("GET", configs.GetConf().KeystoneUrl+"users/"+userID, "", headers)
	//	defer respGetUser.Body.Close()
	//	log.Println("response Status:", respGetUser.Status)
	//	log.Println("response Headers:", respGetUser.Header)
	//	bodyGetUser, _ := ioutil.ReadAll(respGetUser.Body)
	//	log.Println("response Body:", string(bodyGetUser))

	//	jsonParsedUser, _ := gabs.ParseJSON(bodyGetUser)
	//	tenantId, _ = jsonParsedUser.Path("user.tenantId").Data().(string)
	//	log.Printf("tenant id: " + tenantId)

	//	issuedAt, _ := jsonParsed.Path("access.token.issued_at").Data().(string)
	//	expires, _ := jsonParsed.Path("access.token.expires").Data().(string)

	//	log.Println(issuedAt)
	//	log.Println(expires)

	//	n1, en1 := time.Parse("2015-07-19T08:58:37.870608", issuedAt)
	//	n2, en2 := time.Parse("2015-07-19T09:58:37Z", expires)
	//	log.Println(en1)
	//	log.Println(en2)

	//	log.Println("token expiration times...")
	//	ex := n2.Sub(n1).Nanoseconds() / 1000
	//	cacheAPI.PutEx(token, tenantId, ex)
	return true, tenantId

}
