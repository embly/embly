with import <nixpkgs> { };

stdenv.mkDerivation {
  name = "rust-env";
  nativeBuildInputs = [
    rustc
    cargo

  ];
  buildInputs = [ ];

  # Set Environment Variables
  RUST_BACKTRACE = 1;
  RUSTC_WRAPPER = "${pkgs.sccache}/bin/sccache";
}
