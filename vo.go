package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	// RecordTypeA ...
	RecordTypeA = "A"
	// RecordTypeAAAA ...
	RecordTypeAAAA = "AAAA"
)

//
// QueryVO ...
//
type QueryVO struct {
	Error    bool           `json:"error"`
	List     []*DNSRecordVO `json:"list"`
	Page     int            `json:"page"`
	Count    int            `json:"count"`
	PageSize int            `json:"pageSize"`
}

//
// ResultVO ...
//
type ResultVO struct {
	Error         bool           `json:"error"`
	ViolationList []*ViolationVO `json:"violationList"`
	TargetObject  []*DNSRecordVO `json:"targetObject"`
}

//
// DNSRecordVO ...
//
type DNSRecordVO struct {
	ObjectType       string           `json:"objectType"`
	ID               int              `json:"id"`
	Modified         string           `json:"modified"`
	Created          string           `json:"created"`
	Name             string           `json:"name"`
	Value            string           `json:"value"`
	Priority         int              `json:"priority"`
	TTL              int              `json:"ttl"`
	RecordType       string           `json:"recordType"`
	Active           bool             `json:"active"`
	Enabled          bool             `json:"enabled"`
	Paused           bool             `json:"paused"`
	UpstreamOptions  *UpstreamOptions `json:"upstreamOptions"`
	AlternativeCNAME string           `json:"alternativeCname"`
	CAAFlags         int              `json:"caaFlags"`
	Comment          string           `json:"comment"`
}

//
// UpstreamOptions ...
//
type UpstreamOptions struct {
	Backup      bool   `json:"backup"`
	Down        bool   `json:"down"`
	FailTimeout string `json:"failTimeout"`
	MaxFails    int    `json:"maxFails"`
	Weight      int    `json:"weight"`
}

//
// ViolationVO ...
//
type ViolationVO struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

//
// Update the DNSRecord ...
//
func (r *DNSRecordVO) Update(ip net.IP, domainName string) (ResultVO, error) {
	resultVO := ResultVO{}

	// todo: update only A with v4 address and AAAA with v6! (otherwise change/set type, too!)
	log.Printf("Update %s to %s for DNSRecord %s-%s\n", r.Value, ip.String(), r.Name, r.RecordType)

	if ip.To4() == nil {
		r.RecordType = RecordTypeAAAA
	} else {
		r.RecordType = RecordTypeA
	}

	r.Comment = fmt.Sprintf("myra-dyn update from %s to %s at %s", r.Value, ip.String(), time.Now().Format(time.RFC3339))
	r.Value = ip.String()

	request, err := buildPOSTRequest(POSTUrl(domainName), r)
	if err != nil {
		return resultVO, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return resultVO, err
	}

	defer func() {
		derr := resp.Body.Close()
		if derr != nil {
			log.Println("ERROR:", derr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return resultVO, errors.New("unable to fetch data")
	}

	err = json.NewDecoder(resp.Body).Decode(&resultVO)
	return resultVO, err
}
