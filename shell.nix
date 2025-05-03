{pkgs ? (import <nixpkgs> { config.allowUnfree = true; })}:
with pkgs;
  mkShell {
    buildInputs = [
      bluez
      golangci-lint
      picotool
      pioasm
    ];
  }
