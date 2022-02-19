package community

type Community struct {
	Name      string   `json:"name"`
	Id        string   `json:"id"`
	PeerId    string   `json:"peer_id"`
	SwarmKey  string   `json:"swarm_key"`
	Addresses []string `json:"addrs"`
	Peers     []string `json:"peers"`
}
