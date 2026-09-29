#!/bin/sh

set -eu

BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
BINARY_PATH="$BIN_DIR/grove"
TARGET_SHELL="${1:-}"

PLATFORM="$(uname -s)"
case "$PLATFORM" in
    Darwin|Linux)
        ;;
    *)
        printf '%s\n' "unsupported platform: $PLATFORM" >&2
        exit 1
        ;;
esac

if ! command -v go >/dev/null 2>&1; then
    printf '%s\n' "go is required to install grove" >&2
    exit 1
fi

# Build from the repo root regardless of where the script is invoked from.
script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(dirname "$script_dir")

mkdir -p "$BIN_DIR"
(cd "$repo_root" && go build -o "$BINARY_PATH" .)
chmod +x "$BINARY_PATH"

if [ -z "$TARGET_SHELL" ]; then
	TARGET_SHELL=$(basename "${SHELL:-bash}")
fi

# bash_config_path picks the startup file bash will actually read. On macOS,
# terminals start bash as a login shell, which reads the first existing file of
# .bash_profile, .bash_login, .profile — never .bashrc — so append to the file
# bash itself would pick (defaulting to .bash_profile so we never shadow an
# existing .profile that isn't there). On Linux, interactive shells read .bashrc.
bash_config_path() {
    if [ "$PLATFORM" != "Darwin" ]; then
        printf '%s' "$HOME/.bashrc"
        return
    fi
    for candidate in "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
        if [ -f "$candidate" ]; then
            printf '%s' "$candidate"
            return
        fi
    done
    printf '%s' "$HOME/.bash_profile"
}

case "$TARGET_SHELL" in
    zsh)
        config_path="$HOME/.zshrc"
        ;;
    bash)
        config_path=$(bash_config_path)
        ;;
    *)
        printf '%s\n' "unsupported shell for install.sh: $TARGET_SHELL" >&2
        printf '%s\n' "supported shells: bash, zsh" >&2
        exit 1
        ;;
esac

# grove owns this file outright: rewrite it on every run. It defines the shell
# wrapper that turns the path grove prints on stdout into a directory change,
# so the rc file only ever needs the stable source line below.
data_dir="${XDG_DATA_HOME:-$HOME/.local/share}/grove"
init_file="$data_dir/init.sh"
mkdir -p "$data_dir"
cat > "$init_file" <<EOF
grove() {
    local output
    output="\$("$BINARY_PATH" "\$@")"
    local status=\$?
    if [ \$status -ne 0 ]; then
        return \$status
    fi
    if [ -n "\$output" ]; then
        cd "\$output" || return 1
    fi
}
EOF

source_line="[ -f \"$init_file\" ] && . \"$init_file\""

append_line_once() {
    target_file=$1
    line=$2

    mkdir -p "$(dirname "$target_file")"
    if [ ! -f "$target_file" ]; then
        : > "$target_file"
    fi
    if grep -Fqx "$line" "$target_file"; then
        return
    fi
    printf '\n%s\n' "$line" >> "$target_file"
}

append_line_once "$config_path" "$source_line"
printf '%s\n' "installed grove to $BINARY_PATH, wrote $init_file, and updated $config_path"
