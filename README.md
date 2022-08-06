# Message Signer

## Architecture

This project implement a signing service for messages.
A signing request is split into batches with batch size
passed as a parameter. All batches perform message signing in parallel.
A set of keys is used for signing. One key is used per batch. Each 
key is used in round-robin manner.

Signing is done with etherum keys pairs using go-ethereum package. A random salt is
added to each message before Keccak256Hash() computation.

Records before signing:
```
type Record struct {
    Id string
    Msg String
}
```

Records after signing:
```
type Record struct {
    Id string
    Msg String
    Signature string
    Salt String
    PublicKey String
}
```

Signing service exposes the following API

```
GET    /                # liveness         
GET    /sign/:size      # signing request, size if a batch size      
GET    /stats           # show signed and unsigned records         
GET    /batch/:batchId/:batchCount/:key # internal endpoint to launch batch signing
```
Note that APIs are not exposed externally via ingress, which would
require registering a domain name or getting a static IP.
For simplicity, they are available inside the cluster only 
and can be called by connection to any pod. See k8s guide for details.


Both signing service and batch signers are packaged as one
binary. To launch a batch signer, the main service calls a batch endpoint
of the same service. Since the service is fully stateless, the call is routed
to the same pod or another using k8s round-robin scheduling mechanism.
This approach is greatly simplifies implementation and deployment.

Mongodb is used a data store. It is a performant  document db, which has 
k8s helm chart available and can be deployed as is into k8s cluster.

Initially unsigned messages are placed into `records` collection. Batch workers query unsigned records from
this collection and perform signing. As each record is signed , it is inserted into
`signedrecords` collection and removed from `records` collection. This 
ensures that none of the records are lost at any time- if a signing fails,
the original record still remains in `records` collection and can be 
signed later. This ensures "atomic" signing.

To simplify record selection for each batch, a simple sharding 
approach is used. First 8 bytes of record id are used to determine a batch id 
as 
```
batchId := Id[:8] % TotalBatches
```
Such approach allows even distribution of records between batch signers.

## Deployment

Signing service is packaged and deployed and k8s app. This 
creates a portable, vendor-independent implementation which can
be hosted on any provider with k8s support such as GCP, DigitalOcean, Vult and others

Signing keys are pre-generated and packaged as k8s secret object.

Sample record:
```
  {
    _id: ObjectId("62ee2e75d199969df1d5db58"),
    id: '830f559b22b74bfcbb5631fae20462cb',
    key: '0x04e8d44631471324a13c37e029301913c76f6eeb3277d5892a0ef58715b39c3d203d2c60f842442890e8ffaacfa4acc3eb061bab813a6f56ba629691d996a09de0',
    msg: '830f559b22b74bfcbb5631fae20462cb',
    salt: 'b001ebc1-f2f7-44de-b5ff-1da0b16af05b',
    sign: '0xa989699af09ef6cdfcec182b1043c5acd5b565a5444ea0cecb8dfbd0509695ac3d9905b32246774d46af617d122cdcea3a273a71287082feffaa4c3b9fccd82f01'
  }
```

## Performance
A signing test with 100k records was performed on GCP k8s cluster
with 2 VMs each with 2 vCPU and 4 GM memory. 8 signing pods
were used. The test showed
about 10 signing/per second rate. The bottleneck was mongodb
running at full 2 vCPU (2 vCPU is equivalent to 1 real core and 1 hyper core).
Mongodb can support upt to 40k/writes per second, but to achive
such speed more powerful hardware is required


## [Development Guide](DEVELOP.md)

## [K8s Deployment Guide](K8S.md)


