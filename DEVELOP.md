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
GET    /stats           # show signed and unsigned records         
```
Examples:

```
# get stats /stats
$ curl localhost:8080/stats

$ curl localhost:8080/stats
stats: {"signed_records":1400000,"unsigned_records":0}

```