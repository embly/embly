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




## nix notes

```bash
# collect garbage and delete previous generations
nix-collect-garbage -d

# list packages that depend on this one
nix-store -q --referrers /nix/store/jbm5xa

# list packages that this package depends on
nix-store -q --references ./result | grep ""
```


## build implementation notes

use a nix shell --pure, which will then need to reference build files
probably just write them to the embly homedir
then run nix-shell from the homedir, which then runs a command in the tmp files
will need to pass an arg to the nix-shell script to point to target dir and cargo dir

make sure we test this in docker to ensure it doesn't rely on global state

nix-shell can also handle installing dependencies. maybe best to just use it this way so that we don't have to keep build references around? not sure

## cachix

```bash
   0 nix-env -iA cachix -f https://cachix.org/api/v1/install
   1 cachix use embly
   2 nix-env --uninstall cachix
   3 nix-collect-garbage
   4 nix-build ~/embly/protoc.nix
   5 ncdu
   6 cat /etc/nix/nix.conf
   7 nix-env -iA cachix -f https://cachix.org/api/v1/install
   8 export CACHIX_SIGNING_KEY=[removed]
   9 nix-build ~/embly/lucetc.nix | cachix push embly
  10 nix-build ~/embly/wrapper.nix | cachix push embly
  11 history
  ```
