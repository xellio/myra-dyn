package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Myra-Security-GmbH/signature"
)

//
// GETUrl builds and returns the URL for fetching the DNS records
//
func GETUrl(domainName string) string {
	path := fmt.Sprintf("/%s/rapi/dnsRecords/%s/%d", LANG, domainName, 1)
	return fmt.Sprintf("%s://%s%s", "https", APIHost, path)
}

//
// POSTUrl builds and returns the URL for update
//
func POSTUrl(domainName string) string {
	path := fmt.Sprintf("/%s/rapi/dnsRecords/%s", LANG, domainName)
	return fmt.Sprintf("%s://%s%s", "https", APIHost, path)
}

//
// buildGETRequest builds and returns a new GET *http.Request including signature
//
func buildGETRequest(reqURL string, payload map[string]string) (*http.Request, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return req, err
	}

	reqPL := req.URL.Query()

	for key, value := range payload {
		reqPL.Add(key, value)
	}

	req.URL.RawQuery = reqPL.Encode()

	sig := signature.New(config.Secret, config.APIKey, req)
	return sig.Append()
}

//
// buildPOSTRequest builds and returns a new POST *http.Request including signature
//
func buildPOSTRequest(reqURL string, dnsRecordVO *DNSRecordVO) (*http.Request, error) {
	var req *http.Request
	var err error

	data, err := json.Marshal(dnsRecordVO)
	if err != nil {
		return req, err
	}

	req, err = http.NewRequest("POST", reqURL, bytes.NewBuffer(data))
	if err != nil {
		return req, err
	}

	sig := signature.New(config.Secret, config.APIKey, req)
	request, err := sig.Append()

	return request, err
}
