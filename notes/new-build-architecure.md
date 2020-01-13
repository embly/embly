Builds are managed by nix. We set up a local build server in a docker container that runs nix. 
Nix pulls different dependencies and tools for different build setups. We manage versions by 
pinning the has of the nix channel or writing our own deps.

This allow easy decoupling of the files and build system, the embly command is responsible for 
sending files to the build server and the build server handles caching, build optimization
and returning compiled wasm.

Since we're using Nix this could also work on OSX without virtualization. 
