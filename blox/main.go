package main

import (
	"github.com/bloxapp/prysm/blox/client"
)

func main() {
	params := client.New(
		"100.65.90.215:4000",
		"remote",
		"{\"location\":\"localhost:8088\",\"accounts\":[\"NValidator/Account.*\"],\"certificates\":{\"ca_cert\":\"/Users/nick/GoLandProjects/bloxapp_prysm/credentials/localhost/ca.crt\",\"client_cert\":\"/Users/nick/GoLandProjects/bloxapp_prysm/credentials/localhost/clients/1/client.crt\",\"client_key\":\"/Users/nick/GoLandProjects/bloxapp_prysm/credentials/localhost/clients/1/client.key\"}}",
	)
	err := client.Run(params)
	if err != nil {
		println(err)
	}
}
