


api layer

 /build
 /function/:id
 /auth
 /invoke -> needs a stream of bytes and a side channel, probably websockets and a grpc proxy
 
 
build ->
 - compile rust into wasm
     - not safe. disable dependencies for now
     - try and allow only pure rust deps 
 - wasm to lucet
     - this step is safe
     - can share machine
     - can the redis cache be used for this?

 - grpc to both using a stream. stream breaks assume dead.
 - hold stream open for build. 
 - build is either persisted in a db or the request is held open
 - build results (maybe even cache results) are persisted in s3 for rebuilding

view ->
 - hold the code in s3 and have redis act as a cache for quick responses
 
invoke ->
 - invocation pulls the binary from redis and runs it on a machine. 
 - if it's grpc we might be able to implement some kind of simple load balancing to favor certain key ranges
