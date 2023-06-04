package main

import (
	"flag"
	"log"
	"net"
	"time"

	natpmp "github.com/jackpal/go-nat-pmp"
)

const (
	portMappingRequestBuffer = 15 * time.Second
)

var (
	gatewayIPStr string
)

func init() {
	flag.StringVar(&gatewayIPStr, "gateway", "", "gateway ip address")
	flag.Parse()
}

func main() {
	gatewayIP := net.ParseIP(gatewayIPStr)
	if gatewayIP == nil {
		log.Fatalf("failed to parse gateway ip address: %s", gatewayIPStr)
	}

	natpmpClient := natpmp.NewClient(gatewayIP)

	previousExternalPort := uint16(0)

	for {
		portMapping, err := natpmpClient.AddPortMapping("tcp", 0, int(previousExternalPort), 600)
		if err != nil {
			log.Fatalf("failed to add port mapping: %v", err)
		}

		nextPortMappingTime := time.Now().Add(time.Duration(portMapping.PortMappingLifetimeInSeconds)*time.Second - portMappingRequestBuffer)

		if portMapping.MappedExternalPort != previousExternalPort {
			log.Printf("external port changed to %d", portMapping.MappedExternalPort)

			previousExternalPort = portMapping.MappedExternalPort

			// TODO: update Transmission
		}

		time.Sleep(time.Until(nextPortMappingTime))
	}
}
