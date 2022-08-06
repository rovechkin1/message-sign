# Local development guide

## Directories
* service - signing service 
* record-generator - record generator 
* key-generator - signing key generator
* charts - k8s helm charts

## Build
```
# build service
make build-service
```

```
# build key generator utility, writes keys to a file keys.csv
make build-key-gen
```

```
# build record generator which writes records to mongo db
make build-record-gen
```

## Setup
```
# generate keys and save them in keys.csv
bin/key-generator

# start mongo db
./start-local-mongo.sh

# populate mongodb with records
bin/record-generator

# start service
bin/service

```

## API
```
GET    /                # liveness         
GET    /sign/:size      # signing request, size if a batch size      
GET    /stats           # show signed and unsigned records         
GET    /batch/:batchId/:batchCount/:key # internal endpoint to launch batch signing
```
Examples:

```
# sign records /sign/<batch_size>
curl localhost:8080/sign/30000

# get stats /stats
curl localhost:8080/stats
ruslans-MBP:message-sign ruslan$ curl localhost:8080/stats
stats: {"total_records":1000,"signed_records":800,"unsigned_records":200}

```