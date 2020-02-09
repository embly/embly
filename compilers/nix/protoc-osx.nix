with import <nixpkgs> { };

(runCommandCC "protoc" { } ''
  set -x
  mkdir -p $out/bin $out/lib
  ls -lah ${protobuf}/lib/
  ls -lah ${protobuf}/bin/
  otool -L ${protobuf}/bin/protoc
  otool -l ${protobuf}/bin/protoc
  cp ${protobuf}/lib/libprotoc.18.dylib $out/lib/
  cp ${protobuf}/lib/libprotobuf.18.dylib $out/lib/
  cp ${protobuf}/bin/protoc $out/bin/
  chmod +w $out/bin/protoc
  chmod +w $out/lib/libprotoc.18.dylib
  chmod +w $out/lib/libprotobuf.18.dylib
  install_name_tool -delete_rpath /nix/store/vfp33xyy80wmx3c725wsxd0rcvxhanws-swift-corefoundation/Library/Frameworks $out/bin/protoc
  install_name_tool -add_rpath $out/lib/:${stdenv.cc.cc.lib}/lib64:${zlib}/lib $out/bin/protoc
  install_name_tool -add_rpath $out/lib/:${stdenv.cc.cc.lib}/lib64:${zlib}/lib $out/lib/libprotoc.18.dylib
  # strip $out/bin/protoc $out/lib/*

  # patchelf --shrink-rpath $out/bin/protoc
  # patchelf --shrink-rpath $out/lib/libprotoc.18.dylib
  # patchelf --shrink-rpath $out/lib/libprotobuf.18.dylib
  set +x
'')
