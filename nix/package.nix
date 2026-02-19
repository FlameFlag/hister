{
  lib,
  buildGoModule,
  buildNpmPackage,
  importNpmLock,
  sqlite,
  histerRev ? "unknown",
}:
let
  packageJson = builtins.fromJSON (builtins.readFile ../ext/package.json);
  frontend = buildNpmPackage {
    pname = "hister-frontend";
    version = packageJson.version;
    src = ../server/web;
    npmDeps = importNpmLock {
      npmRoot = ../server/web;
    };
    npmConfigHook = importNpmLock.npmConfigHook;
    dontNpmInstall = true;
    installPhase = ''
      runHook preInstall
      mkdir -p "$out"
      cp -r dist/* "$out"
      runHook postInstall
    '';
  };
in
buildGoModule (finalAttrs: {
  pname = "hister";
  version = packageJson.version;

  src = ../.;

  vendorHash = "sha256-JwLV+rXRKF5ynJ6P3EYkzKecnAekDn2QkSJudPA0eQ0=";

  proxyVendor = true;

  buildInputs = [ sqlite ];

  env.CGO_ENABLED = "1";

  preBuild = ''
    export CGO_CFLAGS="-I${sqlite.dev}/include"
    export CGO_LDFLAGS="-L${sqlite.out}/lib -lsqlite3"
    mkdir -p server/web
    cp -r ${frontend} server/web/dist
  '';

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${finalAttrs.version}"
    "-X main.commit=${histerRev}"
  ];

  doCheck = false;

  meta = {
    description = "Web history on steroids - blazing fast, content-based search for visited websites";
    homepage = "https://github.com/asciimoo/hister";
    license = lib.licenses.agpl3Only;
    maintainers = [ lib.maintainers.FlameFlag ];
    mainProgram = "hister";
    platforms = lib.platforms.unix;
  };
})
