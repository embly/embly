with import <nixpkgs> { };

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
'')

