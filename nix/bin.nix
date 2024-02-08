{ 
    stdenv, 
    fetchurl, 
    version, 
    appName,
    binUrl,
    sha256,
    dontUnpack ? true,
}:
stdenv.mkDerivation {
  inherit dontUnpack;
  # Use this nix file in case you want to add a compiled binary 
  # to the Nix environment
  name = "${appName}-${version}";
  
  # Define the URL to download the compiled binary
  src = fetchurl {
    url = "${binUrl}";
    sha256 = "${sha256}";
  };

  # Install the binary to the Nix environment's bin directory
  installPhase = ''
    mkdir -p $out/bin
    cp $src $out/bin/${appName}
    chmod +x $out/bin/${appName}
  '';
}
