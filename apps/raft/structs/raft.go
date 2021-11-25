package structs

const (
	RPC_STATUS_OK    = 0
	RPC_STATUS_ERROR = 1

	RegisterRoute             = "/register"
	UnregisterRoute           = "/unregister"
	AddNodeRoute              = "/add"
	ApplyStateTransitionRoute = "/apply"
)

type RegisterCommunityRequest struct {
	CommunityID  string `json:"community_id"`
	NewCommunity bool   `json:"new_community"`
	JoinAddress  string `json:"join_address"`
}

type RegisterCommunityResponse struct {
}

type UnregisterCommunityRequest struct {
	CommunityID string `json:"community_id"`
}

type UnregisterCommunityResponse struct {
}

type AddNodeRequest struct {
	CommunityID string `json:"community_id"`
	ServerID    string `json:"server_id"`
	Address     string `json:"address"`
}

type AddNodeResponse struct {
}

type ApplyStateTransitionRequest struct {
	CommunityID string `json:"community_id"`
	Transition  []byte `json:"transition"`
}

type ApplyStateTransitionResponse struct {
}
