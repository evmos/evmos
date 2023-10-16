{ pkgs
, config
, evmos ? (import ../. { inherit pkgs; })
}: rec {
  start-evmos = pkgs.writeShellScriptBin "start-evmos" ''
    # rely on environment to provide evmosd
    export PATH=${pkgs.test-env}/bin:$PATH
    ${../scripts/start-evmos.sh} ${config.evmos-config} ${config.dotenv} $@
  '';
  start-geth = pkgs.writeShellScriptBin "start-geth" ''
    export PATH=${pkgs.test-env}/bin:${pkgs.go-ethereum}/bin:$PATH
    source ${config.dotenv}
    ${../scripts/start-geth.sh} ${config.geth-genesis} $@
  '';
  start-scripts = pkgs.symlinkJoin {
    name = "start-scripts";
    paths = [ start-evmos start-geth ];
  };
}
