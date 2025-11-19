{
  config,
  pkgs,
  src ? ./.,
  version ? "v0.7.0-dev",
  ...
}:

pkgs.buildGoModule {
  pname = "otel-tui";
  inherit version src;

  modBuildPhase = ''
    runHook preBuild
    rm -rf vendor
    export GIT_SSL_CAINFO=$NIX_SSL_CERT_FILE
    go work vendor
    runHook postBuild
  '';
  ldflags = [
    "-X main.version=${version}"
  ];
  vendorHash = "sha256-GuniT7fPY/fQb1OCMTNxIMny2c8Wl3JFdqfSuAgli0k=";
  subPackages = [ "." ];
}
