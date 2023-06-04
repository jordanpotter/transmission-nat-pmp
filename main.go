package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/hekmon/transmissionrpc/v2"
	natpmp "github.com/jackpal/go-nat-pmp"
)

const (
	minPortMappingPeriod     = 10 * time.Second
	portMappingLatencyBuffer = 15 * time.Second
)

var (
	gatewayIP            = flag.String("gateway", "", "gateway ip address")
	transmissionHostname = flag.String("hostname", "127.0.0.1", "transmission hostname")
	transmissionPort     = flag.Int("port", 9091, "transmission port")
	transmissionUsername = os.Getenv("TRANSMISSION_USERNAME")
	transmissionPassword = os.Getenv("TRANSMISSION_PASSWORD")
)

func init() {
	flag.Parse()
}

func main() {
	natpmpClient := natpmp.NewClient(net.ParseIP(*gatewayIP))

	transmissionClient, err := transmissionrpc.New(*transmissionHostname, transmissionUsername, transmissionPassword, &transmissionrpc.AdvancedConfig{
		Port: uint16(*transmissionPort),
	})
	if err != nil {
		log.Fatalf("failed to create transmission client: %v", err)
	}

	previousExternalPort := uint16(0)

	for {
		portMappingTime := time.Now()

		portMapping, err := natpmpClient.AddPortMapping("tcp", 0, int(previousExternalPort), int(time.Hour.Seconds()))
		if err != nil {
			log.Fatalf("failed to add port mapping: %v", err)
		}

		if portMapping.MappedExternalPort != previousExternalPort {
			log.Printf("external port changed to: %d", portMapping.MappedExternalPort)

			previousExternalPort = portMapping.MappedExternalPort

			transmissionPeerPort := int64(portMapping.MappedExternalPort)
			err = transmissionClient.SessionArgumentsSet(context.Background(), transmissionrpc.SessionArguments{
				PeerPort: &transmissionPeerPort,
			})
			if err != nil {
				log.Fatalf("failed to set transmission peer port: %v", err)
			}

			log.Printf("updated transmission peer port to: %d", transmissionPeerPort)
		}

		nextPortMappingWait := time.Duration(portMapping.PortMappingLifetimeInSeconds)*time.Second - portMappingLatencyBuffer
		if nextPortMappingWait < minPortMappingPeriod {
			nextPortMappingWait = minPortMappingPeriod
		}

		nextPortMappingTime := portMappingTime.Add(nextPortMappingWait)

		time.Sleep(time.Until(nextPortMappingTime))
	}
}
