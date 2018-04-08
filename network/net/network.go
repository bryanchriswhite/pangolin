package net

import (
	"../utils"
	"time"
	"fmt"
	"math/rand"
	"github.com/boltdb/bolt"
)

const NetworkSize = 3

type Network [NetworkSize]*NetNode

func (n *Network) rNode(notNode *NetNode) *NetNode {
	r1 := rand.Intn(NetworkSize)
	node := n[r1]

	if notNode != nil && node.Id == notNode.Id {
		return n.rNode(notNode)
	}

	return node
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
	// }fmt.Sprintf("%v",
}

func (n *Network) RandomGossip(t time.Time) (err error) {
	node := n.rNode(nil)
	peer := n.rNode(node)

	h, m, s := t.Clock()
	newData := &Diff{
		state1: node.State,
		Data: map[utils.Any]utils.Any{
			t.Unix(): fmt.Sprintf("%d:%d:%d", h, m, s),
		},
	}

	err = node.State.write(newData)
	if err != nil {
		return
	}

	// d1, _ := node.State.Diff(&peer.State)
	// fmt.Println("Node-peer diff1:", d1.Data)
	fmt.Printf("\nnode %d; contacting node: %d\n", node.Id, peer.Id)
	node.sync(peer)
	// nd1 := node.State.Diff(&peer.State)
	// nd2 := peer.State.Diff(&node.State)
	// fmt.Println("Node-peer diff1:", nd1.Data)
	// fmt.Println("Node-peer diff2:", nd2.Data)
	for _, p := range node.Peers {
		if p.Id != peer.Id {
			continue
		}

		// d1 := node.State.Diff(&p.State)
		d1 := peer.State.Diff(&p.State)
		// d2 := p.State.Diff(&node.State)
		d2 := p.State.Diff(&peer.State)
		fmt.Println("Peer-_peer diff1:", d1.Data)
		fmt.Println("Peer-_peer diff2:", d2.Data)
	}

	return err
}

func (n *NetNode) sync(peer *NetNode) {
	selfDiff := peer.State.Diff(&n.State)
	peerDiff := n.State.Diff(&peer.State)
	fmt.Println("peerDiff:", peerDiff.Data)
	// Update node's state
	if selfDiff.isEmpty() == false {
		n.State.write(&selfDiff)
	}

	// Find node's copy of `peer`
	var _peer *NetNode
	for _, p := range n.Peers {
		// fmt.Printf("p.Id: %v\npeer.Id: %v\n\n", p.Id, peer.Id)
		if p.Id == peer.Id {
			_peer = &p
			break
		}
	}
	// fmt.Println("_peer:", _peer)

	// Update node's copy of peer's state
	_peer.State.write(&peerDiff)

	// Update peer's state? (use channel instaed?)
	peer.State.write(&peerDiff)
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
		// fmt.Println("bucketName:", bucketName)
		_, err = tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	return err
}
