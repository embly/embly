with import <nixpkgs> { };

let
  lucetFile = pkgs.fetchurl {
    url = "https://embly-static.s3.amazonaws.com/lucetc-0.5.0";
    sha256 = "0lrcbv2agx6k36hwxarm4s0mijp8ikx9041jqr72qgq9dwm0s39z";
  };
in (runCommandCC "lucetc-0.5.0" {
  buildInputs = [ cacert ];
  propagatedBuildInputs = [ stdenv.cc.cc ];
} ''
  cp ${lucetFile} ./lucetc-0.5.0
  chmod +w lucetc-0.5.0
  patchelf --set-interpreter "$(cat $NIX_CC/nix-support/dynamic-linker)" lucetc-0.5.0
  patchelf --set-rpath ${stdenv.cc.cc.lib}/lib64 lucetc-0.5.0
  mkdir -p $out/bin
  mv lucetc-0.5.0 $out/bin/lucetc
  chmod +x $out/bin/lucetc
'')
