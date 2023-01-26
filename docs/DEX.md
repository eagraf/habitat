# DEX

## Name
When I first tried to write this doc, I kept on saying stuff like "the data system" or "the data API", which was both vague and verbose. To keep things as specific as possible, I started to use the codename `DEX` for the Habitat data layer. It has two possible meanings:
* The data inDEX: `DEX` is like an index or directory for all the data available on the Habitat network. It provides an API that lets you find exactly what data is available, and how to get it.
* Data EXchange: `DEX` supports interoperable data exchange across communities, applications, and users.

This name is subject to change in the future. But for now, we need to call it something.

## What is DEX?
DEX is an entrypoint for programs and humans to access all the data on Habitat. 

DEX is necessary to resolve the conflict between two of Habitat's conflicting objectives:
* Maximum interoperability between applications
* Support for 3rd party components integrated into all parts of Habitat (specifically datastores in this case)

The wide range of (future) available datastores makes interoperability hard. DEX provides a generic method for accessing data spread across different datastores and communities through a single API. In addition, through `DEX Data Sources`, 

## Universal Data Address Space
In DEX, data is organized in a universal address space that accounts for different communities, datastores, and namespaces. Data addresses have the following format:

`/dex/<community_id>/<datastore_id>/<namespace>/<content_identifier>`

### Communities
The highest division of data in Habitat is by community ID. TODO link to the community doc when it exists

### Datastores
Habitat supports many different datastores. Each type of datastore has a unique identifier, which is usually the conventional name of the datastore plus a version specifier. For example, Postgres 13 might have a datastore identifier of `postgres13`.

### Namespaces
Namespaces allows for data to be grouped across datastores. This is useful for permissions and authorization, as well as simply giving a high level grouping for data that can be utilized by applications. The namespace has analogues in the different types of datastores. In a RDBMS, a namespace corresponds roughly to a database, while in a filesystem a namespace might correspond to a disk partition or drive. The most common use of namespaces is to group data written by the same application to different datastrores. Every datasource must partition data by namespace.

Different applications can write to the same namespace, but they must agree on what the namespace is for `TODO figure out how to prevent conflicts`. 

`TODO lets discuss whether this concept is really needed. Having all data partitioned by application might be fine, as long as cross application reads are possible via sources`

### Content Identifier
The format of the content identifier varies from datastore to datastore. Examples could include file paths, content hashes, or SQL queries.

## DEX Sources
Given the ability for many different applications to write data across many different types of datastores, things will quickly grow to complex for an ecosystem of interoperable applications to thrive.

To support data interoperability, `DEX Sources` allow communities to specify common APIs that they implement. This system is split into three components: 
* Schemas
* Interfaces
* Implementations

The purpose of this is to make it easy to tell what APIs a community actually publishes. In a sentence, you might say `"community X implements interface Y which returns schema Z"`. 

`Schemas` are structured data types that are registered with the `data_system_name`. They are content-addressed, so changes to the structure will result in a reference to a fundamentally different schema. They are written in the format of [JSON Schemas](https://json-schema.org/).

`Interfaces` pair a schema with a description of an API being provided. There can be many `Interfaces` with the same `Schema`, as long as their description text differs. The reason for this is that many APIs can serve the same data type while having completely different semantics. For example, if there was a `photo` schema, two very different interfaces for it might be `profile_picture` and `receipt_picture`. The underlying type is the same, but what the API is specifying is completely different.

Each community can maintain a list of `Implementations` for `Interfaces` that it cares about. A `Implementation` is just a path to a specific piece of data following the format `/<community_id>/<datastore>/<namespace>/<content_identifer>` (just the back half of the URL to access any given data in the first place). Note that a community can provide multiple `Implementations` for a single `Interface`. This is to allow different datastores, and different content identifiers within the same store to be able to implement an `Interface`. It is up to applications to choose which `Implementation` to use if multiple are presented. Since not all datastores present their data in a format that converts cleanly to JSON, each datastore must provide a facility to "JSONify" data retrieved via a content identifier. More advanced custom transformations can be defined in the future as well.

Since `Schemas` and `Interfaces` are content-addressed, they are stored by the node, and not by specific communities. `Implementations` are community specific, because they reference paths to data held under specific communities.


## The DEX Node API

The DEX Node API is a read only API accessible over HTTPS through the reverse proxy at addresses following this pattern:

```<node_host>:<reverse_proxy_port>/dex```

### `GET /dex/<community_id>/<datastore_id>/<namespace>/<content_identifier>`


### `GET /dex/schema/<schema_hash>`
The schema hash is a hash of the schema data. Schemas are structued as JSON Schemas. 

### `GET /dex/interface/<interface_hash>`
```
{
    "schema_hash": "age4y43ysgsggh",
    "description": "This is an interface for blog posts"
}
```

### `GET /dex/<community_id>/implementation/<interface_hash>`
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

## The DEX proxy
There are two layers of request forwarding in DEX. The first layer is handled explicitly by DEX, while the second is forwarding handled internally by datastores. The [network topology](src/TOPOLOGY.md) of the community can be used to determine proxying rules.

### DEX Proxy Forwarding
If the current node fielding a request does not host an instance of the datastore that is being requested, it can forward the request to the proper node.

### Datastore Forwarding
Internally, datastore instances that are part of the same process can forward requests between themselves, through a protocol of their own choosing. This will support DHTs, read replicas, and other forms of distributed data.


## The DEX datastore driver API

The driver API is very similare to the Node API. In fact, the for the `schema` and `interface` endpoints, DEX forwards the request to each datastore to see if one has the result. (Eh.. not sure about this, we might want to default to one, but everything should be able to implement it at least).

### `GET /dex/driver/schema`
### `GET /dex/driver/interface/<schema_hash>`
### `GET /dex/driver`
### `GET /dex/<namespace>/<content_identifier>`


### JSON Serialization
Datastores store data in very different formats. To enforce some uniformity across datastores, the Driver API requires that each datastore is able to serialize the data returned by any query into JSON. 

## Permissions 
`TODO`