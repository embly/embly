with import <nixpkgs> { };
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
'')
