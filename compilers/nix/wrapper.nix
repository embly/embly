with import <nixpkgs> { };

let
  wrapperFile = pkgs.fetchurl {
    url = "https://embly-static.s3.amazonaws.com/embly-wrapper-0.0.2";
    sha256 = "0k7sjyrh2rz45wma4m5hmil2jbi3k6sw212jbw3shfg10shkhjf6";
  };
in (runCommandCC "embly-wrapper-0.0.2" {
  buildInputs = [ cacert ];
  propagatedBuildInputs = [ stdenv.cc.cc ];
} ''
  cp ${wrapperFile} ./embly-wrapper-0.0.2
  chmod +w ./embly-wrapper-0.0.2
  patchelf --set-interpreter "$(cat $NIX_CC/nix-support/dynamic-linker)" embly-wrapper-0.0.2
  patchelf --set-rpath ${stdenv.cc.cc.lib}/lib64 embly-wrapper-0.0.2
  mkdir -p $out/bin
  mv embly-wrapper-0.0.2 $out/bin/embly-wrapper
  chmod +x $out/bin/embly-wrapper
'')
