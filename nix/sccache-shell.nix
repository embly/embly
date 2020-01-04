with import <nixpkgs> { };

stdenv.mkDerivation {
  name = "rust-env";
  nativeBuildInputs = [ bash go ];
  # Set Environment Variables
  RUST_BACKTRACE = 1;
  RUSTC_WRAPPER = "${pkgs.sccache}/bin/sccache";
  CARGO_TARGET_DIR = /root/target;

}
