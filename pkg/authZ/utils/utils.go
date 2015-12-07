package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"strings"

	"net/http/httptest"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/swarm/cluster"

	"strconv"

	"github.com/docker/swarm/pkg/authZ/headers"
	"github.com/gorilla/mux"
	"github.com/jeffail/gabs"
)

type ValidationOutPutDTO struct {
	ContainerID string
	Links       map[string]string
	//Quota can live here too?
	//What else
}

//UTILS

func ModifyRequest(r *http.Request, body io.Reader, urlStr string, containerID string) (*http.Request, error) {

	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
		r.Body = rc
	}
	if urlStr != "" {
		u, err := url.Parse(urlStr)

		if err != nil {
			return nil, err
		}
		r.URL = u
		mux.Vars(r)["name"] = containerID
	}

	return r, nil
}

func CheckLinksOwnerShip(cluster cluster.Cluster, tenantName string, r *http.Request, reqBody []byte) (bool, *ValidationOutPutDTO) {
	jsonParsed, _ := gabs.ParseJSON(reqBody)

	//TODO - Consider refactor all to use json parse and not regexp and maybe save memory on de duplication
	log.Debug("Checking links...")
	children, _ := jsonParsed.Path("HostConfig.Links").Children()
	containers := cluster.Containers()
	linkSet := make(map[string]string)
	var c int
	var l int
	log.Debug("**************************************************")
	for _, child := range children {
		log.Debug("_________________")
		c++

		pair := child.Data().(string)
		linkPair := strings.Split(pair, ":")
		log.Debug(pair)
		log.Debug(linkPair)
		for _, container := range containers {
			log.Debug(container.Info.Name)
			if "/"+linkPair[0]+tenantName == container.Info.Name || "/"+linkPair[0] == container.Info.Name {
				log.Debug("#####################################")
				linkSet[container.Info.Id] = linkPair[0]
				l++
			}
		}
	}
	log.Debug("**************************************************")
	if l != c {
		//TODO - Change to pointer and return nil
		return false, &ValidationOutPutDTO{ContainerID: "", Links: linkSet}
	}
	v := ValidationOutPutDTO{ContainerID: "", Links: linkSet}
	return true, &v

}

//TODO - Pass by ref ?
func CheckOwnerShip(cluster cluster.Cluster, tenantName string, r *http.Request) (bool, *ValidationOutPutDTO) {
	containers := cluster.Containers()
	log.Debug("got name: ", mux.Vars(r)["name"])
	if mux.Vars(r)["name"] == ""{
		return true, ""
	}
	tenantSet := make(map[string]bool)
	for _, container := range containers {
		if "/"+mux.Vars(r)["name"]+tenantName == container.Info.Name {
			log.Debug("Match By name!")
			return true, &ValidationOutPutDTO{ContainerID: container.Info.Id, Links: nil}
		} else if "/"+mux.Vars(r)["name"] == container.Info.Name {
			if container.Labels[headers.TenancyLabel] == tenantName {
				return true, &ValidationOutPutDTO{ContainerID: container.Info.Id, Links: nil}
			}
		} else if mux.Vars(r)["name"] == container.Info.Id {
			log.Debug("Match By full ID! Checking Ownership...")
			log.Debug("Tenant name: ", tenantName)
			log.Debug("Tenant Lable: ", container.Labels[headers.TenancyLabel])
			if container.Labels[headers.TenancyLabel] == tenantName {
				return true, &ValidationOutPutDTO{ContainerID: container.Info.Id, Links: nil}
			}
			return false, nil

		}
		if container.Labels[headers.TenancyLabel] == tenantName {
			tenantSet[container.Id] = true
		}
	}

	//Handle short ID
	ambiguityCounter := 0
	var returnID string
	for k := range tenantSet {
		if strings.HasPrefix(cluster.Container(k).Info.Id, mux.Vars(r)["name"]) {
			ambiguityCounter++
			returnID = cluster.Container(k).Info.Id
		}
		if ambiguityCounter == 1 {
			log.Debug("Matched by short ID")
			return true, &ValidationOutPutDTO{ContainerID: returnID, Links: nil}
		}
		if ambiguityCounter > 1 {
			log.Debug("Ambiguiy by short ID")
			//TODO - ambiguity
		}
		if ambiguityCounter == 0 {
			log.Debug("No match by short ID")
			//TODO - no such container
		}
	}
	return false, nil
}

func CleanUpLabeling(r *http.Request, rec *httptest.ResponseRecorder) []byte {
	newBody := bytes.Replace(rec.Body.Bytes(), []byte(headers.TenancyLabel), []byte(" "), -1)
	//TODO - Here we just use the token for the tenant name for now so we remove it from the data before returning to user.
	newBody = bytes.Replace(newBody, []byte(r.Header.Get(headers.AuthZTenantIdHeaderName)), []byte(" "), -1)
	newBody = bytes.Replace(newBody, []byte(",\" \":\" \""), []byte(""), -1)
	log.Debug("Got this new body...", string(newBody))
	return newBody
}

func ParseField(field string, fieldType interface{}, body []byte) (interface{}, error) {
	log.Debugf("In parseField, field: %s Request body: %s", field, string(body))
	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Error("failed to parse!")
		return nil, err
	}

	switch v := fieldType.(type) {
	case float64:
		log.Debug("Parsing type: ", v)
		parsedField, ok := jsonParsed.Path(field).Data().(float64)
		if ok {
			res := strconv.FormatFloat(parsedField, 'f', -1, 64)
			log.Debugf("Parsed field: " + res)
			return parsedField, nil
		}
	case []string:
		log.Debug("Parsing type: ", v)
		parsedField, ok := jsonParsed.Path(field).Data().([]string)
		if ok {
			log.Debug(parsedField)
			return parsedField, nil
		}
	default:
		log.Error("Unknown field type to parse")
	}

	return nil, errors.New(fmt.Sprintf("failed to parse field %s from request body %s", field, string(body)))
}
