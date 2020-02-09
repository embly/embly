with import <nixpkgs> { };

(runCommandCC "protoc" { } ''
  set -x
  mkdir -p $out/bin $out/lib

  cp ${protobuf}/lib/libprotoc.so.18 $out/lib/
  cp ${protobuf}/lib/libprotobuf.so.18 $out/lib/
  cp ${protobuf}/bin/protoc $out/bin/
  chmod +w $out/bin/protoc
  chmod +w $out/lib/libprotoc.so.18
  chmod +w $out/lib/libprotobuf.so.18
  patchelf --set-rpath $out/lib/:${stdenv.cc.cc.lib}/lib64:${zlib}/lib $out/bin/protoc
  patchelf --set-rpath $out/lib/:${stdenv.cc.cc.lib}/lib64:${zlib}/lib $out/lib/libprotoc.so.18
  strip $out/bin/protoc $out/lib/*

  patchelf --shrink-rpath $out/bin/protoc
  patchelf --shrink-rpath $out/lib/libprotoc.so.18
  patchelf --shrink-rpath $out/lib/libprotobuf.so.18
  set +x
'')
