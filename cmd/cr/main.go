package main

import (
	"encoding/json"
	"fmt"
	"os"

	cloudresolver "github.com/squarescale/cloudresolver"
)

func resolve(name string,
	resolver cloudresolver.CloudResolver,
	ch chan<- []cloudresolver.Host,
	config map[string]interface{}) {
	hosts, _ := resolver.Resolve(name, config)
	ch <- hosts
}

func main() {
	var config map[string]interface{}
	err := json.Unmarshal([]byte(`{ "providers" : {"gce": { "zone" : "europe-west1-b" }}}`), &config)
	if err != nil {
		panic(fmt.Sprintf("could not unmarshal config: %#v", err))
	}

	ch := make(chan []cloudresolver.Host)
	for _, resolver := range cloudresolver.Resolvers {
		go resolve(os.Args[1], resolver, ch, config)
	}

	var allHosts [][]cloudresolver.Host

	for _, _ = range cloudresolver.Resolvers {
		hs := <-ch
		allHosts = append(allHosts, hs)
	}

	for _, provider := range allHosts {
		if len(provider) > 0 {
			fmt.Printf("Provider: %s\n", provider[0].Provider)
			for _, host := range provider {
				fmt.Printf("\tId: \t\t%s\n", host.Id)
				fmt.Printf("\tRegion: \t%s\n", host.Region)
				fmt.Printf("\tZone: \t\t%s\n", host.Zone)
				fmt.Printf("\tprivate ipv4: \t%s\n", host.PrivateIpv4)
				fmt.Printf("\tpublic ipv4: \t%s\n", host.PublicIpv4)
				fmt.Printf("\tprivate ipv6: \t%s\n", host.PrivateIpv6)
				fmt.Printf("\tpublic ipv6: \t%s\n", host.PublicIpv6)
				fmt.Printf("\tprivate name: \t%s\n", host.PrivateName)
				fmt.Printf("\tpublic name: \t%s\n", host.PublicName)
				fmt.Printf("\tprivate: \t%s\n", host.Private)
				fmt.Printf("\tpublic: \t%s\n", host.Public)
				fmt.Println("")
			}
		}
	}
}
