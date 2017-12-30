package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"github.com/aprice/observatory/utils"
)

// Peers type handles peer management.
type Peers struct {
	knownPeers         PeerSet
	alivePeers         PeerSet
	httpClient         *http.Client
	PeerCheckDuration  time.Duration
	PeerUpdateDuration time.Duration
	myID               uuid.UUID
	myEndpoint         string
	add                chan PeerSet
	live               chan PeerSet
	outKnown           chan PeerSet
	outLive            chan PeerSet
}

// NewPeers constructor.
func NewPeers() *Peers {
	return &Peers{
		knownPeers: make(PeerSet),
		alivePeers: make(PeerSet),
		httpClient: new(http.Client),
		add:        make(chan PeerSet),
		live:       make(chan PeerSet),
		outKnown:   make(chan PeerSet),
		outLive:    make(chan PeerSet),
	}
}

// Run the peer management routine and return a channel for signalling the
// routine to stop. Once stopped, all calls to AlivePeerSet/KnownPeerSet/Add
// will block indefinitely.
func (p *Peers) Run(conf Configuration) utils.SentinelChannel {
	p.PeerCheckDuration = conf.PeerCheckDuration()
	p.PeerUpdateDuration = conf.PeerUpdateDuration()
	p.myID = conf.ID
	p.myEndpoint = conf.Endpoint()

	q := make(utils.SentinelChannel)
	go func() {
		checkTicker := time.NewTicker(p.PeerCheckDuration)
		updateTicker := time.NewTicker(p.PeerUpdateDuration)
		for {
			select {
			case <-q:
				return
			case <-checkTicker.C:
				go p.checkPeers()
			case <-updateTicker.C:
				p.knownPeers = p.alivePeers.Copy()
				go p.updatePeers()
			case peers := <-p.add:
				for id, url := range peers {
					if id != p.myID {
						p.knownPeers[id] = url
					}
				}
			case peers := <-p.live:
				p.alivePeers = peers
			case p.outKnown <- p.knownPeers.Copy():
			case p.outLive <- p.alivePeers.Copy():
			}
		}
	}()
	return q
}

// AlivePeerSet returns the set of known live peers as of last check.
func (p *Peers) AlivePeerSet() PeerSet {
	return <-p.outLive
}

// KnownPeerSet returns the set of known peers (regardless of status) as of last update.
func (p *Peers) KnownPeerSet() PeerSet {
	return <-p.outKnown
}

// AddPeer to the known peers list. It will not be added to alive peers until
// the next peer check, and its known peers won't be added until the next
// peer update.
func (p *Peers) AddPeer(peer string) {
	go func() {
		peers, err := p.updatePeer(peer)
		if err != nil {
			log.Printf("Failed to add peer %s: %v", peer, err)
			return
		}
		p.add <- peers
	}()
}

func (p *Peers) checkPeers() {
	nowAlivePeers := PeerSet{}
	nowKnownPeers := p.KnownPeerSet()

	for id, peer := range nowKnownPeers {
		url := fmt.Sprintf("http://%s/up", peer)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		resp, err := p.httpClient.Do(req)
		if resp != nil {
			defer func() {
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
			}()
		}
		if err != nil || resp.StatusCode > 399 {
			continue
		} else {
			nowAlivePeers[id] = peer
		}
	}

	p.live <- nowAlivePeers
}

func (p *Peers) updatePeers() {
	nowAlivePeers := p.AlivePeerSet()
	nowKnownPeers := make(PeerSet)

	for _, peer := range nowAlivePeers {
		payload, err := p.updatePeer(peer)
		if err != nil {
			continue
		}
		for pid, purl := range payload {
			if pid != p.myID {
				nowKnownPeers[pid] = purl
			}
		}
	}

	p.add <- nowKnownPeers
}

func (p *Peers) updatePeer(peer string) (PeerSet, error) {
	url := fmt.Sprintf("http://%s/peers?iam=%s&endpoint=%s", peer, p.myID.String(), p.myEndpoint)
	log.Printf("Updating from %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.httpClient.Do(req)
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Peer update failed, status %d", resp.StatusCode)
	}

	payload := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	ps := make(PeerSet, len(payload))
	for id, url := range payload {
		uid, err := uuid.FromString(id)
		if err != nil {
			log.Printf("Could not decode %s: %s", id, err.Error())
			continue
		}
		ps[uid] = url
	}
	return ps, nil
}

// PeerSet manages a set of peer Coordinators, mapping unique IDs to endpoints.
type PeerSet map[uuid.UUID]string

// Contains returns true if the PeerSet contains the given coordinator ID.
func (ps PeerSet) Contains(id uuid.UUID) bool {
	_, ok := ps[id]
	return ok
}

// Copy creates a deep copy of this PeerSet.
func (ps PeerSet) Copy() PeerSet {
	copy := make(PeerSet, len(ps))
	for k, v := range ps {
		copy[k] = v
	}
	return copy
}

// EndpointArray returns an array of the endpoints in this PeerSet.
func (ps PeerSet) EndpointArray() []string {
	arr := make([]string, len(ps))
	i := 0
	for _, v := range ps {
		arr[i] = v
		i++
	}
	return arr
}

// IsLeader returns true if the given node is leader (most senior by ID).
func (ps PeerSet) IsLeader(id uuid.UUID) bool {
	for peerID := range ps {
		for i, v := range peerID {
			// First difference we hit and our byte is lower - they're junior
			if id[i] < v {
				break
			}
			// First difference we hit and our byte is higher - we're not leader
			if id[i] > v {
				return false
			}
			// Bytes match - continue to next byte
		}
	}
	return true
}
