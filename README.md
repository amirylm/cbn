# CBN

POC of a decentralized libp2p based network for storage and delivery of Content Buckets.

## Mini-Design

Content buckets are stored in a way that ensures immutability, integrity and authenticity.
They contain references to data that is located in some decentralized storage.
The reference decouples the bucket from data and therefore allows to store data anywhere 
or define things like data protection or authorization.

The network consists of several types of peers:
* `Bootnode` is the way to connect to the network with some other libp2p peer
* `HTTP Gateway` allows to interact with the network through HTTP
* `Storage Peer` or just `Node` - stores buckets or data

![](/assets/diagram.png)

Data is referenced by buckets and saved across several storages, regardless of where the bucket is stored.

`Bucket Registry` is responsible for storing the buckets headers / metadata, 
while `Bucket Source` holds the actual bucket implementation.
`Data Source` is the actual storage for the data.

Finally, `Domain Registry` should provide decentralized naming. 

Note that Bucket/Data Sources are required to "speak" IPLD (nodes, CIDs..)

As buckets can be signed offline (by the owning identity), 
it is possible to manage buckets w/o having an own storage peer in the network.

![](/assets/diagram-data.png)

### Vision

The vision in this project is to provide a decentralized solution 
that enables an eco-system / marketplace with multiple players and incentives: 
* Provides / Consumers
* Content hosting / pinning
* Distribution across data sources
* Relayers
* etc... 

During the work I came across [textile.io](https://textile.io/) 
which provides managed solutions and tools on top of IPFS and filecoin.
Future work of this POC should probably take [textileio/powergate](https://github.com/textileio/powergate) into account.

### TODO List

#### 1. Merkle CRDT

Need to fork/contribute ipfs/go-ds-crdt to add the missing logic for having ownership of data in the tree

i.e. only the peer that created a key (in a specific namespace, e.g. /{pub-key}/...) is able to modifiy it's value
currently this logic is implemented in this project, which is wrong.
in addition, anyone can mess the merkle crdt as peers are not validating changes (hooks are triggered post storage).

#### 2. Multiple Data/Bucket Sources

Need to add more data/bucket sources (e.g. IPFS or filecoin).
currently the only source is `p2p`, which uses IPLD infra from [amirylm/libp2p-facade](https://github.com/amirylm/libp2p-facade)

[libp2p protocols negotiation](https://docs.libp2p.io/concepts/protocols) can be used to apply multiple sources, 
see [ProtoBook interface](https://github.com/libp2p/go-libp2p-core/blob/a39b84ea2e340466d57fdb342c7d62f12957d972/peerstore/peerstore.go#L239)
which can be used to find supporting peers
 * e.g. find peers that supports `/bucket/ipfs/1.0.0` protocol
 * `/bucket/filecoin/1.0.0` protocol for filecoin peers
 * etc...

#### 3. Data Ref

Data Ref defines how to access data, which decouples data from the underlaying data source, 
and the authorization (e.g. protected data).

Content metadata like type or size is also a good candidate to be included in `DataRef`. 

`ProtectedDataRef` should provide the needed functionality to achieve content authorization, 
it will act as an individual (per identity) access key to data.

A good option is to use JWS (see [alexjg/go-dag-jose](https://github.com/alexjg/go-dag-jose)) 
as a wrapper for this object
   
#### 4. Content Authorization

In order to apply subscriptions, content providers creates a 
[subscription channel](https://github.com/amirylm/subscription-channel)
contract on Ethereum, which is used to receive tokens from subscribed users.

Each subscription channel have a corresponding channel key that is used to ensure content protection. 
   
#### 5. Domain Registry

In order to achieve integration with other platforms, standard DNS records should be used..

Alternatives:
* [IPNS](https://docs.ipfs.io/concepts/ipns/)
* [Ethereum Naming System](https://ens.domains/)


## Usage

Install: 

```bash
git clone git@github.com:amirylm/cbn.git
```

Available flavors can be found at `./cmd/...`

Note that you need to change `.env` to include a bootnode peer

### Docker

Use docker to start local network:

```bash
docker-compose up
```

### Node

Run node with terminal mode:

(change `.env` to have bootnode peer)

```bash
cd cmd/node
go run .
```

### HTTP Gateway

http-gateway is available when running docker-compose ([localhost:3010](http://localhost:3010)) 

```bash
curl http://localhost:3010/buckets

> {"data":[...],"time":1605125320}


curl http://localhost:3010/buckets/{bucket_hash}/{file_name}

> <file content...>
``` 

