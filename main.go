package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	myrasec "github.com/Myra-Security-GmbH/myrasec-go"

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

const (
	// RecordTypeA ...
	RecordTypeA = "A"
	// RecordTypeAAAA ...
	RecordTypeAAAA = "AAAA"
)

//
// config ...
//
var config *configuration

var configFile string

var api *myrasec.API

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
}

//
// main function
//
func main() {

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

	domains := flag.Args()
	if len(domains) == 0 {
		help(errors.New("missing domain(s)"))
		return
	}

	api, err = myrasec.New(config.APIKey, config.Secret)
	if err != nil {
		help(err)
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

			go func(record myrasec.DNSRecord, domainName string) {
				defer func() {
					wg.Done()
				}()

				log.Printf("Update %s to %s for DNSRecord %s-%s\n", record.Value, ip.String(), record.Name, record.RecordType)

				record.Value = ip.String()

				if ip.To4() == nil {
					record.RecordType = RecordTypeAAAA
				} else {
					record.RecordType = RecordTypeA
				}

				record.Comment = fmt.Sprintf("myra-dyn update from %s to %s at %s", record.Value, ip.String(), time.Now().Format(time.RFC3339))

				_, err = api.UpdateDNSRecord(&record, domainName)
				if err != nil {
					log.Println("ERROR", err)
				}

			}(dnsRecords[0], domainName)
		}
	}
	wg.Wait()

}

//
// fetchResults calls fetch function vor every passed domain and returns a map containing records which require an update
//
func fetchResults(domains []string, ip net.IP) (map[string]map[string][]myrasec.DNSRecord, error) {
	resultMap := make(map[string]map[string][]myrasec.DNSRecord)

	recordTypes := strings.Join([]string{RecordTypeA, RecordTypeAAAA}, ",")
	params := map[string]string{
		"recordTypes": recordTypes,
		"activeOnly":  "true",
		"pageSize":    "1000",
	}
	for _, domain := range domains {
		resultMap[domain] = make(map[string][]myrasec.DNSRecord)

		records, err := api.ListDNSRecords(domain, params)
		if err != nil {
			return resultMap, err
		}

		for _, rec := range records {
			if rec.Value == ip.String() {
				log.Printf("INFO: No change required for %s (%s)\n", rec.Name, rec.Value)
				continue
			}

			resultMap[domain][rec.Name] = append(resultMap[domain][rec.Name], rec)
		}
	}

	return resultMap, nil

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
