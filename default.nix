{ buildGoModule, libpcap }:
buildGoModule {
  name = "go-babel";
  src = ./.;
  vendorHash = "sha256-1ntfNzHLv46ni4wOLMoX1fSaZe+kfvIUzeJnrAnglZQ=";
  buildInputs = [ libpcap ];
  doCheck = false;
}
