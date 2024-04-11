{
  buildGoModule,
  lib,
  ...
}:
buildGoModule {
  pname = "NeoFrame";
  version = "0.0.1";

  src = ./.;

  subPackages = ["."];

  vendorHash = null;
}
