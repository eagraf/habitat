# Architecture

## Layers
Habitat can be split into roughly 4 layers:

1. P2P Transport
2. Node Services
3. Control
4. Data Sources
5. Applications

The majority of Habitat's core code is within layers two and three. Layer one relies heavilly on `libp2p`, and layer four is intended to be composed mainly of third party applications.

### P2P Transport
The peer-to-peer transport layer is implemented using `libp2p`, which provides critical features such as NAT holepunching, node discovery, relaying, address determination, encrypted message transport, and more. 

The most important feature provided by `libp2p` is baked in NAT holepunching support. While this is very well supported by the library, its important to understand the limitations that are naturally imposed by NATs, and the resulting contraints placed on `libp2p`. The following are some of the major challenges that have to be considered:

* Hole-punching requires a relay node to perform the synchronization step of the DCUTR protocol. Both end nodes need to be able to connect to this temporary relay, so a dialing node must know beforehand what relays a destination node is attached to. Habitat identifies relays through community configuration, but this implies that each community must have at least one relay. In the future, it could be advantageous to optionally use external nodes as relays to perform synchronization.
* `libp2p` automatically identifies potential relay nodes by having each node ask its peers to attempt to dial it across the public internet. If these peers are able to reach the node without a relay, and if the node presents a consistent IP address (its not behind a symmetric NAT), then it can be marked as reachable and become a relay. Importantly, `libp2p` requires that three distinct peers make a successful connection before a node can be marked public. This might make things challenging for small communities. Even if they have a publicly accessible node, if they don't have enough peers for `libp2p` to confirm the node's public status, it will be unable to act as a relay, and therefore NAT holepunching capabilities will not be available to the community. 

For more on how `libp2p manages hole-punching, check out this [paper](https://research.protocol.ai/publications/decentralized-hole-punching/seemann2022.pdf).

### Node Services
The node layer comprises all facilities in Habitat that are shared between communities. These include:

* Reverse proxy
* Distributed file system
* Process manager
* Data sources

#### Reverse Proxy
An important part of the design philosophy of Habitat is to open as few ports as necessary. Since a Habitat instance could potentially run dozens of applications at at time, a reverse proxy is used to map requests to the proper application based on the path expressed in the URL.

The reverse proxy is served on port 2041. In addition, the reverse proxy is available via `libp2p` transport on port 2042. The `libp2p` reverse proxy just redirects to the first.

There are two main types of reverse proxy rules. Redirect rules redirect traffic to another server running on the nodee. In the future, this should also be able to redirect traffic to another node as well. The second is file server rules, which allow for a directory of files to be served, most likely for web content. The correct rule for a given request is determined based off of the path in the request URL. Right now, the matching is just done by prefix, but in the future a glob based matcher will be implemented that can also detect rule conflicts.

In the future, the reverse proxy might integrate more tightly with the data proxying facilities in Habitat.

#### Distributed File System
Habitat will eventually come packaged with a distributed file system, resembling IPFS. The most important featured of this system is content addressing, which creates a number of benefits for running the filesystem in a distributed environment.

Until the distributed file system can be implemented, an instance of IPFS is run alongside Habitat. 

#### Process Manager
The process manager allows for applications to be started and stopped on a Habitat node. When a process is started, a child process is created and added to the same process group as the Habitat server. If the Habitat server is killed, all processes will also be destroyed.

In the future, we should consider whether this functionality should mostly be handled by systemd (and equivalent on other OSes).

#### Data Sources
To better facilitate data sharing across nodes, Habitat communities can publish "data sources" that conform to a specific schema specified in a schema registry. Depending on permissions, these data sources can be publicly available, allowing for the creation of aggregating applications, which can range from social networks to feed readers. 

Since the ultimate source of data might not be contained by a node fielding a request, the `data proxy` can direct requests to the node that is able to serve the correct data. The data proxy layer can also handler authentication and permissions.

### Control

There are two systems that control Habitat nodes. The first is the `habitatctl` client, which directly accesses the `Habitat API` to perform actions. The second, and more important, is the community management service.

#### Habitat Client
The `habitatctl` CLI can be used to control Habitat remotely or locally. It provides commands that interface directly with the Habitat API. It gives the option for custom addresses, as well as the ability to hit the API via `libp2p` to circumvent NATs.

In the future, a more interactive CLI can be created using the Bubble Tea library. In addition, `habitatctl` can run a local server for a client web app to make interacting with Habitat easier for non-technical users.

#### Community Management
A community is a cluster of nodes that collectively maintain a shared filesystem, shared applications, and more. These nodes can be spread out geographically and can number from one to thousands. Communities have a consistent state that is maintained by a subset of the community nodes executing the [Raft](https://raft.github.io/) consensus algorithm. The community control layer is executed by the `CommunityManager` struct.

##### Consensus
All nodes need to agree on the community state for Habitat to function. To reach agreement on the state, nodes in a community run the Raft algorithm to create an ordered log of state updates. This effectively turns the community into a replicated state machine. Since Raft does not scale well beyond ~10ish nodes (although I don't know if this has been tested at large scales), larger communities will only have some nodes elected as Raft participants. When a new state is agreed upon, the new updates will be broadcast to all other nodes. A algorithm needs to be created to determine what nodes are Raft nodes.



##### State Management
Community state is represented as a JSON object conforming to a JSON schema. There is a limited set of allowed state transitions defined in `transitions.go` in the `state` package. Each transition is defined by a JSON patch that specifies the changes to the community state, as well as a validator function that checks the transition and the resulting community state for errors. Instances of the `Executor` interface receive these state updates via a channel, and can act on them. Each node service should implement this interface to be able to react to these updates.

When a node restarts, all of the layer two node services are reset to the last known state for each community. Then, it asks the community for state updates it missed, and each is applied individually. After that, the node will receive state updates normally.

### Applications
Applications are the top level of Habitat. Applications can contain server and web content portions. They are started by the process manager based off of community state.

In the future, a SDK will be provided to make developing Habitat applications easier. In addition, a package manager will be provided that can be used to install and update Habitat apps.
