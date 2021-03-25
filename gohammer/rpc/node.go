package rpc

import (
	log "github.com/sirupsen/logrus"

	"github.com/tubuarge/GoHammer/config"
)

// checkNodes calls isNodeUp function to ensure that every node is running
// before starting the test.
// If there is a failed node then terminates the GoHammer.
func (r *RPCClient) CheckNodes(cfg *config.Config) {
	isOK := true

	profiles := cfg.TestProfiles

	for _, profile := range profiles {
		nodes := profile.Nodes
		for _, node := range nodes {
			isNodeUp, err := r.IsNodeUp(node.URL)
			if err != nil {
				isOK = false
				log.Errorf("%s node is not running: %v", node.Name, err)
				continue
			}
			if !isNodeUp {
				isOK = false
				log.Errorf("%s node is not running.", node.Name)
				continue
			}
			log.Infof("%s node is OK.", node.Name)
		}
	}

	if !isOK {
		log.Fatalf("Make sure every node given in the test-profile file is running.")
	}
}
