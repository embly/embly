self: super:

let
  src = super.fetchFromGitHub {
    owner = "mozilla";
    repo = "nixpkgs-mozilla";
    # commit from: 2019-05-15
    rev = "9f35c4b09fd44a77227e79ff0c1b4b6a69dff533";
    sha256 = "18h0nvh55b5an4gmlgfbvwbyqj91bklf1zymis6lbdh75571qaz0";
  };
  lucetc = with super;
    stdenv.mkDerivation {
      pname = "lucetc";
      version = "0.4.1";
      buildInputs = [ cargo rustc cmake ];
      src = fetchurl {
        url = "https://static.crates.io/crates/lucetc/lucetc-0.4.1.crate";
        sha256 = "06v05s9jma84x3fvpm9xnq1lnjd5pfz2cvvg7mvbg7x9mapz574i";
      };
      unpackCmd = "tar -zxvf $src";
      dontConfigure = true;
      buildPhase = ''
        env
                cargo build --release
              '';
      installPhase = ''
        mkdir -p $out/bin
        mv ./target/release/lucetc $out/bin
        strip $out/bin/lucetc
      '';
    };
  lucetc2 = with super;
    (runCommandCC "lucetc-0.4.1" {
      buildInputs = [ cacert ];
      propagatedBuildInputs = [ stdenv.cc.cc ];
    } ''
      ${wget}/bin/wget https://embly-static.s3.amazonaws.com/lucetc-0.4.1

      patchelf --set-interpreter "$(cat $NIX_CC/nix-support/dynamic-linker)" lucetc-0.4.1
      patchelf --set-rpath ${stdenv.cc.cc.lib}/lib64 lucetc-0.4.1
      mkdir -p $out/bin
      mv lucetc-0.4.1 $out/bin/lucetc
      chmod +x $out/bin/lucetc
    '');
  wrapper2 = with super;
    (runCommandCC "embly-wrapper-0.0.2" {
      buildInputs = [ cacert ];
      propagatedBuildInputs = [ stdenv.cc.cc ];
    } ''
      ${wget}/bin/wget https://embly-static.s3.amazonaws.com/embly-wrapper-0.0.2

      patchelf --set-interpreter "$(cat $NIX_CC/nix-support/dynamic-linker)" embly-wrapper-0.0.2
      patchelf --set-rpath ${stdenv.cc.cc.lib}/lib64 embly-wrapper-0.0.2
      mkdir -p $out/bin
      mv embly-wrapper-0.0.2 $out/bin/embly-wrapper
      chmod +x $out/bin/embly-wrapper
    '');
  wrapper = with super;
    stdenv.mkDerivation {
      pname = "embly-wrapper";
      version = "0.0.2";
      buildInputs = [ cargo rustc ];
      src = fetchurl {
        url =
          "https://static.crates.io/crates/embly-wrapper/embly-wrapper-0.0.2.crate";
        sha256 = "08l0s3cw6h44r2mnx1x4409pzp5wm20mslq1shxl2b0crjhrinnh";
      };
      unpackCmd = "tar -zxvf $src";
      dontConfigure = true;
      buildPhase = ''
        cargo build --release
      '';
      installPhase = ''
        mkdir -p $out/bin
        mv ./target/release/embly-wrapper $out/bin
        strip $out/bin/embly-wrapper
      '';
    };
in with import "${src.out}/rust-overlay.nix" super super; rec {
  emblyRust = (rustChannelOfTargets "nightly" "2019-12-30" [ "wasm32-wasi" ]);
  emblyLucetc = lucetc2;
  emblyWrapper = wrapper2;
}
