{ src
, lib
, stdenv
, darwin
, rustPlatform
, symlinkJoin
, openssl
}:
rustPlatform.buildRustPackage rec {
  name = "hermes";
  inherit src;
  cargoSha256 = "sha256-DBHDJlD2kTUeU0LCG5ZUij+v+lWUYRPGVvUKUYPfYQA=";
  cargoBuildFlags = "-p ibc-relayer-cli";
  buildInputs = lib.optionals stdenv.isDarwin [
    darwin.apple_sdk.frameworks.Security
    darwin.libiconv
  ];
  doCheck = false;
  RUSTFLAGS = "--cfg ossl111 --cfg ossl110 --cfg ossl101";
  OPENSSL_NO_VENDOR = "1";
  OPENSSL_DIR = symlinkJoin {
    name = "openssl";
    paths = with openssl; [ out dev ];
  };
}
