
# The Data Layer

Users controlling their own data is a core feature of Habitat. To achieve this, each user's personal node must be capable of serving data in a variety of different formats. Its conceivable that user's will be running applications that simultaneously require relational databases, key-value stores, distributed filesystem access, and more. Habitat refers to these different types of data systems as `datastores`. While these different datastores typically have very different access patterns, Habitat attempts to present a unified data layer that will allow applications to access the datastores through a single API (we need a nice name for this).

```
Note: This system is probably important enough that we give it its own name, that is not just purely descriptive. 
```

Why have a single interface for so many different types of datastores? After all, expected data both in and out varies greatly between the different types of datastores. However, there are several major advantages of a single data access layer across all of Habitat:
1. Interoperable applications are way easier to build if they all use the same API for accessing data.
2. Consolidating the access point for these datastores reduces the number of ports that have to be opened on each node, simplifying the system.
3. Common proxying logic can be implemented to forward requests for data not immediately available on the queried node's machine. 
4. A common authorization and permissions system can be implemented as middleware for this data access API. This also increases the interoperability of applications.

## The API

The data layer is served over `http`. Using the Habitat proxy, HTTP requests can be made over a `LibP2P` transport layer to bypass NATs. All data in the system can be accessed by a unique identifier following this pattern:

unique_identifier: `/<community_id>/<datastore>/<namespace>/<content_identifier>`

This data can be accessed through a node's `data` API, following this format:

URL: `https://<node_host>:<reverse_proxy_port>/data/<unique_identifier>`

The components:
* `node_host`: IP or domain name for the node being queried
* `reverse_proxy_port`: The node's reverse proxy port, default 2041
* `community_id`: The id for the community storing the data you are trying to access
* `datastore`: The datastore that the request should be routed towards
* `namespace`: The namespace that the data is under.
* `content_identifier`: String the datastore uses to find the data you are querying.

Any HTTP method can be used, as long as the datastore the request is routed to supports it. The `content_identifier` format will differ between datastores, and can be anything from a file path to a content hash, to a SQL query.

Generally, applications will rotue requests through the node serving the application to simplify client logic. Via proxying, the request should be completed even if the data is not stored directly on the node. This is expanded on in the `Proxying` section of this file. Data is not coupled to any given node, and should generally be replicated across several nodes whenever possible.

## Namespaces
Namespaces allows for data to be grouped across datastores. This is useful for permissions and authorization, as well as simply giving a high level grouping for data that can be utilized by applications. The namespace has analogues in the different types of datastores. In a RDBMS, a namespace corresponds roughly to a database, while in a filesystem a namespace might correspond to a disk partition or drive. The most common use of namespaces is to group data written by the same application to different datastrores. Every datasource must partition data by namespace.

Different applications can write to the same namespace, but they must agree on what the namespace is for `TODO figure out how to prevent conflicts`.

## Schemas and Sources
Given the ability for many different applications to write data across many different types of datastores, things will quickly grow to complex for an ecosystem of interoperable applications to thrive.

To support data interoperability, Habitat provides an API for specifying common APIs that communities can implement. This system is split into three components: 
* Schemas
* Interfaces
* Implementations

The purpose of this is to make it easy to tell what APIs a community actually publishes. In a sentence, you might say `"community X implements interface Y which returns schema Z"`. 

`Schemas` are structured data types that are registered with the `data_system_name`. They are content-addressed, so changes to the structure will result in a reference to a fundamentally different schema. They are written in the format of [JSON Schemas](https://json-schema.org/).

`Interfaces` pair a schema with a description of an API being provided. There can be many `Interfaces` with the same `Schema`, as long as their description text differs. The reason for this is that many APIs can serve the same data type while having completely different semantics. For example, if there was a `photo` schema, two very different interfaces for it might be `profile_picture` and `receipt_picture`. The underlying type is the same, but what the API is specifying is completely different.

Each community can maintain a list of `Implementations` for `Interfaces` that it cares about. A `Implementation` is just a path to a specific piece of data following the format `/<community_id>/<datastore>/<namespace>/<content_identifer>` (just the back half of the URL to access any given data in the first place). Note that a community can provide multiple `Implementations` for a single `Interface`. This is to allow different datastores, and different content identifiers within the same store to be able to implement an `Interface`. It is up to applications to choose which `Implementation` to use if multiple are presented. Since not all datastores present their data in a format that converts cleanly to JSON, each datastore must provide a facility to "JSONify" data retrieved via a content identifier. More advanced custom transformations can be defined in the future as well.

Since `Schemas` and `Interfaces` are content-addressed, they are stored by the node, and not by specific communities. `Implementations` are community specific, because they reference paths to data held under specific communities.

``` Note: also not very convinced by these names.```

### Schemas
`/data/schema/<schema_hash>`

The schema hash is a hash of the schema data. Schemas are structued as JSON Schemas. 

### Interfaces
`/data/interface/<interface_hash>`
```
{
    "schema_hash": "age4y43ysgsggh",
    "description": "This is an interface for blog posts"
}
```

### Implementations
`/data/implementations/<interface_hash>`
```
{
    "interface_hash": "ageg32543ghdfhdfh",
    "implementations": {
        "datastore_1": "<content_identifier>",
        "datastore_2": "<content_identifier>",
        "datastore_3": "<content_identifier>",
    }
}
```

### Migrating Schemas and Sources
TODO

## Proxying
TODO
