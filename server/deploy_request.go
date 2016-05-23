package server

import "encoding/json"

// DeployRequest is a struct used to demarshal requests for a deploy and also to process.
type DeployRequest struct {
	DeployID     string `json:"deployID"`     // A UUID for the request and for this deploy (client filled).
	ImageName    string `json:"imageName"`    // Image name in the repository in docker registry (client filled).
	ImageTag     string `json:"imageTag"`     // Image tag to deploy (client filled).
	Environment  string `json:"environment"`  // Environment from config.yml (client filled).
	EnvTag       string `json:"envTag"`       // Used to resolve machine names that are in the env (machine filled).
	EtcdEndpoint string `json:"etcdEndpoint"` // Etcd hostname and port (machine filled).
	Machine      string `json:"machine"`      // Master machine node for the cluster or local (machine filled).
	MetaMount    string `json:"metaMount"`    // Remote directory on a machine to place the metadata (machine filled).
	NumCont      int    `json:"numCont"`      // Default number of containers for this environment (machine filled).
	Registry     string `json:"registry"`     // Docker registry for this environment (machine filled).
	Swarm        bool   `json:"swarm"`        // Is this machine apart of a cluster (machine filled)?
}

// NewDeployRequest is a factory function that returns a DeployRequest instance.
func NewDeployRequest(deployID string, imageName string, imageTag string, environment string,
	envTag string, etcdEndpoint string, machine string, metaMount string, numCont int,
	registry string, swarm bool) *DeployRequest {
	return &DeployRequest{
		DeployID:     deployID,
		ImageName:    imageName,
		ImageTag:     imageTag,
		Environment:  environment,
		EnvTag:       envTag,
		EtcdEndpoint: etcdEndpoint,
		Machine:      machine,
		MetaMount:    metaMount,
		NumCont:      numCont,
		Registry:     registry,
		Swarm:        swarm,
	}
}

// String is an implentation of the Stringer interface so the structure is returned as a string
// to fmt.Print() etc.
func (r *DeployRequest) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}
