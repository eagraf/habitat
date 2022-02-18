export type CreateCommunityResponse = {
    name: string
    id: string
    peer_id: string
    swarm_key: string
    peers: string[]
    addrs: string[]
  }

export type JoinCommunityResponse = CreateCommunityResponse

export type ConnectCommunityResponse = {
  Addresses: string[]
  AgentVersion: string
  ID: string
  ProtocolVersion: string
  Protocols: string[]
  PublicKey: string
  SwarmKey: string
}

export type ListCommunitiesResponse = {
  Communities: string[]
}

// map from community name / "id" to url
export type ConnectedCommunities = Map<string, string[]>