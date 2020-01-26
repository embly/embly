with import <nixpkgs> { };
let lucetc = import ./lucetc-0.5.0.nix;
in with pkgs; [ lucetc wabt busybox gcc binutils ]
