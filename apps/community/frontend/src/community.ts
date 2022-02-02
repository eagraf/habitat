export type CreateCommunityResponse = {
    name: string
    swarm_key: string
    btstp_peers: string[]
    peers: string[]
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