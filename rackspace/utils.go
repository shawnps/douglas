package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func Request(method, url, token, path string, tenantId int,
	body io.Reader) (root map[string]interface{}, err error) {
	defer func() {
		if err_ := recover(); err_ != nil {
			err = err_.(error)
		}
	}()
	client := &http.Client{}
	requestUrl := url + strconv.Itoa(tenantId) + path
	log.Printf("Making request URL %s Token %s Method %s\n", requestUrl,
		token, method)
	req, err := http.NewRequest(method, requestUrl, body)
	req.Header.Add("X-Auth-Token", token)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if body != nil {
		log.Printf("Body: %s\n", body)
	}
	responseBody, err := ioutil.ReadAll(resp.Body)
	log.Println("responseBody: " + string(responseBody))
	if len(responseBody) < 1 {
		return
	}
	var f interface{}
	err = json.Unmarshal(responseBody, &f)
	if err != nil {
		log.Fatal(err)
	}
	root_, ok := f.(map[string]interface{})
	if ok == true {
		root = root_
	}
	return
}

func EscapeJsonValue(raw string) string {
	return strings.Replace(raw, `"`, `\"`, -1)
}
