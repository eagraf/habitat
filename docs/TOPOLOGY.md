# Process Topology
```
Note: this document describes the connections between applications, drivers, datastores, and nodes. It doesn't go into the details of the p2p connections between nodes.
```

One of the core functions of each community's state machine is to keep track of all the processes running on that community. There are several main types of processes managed by Habitat's process manager:
* Applications
* Datastore drivers
* Datastores

In addition, the topology must track which nodes these different processes are running on.

To keep the network topology of interconnected processes in Habitat easy to understand, a set of simple rules are enforced by all Habitat nodes. These rules define which processes are allowed to comunicate with which other ones.

## Rules
The following is the basic set of rules that apply to the different types of vertexes in the topology:
* Nodes
* Processes(applications, datastores, drivers)
* Process Instances (...)

### Process rules:
* Processes cannot be connected to nodes
* Each process must have at least one instance of the same type

### Process Instance Rules:
* All process instances must be attached to a node

In addition, there are type specific rules for process instances:

#### Application Instance rules:
* Applications instances cannot directly communicate with other applications. They must instead use the Habitat reverse proxy
* Application instances can communicate directly with datastores, both locally and on remote nodes.
* Application instances cannot directly communicate to datastore drivers, because they just implement the driver API, which applications can reach throught the proxy.

#### Datastore Driver Rules:
* Datastore drivers can only have direct communication with datastores of the corresponding type.

#### Datastore Instance Rules:
* Datastores instances can have communication with one driver process of the corresponding type.
* Datastores instances can have direct communication with other datastores of the same type on the same community. This allows for distributed datastores like DHTs, and more.

## Enforcement
The Habitat state machine enforces these rules through validaton on state updates. No state updates can be made that violate these conditions. 

Several subsystems in Habitat are also informed by these rules, including port management and firewall management.