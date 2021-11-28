This package implements a basic consensus protocol for Habitat communities, based off of the HashiCorp Raft library. Habitat can run multiple communities on one node, necesitating multiple instances of a Raft agent to run side by side, which influences most of the major design decisions in this package.

## HTTP Transport Implementation
One constraint on the design is that we want each node to expose only one port to the internet for the purpose of sharing essential state with other nodes. To do this, we need to have each community listen for state updates through the Raft protocol on the same port without interfering with each other. To keep communities isolated, we append a path to the server address of the node. Unfortunately, Raft only implements a TCP transport layer which can only handler ip:port formatted server addresses. To get around this, we implement our own transport layer that uses HTTP rather than TCP streams. While this may be slower, it allows us to specify paths that a router can use to determine which community's cluster should be modified.

## Multiplexer
The multiplexer acts as the routing layer that redirects incoming Raft protocol requests to the right instance of the Raft agent. This allows for multiple communities to listen for Raft messages on the same server, which reduces complexity in terms of concurrent servers listening to ports. The multiplexer maintains a map linking community id's to their corresponding http transport instance.

## Finite State Machine
CommunityStateMachine implements the raft.FSM interface, and keeps track of the state of a community. The data is maintained as JSON stored in a byte array, with updates being described as JSON patches (http://jsonpatch.com/).