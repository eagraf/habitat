# RPC

Habitat RPC will allow for processes to communicate with each other without using publicly exposed channels that are exposed by the proxy server.

## Future

* We should decide whether we want to stick to a specific serialization format or leave that up to the implementer. If its left up to the implementer, then all RPC will be passed as simple byte arrays and serialization/deserialization will be handled in the calling/receiving code.
* We should figure out how to handle routing. We could either have a flat system where each message handler is top level, or we could implement sub routes.
* We need to figure out authentication for these calls
* We should think about having Raft consensus use this RPC library as part of the transport library