package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

//
// configuration holds information defined in config.yml file
//
type configuration struct {
	APIKey      string `yaml:"apikey"`
	Secret      string `yaml:"secret"`
	Host        string `yaml:"host"`
	Language    string `yaml:"language"`
	DialAddress string `yaml:"ipdetection"`
}

const (
	// APIHost ...
	APIHost = "api.myracloud.com"
	// LANG ...
	LANG = "en"
)

//
// config ...
//
var config *configuration

var configFile string

//
// init ...
//
func init() {

	const (
		defaultConfigFile = "./config.yml"
		usage             = "./myra-dyn --config=/path/to/config.yml example.com"
		usageShort        = "./myra-dyn -c=/path/to/config.yml example.com"
	)
	flag.StringVar(&configFile, "config", defaultConfigFile, usage)
	flag.StringVar(&configFile, "c", defaultConfigFile, usageShort)

	flag.Parse()

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		help(err)
		return
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		help(err)
		return
	}
}

//
// main function
//
func main() {
	domains := os.Args[1:]
	if len(domains) == 0 {
		help(errors.New("missing domain(s)"))
		return
	}

	ip, err := GetOwnIP(config.DialAddress)
	if err != nil {
		log.Println("ERROR", err)
		return
	}

	results, err := fetchResults(domains, ip)
	if err != nil {
		log.Println("ERROR", err)
	}

	var wg sync.WaitGroup

	for domainName, records := range results {
		for dnsName, dnsRecords := range records {
			if len(dnsRecords) > 1 {
				log.Printf("WARNING: Skipping update on %s: More than one A or AAAA record found\n", dnsName)
				continue
			}
			wg.Add(1)

			go func(record *DNSRecordVO, domainName string) {
				defer func() {
					wg.Done()
				}()
				resultVO, err := record.Update(ip, domainName)
				if err != nil {
					log.Println("ERROR", err)
					return
				}

				if resultVO.Error {
					for _, violation := range resultVO.ViolationList {
						log.Println("RESPONSE-ERROR", record.Name, violation.Message)
					}
				}

			}(dnsRecords[0], domainName)
		}
	}
	wg.Wait()

}

//
// fetchResults calls fetch function vor every passed domain and returns a map containing records which require an update
//
func fetchResults(domains []string, ip net.IP) (map[string]map[string][]*DNSRecordVO, error) {
	resultMap := make(map[string]map[string][]*DNSRecordVO)

	for _, domain := range domains {
		resultMap[domain] = make(map[string][]*DNSRecordVO)

		queryVO, err := fetch(domain)
		if err != nil {
			return resultMap, err
		}

		for _, itm := range queryVO.List {
			if itm.Value == ip.String() {
				log.Printf("INFO: No change required for %s (%s)\n", itm.Name, itm.Value)
				continue
			}
			resultMap[domain][itm.Name] = append(resultMap[domain][itm.Name], itm)
		}
	}

	return resultMap, nil

}

//
// fetch exising A and AAAA DNS records for the given domainName
//
func fetch(domainName string) (QueryVO, error) {
	queryVO := QueryVO{}

	recordTypes := strings.Join([]string{RecordTypeA, RecordTypeAAAA}, ",")
	request, err := buildGETRequest(GETUrl(domainName), map[string]string{
		"recordTypes": recordTypes,
		"activeOnly":  "true",
	})
	if err != nil {
		return queryVO, err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return queryVO, err
	}
	defer func() {
		derr := resp.Body.Close()
		if derr != nil {
			log.Println(derr)
		}
	}()

	if resp.StatusCode != http.StatusOK {

		fmt.Println(resp)

		return queryVO, errors.New("unable to fetch data")
	}

	err = json.NewDecoder(resp.Body).Decode(&queryVO)
	return queryVO, err
}

func help(err error) {
	log.Println("ERROR", err)
	fmt.Println(`myra-dyn - dynamic DNS records, protected by Myra
CONFIGURATION:
    Replace the apikey and the secret defined in the config.yml file
    config.yml
        - apikey: 'ADD_YOUR_API_KEY'
        - secret: 'ADD_YOUR_SECRET'
		
BUILDING:
    running 'make' will create a binary in ./bin/ called myra-dyn
        
USAGE:
    Add the domains(!) you want to update using your machines IP address.
    ./bin/myra-dyn your-domain.tld
    ./bin/myra-dyn example1.tld example2.tld example3.tld
        
NOTE:
    - Right now, myra-dyn only works for whole domains. It is not possible to specify a single subdomain for dynamic DNS usage.
    - If your machines IP changes from IPv4 to IPv6 (or the other way around), the DNS recordType will be changed, to (e.g. A instead of AAAA)`)
}
