with import <nixpkgs> { };

with pkgs;

rec {
  shell = stdenv.mkDerivation {
    name = "rust-env";
    nativeBuildInputs = [
      rustc
      cargo

    ];
    buildInputs = [ ];

    # Set Environment Variables
    RUST_BACKTRACE = 1;
    RUSTC_WRAPPER = "${pkgs.sccache}/bin/sccache";
  };

  wrapper = rustPlatform.buildRustPackage rec {
    src = "./";
    buildInputs = [ shell ];
    verifyCargoDeps = false;

    meta = with stdenv.lib; { description = "hello"; };
  };

}
