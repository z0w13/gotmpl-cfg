{ lib, buildGoModule }:
buildGoModule {
  pname = "gotmpl-cfg";
  version = "unstable";

  src = ./.;
  vendorHash = null; # NOTE: no deps, so null

  meta = {
    description = "simple command-line utility to generate configs from templates";
    homepage = "https://github.com/z0w13/gotmpl-cfg";
    license = lib.licenses.cc-by-nc-sa-40;
    maintainers = [ ];
  };
}
