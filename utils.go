package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

//
// GetOwnIP retuns the machines external net.IP address
//
func GetOwnIP(dialAddress string) (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr.String())
		if ip == nil {
			continue
		}
		if IsPublicIP(ip) {
			return ip, nil
		}
	}

	return QueryIP(dialAddress)
}

//
// QueryIP fetches external IP from IPDetectionSite using a GET request and parsing the content
//
func QueryIP(dialAddress string) (net.IP, error) {
	resp, err := http.Get(dialAddress)
	if err != nil {
		return nil, err
	}
	defer func() {
		derr := resp.Body.Close()
		if derr != nil {
			log.Println("ERORR:", derr)
		}
	}()
	ipBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(strings.TrimSpace(string(ipBytes)))
	if ip != nil {
		return ip, nil
	}

	return nil, errors.New("unable to detect external IP address")
}

//
// IsPublicIP returns true for a public IP beeing passed
//
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
