// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	. "github.com/thethingsnetwork/core/core"
	"github.com/thethingsnetwork/core/core/adapters/http/broadcast"
	"github.com/thethingsnetwork/core/core/adapters/semtech"
	"github.com/thethingsnetwork/core/core/components"
	"github.com/thethingsnetwork/core/utils/log"
	"strconv"
	"strings"
)

func main() {
	// Parse options
	brokers, udpPort := parseOptions()

	// Instantiate all components
	gtwAdapter, err := semtech.NewAdapter(uint(udpPort), log.DebugLogger{Tag: "Gateway Adapter"})
	if err != nil {
		panic(err)
	}

	brkAdapter, err := broadcast.NewAdapter(brokers, log.DebugLogger{Tag: "Broker Adapter"})
	if err != nil {
		panic(err)
	}

	router, err := components.NewRouter(log.DebugLogger{Tag: "Router"})
	if err != nil {
		panic(err)
	}

	// Bring the service to life
	go func() {
		for {
			packet, an, err := gtwAdapter.Next()
			if err != nil {
				fmt.Println(err)
				continue
			}
			go func(packet Packet, an AckNacker) {
				if err := router.HandleUp(packet, an, brkAdapter); err != nil {
					fmt.Println(err)
				}
			}(packet, an)
		}
	}()
}

func parseOptions() (brokers []Recipient, udpPort uint64) {
	var brokersFlag string
	var udpPortFlag string
	flag.StringVar(&brokersFlag, "brokers", "", `Broker addresses to which broadcast packets.
 	For instance: 10.10.3.34:8080,thethingsnetwork.broker.com:3000
 	`)
	flag.StringVar(&udpPortFlag, "udpPort", "", "Udp port on which the router should listen to.")
	flag.Parse()

	var err error
	udpPort, err = strconv.ParseUint(udpPortFlag, 10, 64)
	if err != nil {
		panic(err)
	}
	if brokersFlag == "" {
		panic("Need to provide at least one broker address")
	}

	brokersStr := strings.Split(brokersFlag, ",")
	for i := range brokersStr {
		brokers = append(brokers, Recipient{
			Address: strings.Trim(brokersStr[i], " "),
			Id:      i,
		})

	}
	return
}
