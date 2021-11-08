export type CreateCommunityResponse = {
    name: string
    swarm_key: string
    btstp_peers: string[]
    peers: string[]
  }

export type JoinCommunityResponse = CreateCommunityResponse