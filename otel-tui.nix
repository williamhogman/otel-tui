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

  overrideModAttrs = (
    _: {
      buildPhase = ''
        go work vendor
      '';
    }
  );
  ldflags = [
    "-X main.version=${version}"
  ];
  vendorHash = "sha256-CJXYa3CzKOkZKJf5ukmAoA0kSHWtEufe3FQgV3Z1hQQ=";
  subPackages = [ "." ];
}
