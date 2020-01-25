with import <nixpkgs> { };
let
  rust = import ./rust.nix;
  lucetc = import ./lucetc-0.5.0.nix;
in with pkgs;
stdenv.mkDerivation {
  name = "rust-env";
  buildInputs = [
    rust
    lucetc
    wabt
    # Add some extra dependencies from `pkgs`
    pkgconfig
    openssl
  ];

  # Set Environment Variables
  RUST_BACKTRACE = 1;
}
