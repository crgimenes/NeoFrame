{pkgs, ...}:
pkgs.mkShell {
  name = "NeoFrame";
  version = "0.0.1";

  LD_LIBRARY_PATH = "$LD_LIBRARY_PATH:${
    with pkgs;
      pkgs.lib.makeLibraryPath [libGL]
  }";

  buildInputs = with pkgs; [
    go
    gopls
    gotools

    gcc

    xorg.libXrandr
    xorg.libXxf86vm
    xorg.libX11
    xorg.libXcursor
    xorg.libXinerama
    xorg.libXi
    xorg.xinput

    libGL
  ];
}
