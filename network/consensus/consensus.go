package consensus

import (
	"../net"
	"../utils"
	"time"
	"os"
	"log"
)

type Options struct {
	Interval time.Duration
	Path     string
	Mode     os.FileMode
}

func Run(p *utils.Program, network net.Network, o Options) {
		go (func(n *net.Network) {
			// Loop forever every `interval`
			ticker := time.NewTicker(o.Interval)
			for t := range ticker.C {
				err := n.RandomGossip(t)
				if err != nil {
					log.Fatal(err)
					p.Exit(2)
				}
			}
		})(&network)
}
