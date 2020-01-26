with import <nixpkgs> { };
with pkgs;
stdenv.mkDerivation {
  name = "cctools";
  nativeBuildInputs = [ autoconf automake libtool autoreconfHook ];
  buildInputs = [ libuuid clang ];
  src = fetchFromGitHub {
    owner = "tpoechtrager";
    repo = "cctools-port";
    rev = "ae83e7fcdeb2b8bd838b6008f1972bd853fa0e76";
    sha256 = "1zncypdn3a8cqlhzqa975d342yclw5ncm7l3g7rb9n8pp860pcxm";
  };
  #   configureFlags = [ "--target=x86_64-apple-darwin" ];
  postPatch = ''
    export CC=clang
    export CXX=clang++
    cd cctools
  '';
}

# with import <nixpkgs> { };
# with pkgs;
# (runCommandCC "cctools" {
#   buildInputs = [ clang ];
#   src = fetchFromGitHub {
#     owner = "tpoechtrager";
#     repo = "cctools-port";
#     rev = "ae83e7fcdeb2b8bd838b6008f1972bd853fa0e76";
#     sha256 = "1zncypdn3a8cqlhzqa975d342yclw5ncm7l3g7rb9n8pp860pcxm";
#   };
# } ''
#   cp -r $src/cctools .
#   chmod -R +w .
#   cd cctools
#   export CC=clang
#   ./configure  --target=x86_64-apple-darwin11

# '')
