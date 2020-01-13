## Nix-based build architecture

Use nix to manage dependency versions and build tools for various languages. Allow it to work locally and over grpc. Allow for it to work on a computer that has nix installed or in a docker container (using grpc).


Three build use-cases to support:
 - local computer with nix installed (darwin or linux)
 - local computer in a docker container
 - remote build server

Should make sure that all build actions are triggered by protobuf structs to allow
for easy decoupling.

Client:
 - listens for file changes
 - writes object files to disk
 - uploads file contents and keeps file hashes
 - tracks source dependency graph

Server:
 - tracks build space and building files in that space
 - loads necessary dependencies for build and version
 - builds wasm and does post-processing
 - cross-compiles lucetc


startup:
 - crawl for embly.hcl
 - index file tree (ignoring gitignore?)
 - if nix is available downloaded needed dependencies
 - prompt the user for docker or nix version? write to config file
 - if there's not nix start downloading the docker version
 - build all projects, oh, still build in a tmpdir no matter what, so that we can copy files and guarantee isolation?
 - (consider rewriting stack trace file locations in actual location)
 - copy resulting object files to correct directory and alert client

