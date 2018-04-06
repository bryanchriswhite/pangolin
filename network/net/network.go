package net

import (
	"../utils"
	"time"
	"fmt"
	"math/rand"
	"github.com/boltdb/bolt"
)

const NetworkSize = 4

type Network [NetworkSize]*NetNode

func (n *Network) rNode() *NetNode {
	r1 := rand.Intn(NetworkSize)
	return n[r1]
}

type NetNode struct {
	Id    int
	Peers []NetNode
	State State
	C     chan *State
}

func (n *NetNode) handlePeerStateTransitions(peer *NetNode) {
	// for state := range peer.C {
	// 	n.sync(peer)
	// }
}

func (n *Network) RandomGossip(t time.Time) {
	node, peer := n.rNode(), n.rNode()

	newData := &Diff{
		state1: node.State,
		data: map[utils.Any]utils.Any{
			fmt.Sprintf("%v", t): fmt.Sprintf("%v", t),
		},
	}

	node.State.write(newData)
	node.sync(peer)
	// fmt.Printf("node %d; contacting node: %d\n", n.Id, peer.Id)
	// peer.sync(n)
}

func (n *NetNode) sync(peer *NetNode) {
	selfDiff, peerDiff := n.State.diff(&peer.State)
	// Update node's state
	if selfDiff.isEmpty() == false {
		n.State.write(&selfDiff)
	}

	// Find node's copy of `peer`
	var _peer *NetNode
	for _, p := range n.Peers {
		if p.Id == peer.Id {
			_peer = &p
		}
	}

	// Update node's copy of peer's state
	_peer.State.write(&peerDiff)

	// Update peer's state? (use channel instaed?)
	// peer.State.write(&peerDiff)
	// n.C <- &n.State
}

func Bootstrap(p *utils.Program, db *bolt.DB) (network Network) {
	network = Network{}

	// Generate nodes
	for i := range network {
		network[i] = &NetNode{Id: i}
	}

	// Copy peers to nodes and initialize all states
	for i, node := range network {
		node.Peers = make([]NetNode, 0)
		network[i].State = State{
			Db:     db,
			Bucket: []byte(fmt.Sprintf("node-%d", i)),
		}

		for j, peer := range network {
			if i != j {
				_peer := *peer
				_peer.State = State{
					Db:     db,
					Bucket: []byte(fmt.Sprintf("node-%d:peer-%d", i, j)),
				}

				network[i].Peers = append(node.Peers, _peer)
			}
		}
	}

	bucketErrors := EnsureBuckets(network)
	for _, err := range bucketErrors {
		p.ErrCheck(err)
	}

	return
}

func EnsureBuckets(network Network) (errors []error) {
	errors = make([]error, 0)

	for _, node := range network {
		err := EnsureBucket(node, fmt.Sprintf("node-%d", node.Id))

		if err != nil {
			errors = append(errors, err)
		}

		for _, p := range node.Peers {
			err = EnsureBucket(&p, fmt.Sprintf("node-%d:peer-%d", node.Id, p.Id))
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func EnsureBucket(node *NetNode, bucketName string) (err error) {
	node.State.Db.Update(func(tx *bolt.Tx) (err error) {
		fmt.Println("bucketName:", bucketName)
		_, err = tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	return err
}
