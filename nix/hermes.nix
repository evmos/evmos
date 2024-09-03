{ src
, lib
, stdenv
, darwin
, symlinkJoin
, openssl
, platform
}:
platform.buildRustPackage rec {
  name = "hermes";

  inherit src;
  cargoSha256 = "sha256-4E1FYffh7qu5e09UjYw83eW9uPuKHGfOAaZbLtLCLA0=";
  cargoBuildFlags = "--no-default-features --bin hermes";
  buildInputs = lib.optionals stdenv.isDarwin [
    darwin.apple_sdk.frameworks.Security
    darwin.apple_sdk.frameworks.SystemConfiguration
    darwin.libiconv
  ];
  doCheck = false;
  # RUSTFLAGS = "--cfg ossl111 --cfg ossl110 --cfg ossl101";
  OPENSSL_NO_VENDOR = "1";
  OPENSSL_DIR = symlinkJoin {
    name = "openssl";
    paths = with openssl; [ out dev ];
  };
}
