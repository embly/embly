with import <nixpkgs> { crossSystem = { config = "x86_64-apple-darwin"; }; };

mkShell {
  buildInputs = [ bash ]; # your dependencies here
}
