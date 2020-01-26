with import <nixpkgs> { };
let rust = import ./rust.nix;
in with pkgs; [ rust gcc ]
