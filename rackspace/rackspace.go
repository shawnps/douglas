/* From https://github.com/btipling/rose/ */
package rackspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"utils"
)

const (
	NOT_FOUND    = "notFound"
	UNAUTHORIZED = "unauthorized"
	INVALID      = "invalid"
)

var url string = "https://auth.api.rackspacecloud.com/v2.0/tokens"

type AuthResponse struct {
	Name           string
	Token          string
	Id             int
	NovaTenantId   int
	DnsTentantId   int
	FilesTentantId string
}

func Auth(username string, password string) (authResp *AuthResponse, respErr error) {
	body := fmt.Sprintf(`{"auth":{"passwordCredentials":{"username"`+
		`:"%s","password":"%s"}}}`, username, password)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		log.Println("Error with auth: " + err.Error())
		respErr = errors.New(INVALID)
		return
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error with auth: " + err.Error())
		respErr = errors.New(INVALID)
		return
	}
	log.Println("Response body: " + string(responseBody))
	var f interface{}
	err = json.Unmarshal(responseBody, &f)
	if err != nil {
		log.Println("Error parsing response.")
	}
	root := f.(map[string]interface{})
	if _, found := root["unauthorized"].(map[string]interface{}); found == true {
		respErr = errors.New(UNAUTHORIZED)
		return
	}
	if _, found := root["itemNotFound"].(map[string]interface{}); found == true {
		respErr = errors.New(NOT_FOUND)
		return
	}
	if _, found := root["badRequest"].(map[string]interface{}); found == true {
		respErr = errors.New(INVALID)
		return
	}
	authResp = &AuthResponse{}
	access := root["access"].(map[string]interface{})
	tokenMap := access["token"].(map[string]interface{})
	authResp.Token = tokenMap["id"].(string)
	user := access["user"].(map[string]interface{})
	id := user["id"].(string)
	authResp.Id, err = strconv.Atoi(id)
	if err != nil {
		return
	}
	authResp.Name = user["name"].(string)
	serviceCatalog := access["serviceCatalog"].([]interface{})
	for _, service_ := range serviceCatalog {
		service := service_.(map[string]interface{})
		endPoints := service["endpoints"].([]interface{})
		endPoint := endPoints[0].(map[string]interface{})
		tenantIdStr := endPoint["tenantId"].(string)
		tenantId, err := strconv.Atoi(tenantIdStr)
		name := service["name"].(string)
		if err != nil && name != "cloudFiles" {
			continue
		}
		switch name {
		case "cloudDNS":
			authResp.DnsTentantId = tenantId
		case "cloudFiles":
			authResp.FilesTentantId = tenantIdStr
		case "cloudServersOpenStack":
			authResp.NovaTenantId = tenantId
		}
	}
	return
}
