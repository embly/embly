with import <nixpkgs> { };

let
  lucetFile = pkgs.fetchurl {
    url = "https://embly-static.s3.amazonaws.com/lucetc-0.4.1";
    sha256 = "0swmsngxh0p3lk5r88s0kp8nk934qkq52nqnjasa2cv08ghr9w3j";
  };
in (runCommandCC "lucetc-0.4.1" {
  buildInputs = [ cacert ];
  propagatedBuildInputs = [ stdenv.cc.cc ];
} ''
  cp ${lucetFile} ./lucetc-0.4.1
  chmod +w lucetc-0.4.1
  patchelf --set-interpreter "$(cat $NIX_CC/nix-support/dynamic-linker)" lucetc-0.4.1
  patchelf --set-rpath ${stdenv.cc.cc.lib}/lib64 lucetc-0.4.1
  mkdir -p $out/bin
  mv lucetc-0.4.1 $out/bin/lucetc
  chmod +x $out/bin/lucetc
'')
