{
  lib,
  stdenvNoCC,
  buildGoModule,
  nodejs,
  pnpm,
  fetchPnpmDeps,
  pnpmConfigHook,
  pkg-config,
  dbus,
  sipp,
  wrapGAppsHook3,
  gtk3,
  webkitgtk_4_1,
  libpulseaudio,
  alsa-lib,
  libopus,
  opusfile,
  libogg,
  speexdsp,
  libsecret,
  makeDesktopItem,
  copyDesktopItems,
}:
let
  version = "0.1.0";
  source = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.unions [
      ../go.mod
      ../go.sum
      ../cmd
      ../internal
      ../pkg
      ../LICENSE
      (lib.fileset.difference ../web (
        lib.fileset.unions [
          (lib.fileset.maybeMissing ../web/node_modules)
          (lib.fileset.maybeMissing ../web/dist)
        ]
      ))
      ../build/package/appicon.png
      ../build/package/embed.go
    ];
  };
  frontendSource = lib.cleanSourceWith {
    src = ../web;
    filter =
      path: type:
      !(builtins.elem (baseNameOf path) [
        "node_modules"
        "dist"
        "package-lock.json"
        "package.json.md5"
      ]);
  };
  frontend = stdenvNoCC.mkDerivation {
    pname = "voiper-frontend";
    outputs = [
      "out"
      "licenses"
    ];
    inherit version;
    src = frontendSource;
    nativeBuildInputs = [
      nodejs
      pnpm
      pnpmConfigHook
    ];
    pnpmDeps = fetchPnpmDeps {
      pname = "voiper-frontend";
      src = frontendSource;
      inherit pnpm;
      fetcherVersion = 4;
      hash = "sha256-tVRktjzZsZXWHgdZt6l0XQ6N9tFdwkvyCmpkRYWZa5w=";
    };
    buildPhase = ''
      runHook preBuild
      pnpm run build
      runHook postBuild
    '';
    installPhase = ''
      runHook preInstall
      cp -r dist $out
      mkdir -p $licenses
      while IFS= read -r -d "" notice; do
        install -Dm644 "$notice" "$licenses/''${notice#node_modules/.pnpm/}"
      done < <(find node_modules/.pnpm -type f \( -iname 'LICENSE*' -o -iname 'COPYING*' -o -iname 'NOTICE*' \) -print0)
      runHook postInstall
    '';
  };
in
buildGoModule {
  pname = "voiper";
  inherit version;
  src = source;
  vendorHash = "sha256-Fx6mcfAGJeUqVNkL8SXzRJBnKnLJsyjhlg0pVxFfcqY=";
  nativeBuildInputs = [
    pkg-config
    dbus
    wrapGAppsHook3
    copyDesktopItems
  ];
  buildInputs = [
    gtk3
    webkitgtk_4_1
    libpulseaudio
    alsa-lib
    libopus
    opusfile
    libogg
    speexdsp
    libsecret
  ];
  env = {
    CGO_ENABLED = "1";
    GOTOOLCHAIN = "local";
  };
  tags = [
    "desktop"
    "production"
    "webkit2_41"
    "opus"
    "speex"
    "secretservice"
  ];
  nativeCheckInputs = [ sipp ];
  subPackages = [ "cmd/voiper" ];
  ldflags = [
    "-X github.com/megakuul/voiper/internal/version.VersionOverride=v${version}"
    "-s"
    "-w"
  ];
  preBuild = ''
    mkdir -p web/dist
    cp -r ${frontend}/. web/dist/
  '';
  checkPhase = ''
    runHook preCheck
    go test -race -timeout=3m -tags=webkit2_41,opus,speex,secretservice ./pkg/... ./internal/... ./cmd/...
    go vet -tags=webkit2_41,opus,speex,secretservice ./...
    runHook postCheck
  '';
  desktopItems = [
    (makeDesktopItem {
      name = "voiper";
      desktopName = "Voiper";
      comment = "SIP desktop phone";
      exec = "voiper %u";
      mimeTypes = [
        "x-scheme-handler/sip"
        "x-scheme-handler/sips"
        "x-scheme-handler/tel"
      ];
      icon = "voiper";
      categories = [
        "Network"
        "Telephony"
      ];
      terminal = false;
    })
  ];
  postInstall = ''
    install -Dm644 build/package/appicon.png $out/share/pixmaps/voiper.png
    install -Dm644 LICENSE $out/share/licenses/voiper/LICENSE
    while IFS= read -r -d "" notice; do
      install -Dm644 "$notice" "$out/share/doc/voiper/third-party/''${notice#vendor/}"
    done < <(find vendor -type f \( -iname 'LICENSE*' -o -iname 'COPYING*' -o -iname 'NOTICE*' \) -print0)
    install -Dm644 vendor/github.com/gotranspile/g722/README.md $out/share/doc/voiper/third-party/github.com/gotranspile/g722/README.md
    cp -r ${frontend.licenses} $out/share/doc/voiper/third-party/frontend
  '';
  preFixup = ''
    gappsWrapperArgs+=(--prefix LD_LIBRARY_PATH : "${
      lib.makeLibraryPath [
        libpulseaudio
        alsa-lib
      ]
    }")
  '';
  passthru = { inherit frontend; };
  meta = {
    description = "Linux SIP desktop softphone";
    homepage = "https://github.com/megakuul/voiper";
    license = lib.licenses.mit;
    mainProgram = "voiper";
    platforms = [ "x86_64-linux" ];
  };
}
