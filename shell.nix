{pkgs ? (import <nixpkgs> { config.allowUnfree = true; })}:
with pkgs;
  stdenvNoCC.mkDerivation {
    name = "shell";
    buildInputs = [
      bluez
      golangci-lint
      picotool
      pioasm
    ];
  }
