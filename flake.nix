{
  description = "KnowHub lightweight development shell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { nixpkgs, ... }:
    let
      systems = [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              nodejs
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              grpcurl
              gnumake
              git
              curl
              jq
              postgresql
              zsh
              neovim
              tmux
              ripgrep
              fd
              fzf
              eza
              bat
              dnsutils
              zoxide
              autojump
              thefuck
            ];

            GOTOOLCHAIN = "local";

            shellHook = ''
              export GOCACHE="$PWD/.gocache"
              export PATH="$PWD/frontend/node_modules/.bin:$PATH"
              export PATH="$PATH:/opt/homebrew/bin:/usr/local/bin"

              echo "KnowHub dev shell"
              echo "  make proto"
              echo "  nix develop -c make proto"
              echo "  nix develop -c zsh -l"
            '';
          };
        });
    };
}
