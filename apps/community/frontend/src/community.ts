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
}