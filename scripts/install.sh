#!/bin/sh

set -eu

BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
BINARY_PATH="$BIN_DIR/grove"
TARGET_SHELL="${1:-}"

case "$(uname -s)" in
    Darwin|Linux)
        ;;
    *)
        printf '%s\n' "unsupported platform: $(uname -s)" >&2
        exit 1
        ;;
esac

if ! command -v go >/dev/null 2>&1; then
    printf '%s\n' "go is required to install grove" >&2
    exit 1
fi

mkdir -p "$BIN_DIR"
go build -o "$BINARY_PATH" .
chmod +x "$BINARY_PATH"

if [ -z "$TARGET_SHELL" ]; then
	TARGET_SHELL=$(basename "${SHELL:-bash}")
fi

case "$TARGET_SHELL" in
    zsh)
        init_shell="zsh"
        config_path="$HOME/.zshrc"
        ;;
    bash|sh)
        init_shell="bash"
        config_path="$HOME/.bashrc"
        ;;
    *)
        printf '%s\n' "unsupported shell for install.sh: $TARGET_SHELL" >&2
        printf '%s\n' "supported shells: bash, zsh" >&2
        exit 1
        ;;
esac

# grove owns this file outright: rewrite it on every run. It evaluates the
# wrapper straight from the binary, so the binary stays the single source of
# truth and the rc file only ever needs the stable source line below.
data_dir="${XDG_DATA_HOME:-$HOME/.local/share}/grove"
init_file="$data_dir/init.sh"
mkdir -p "$data_dir"
printf 'eval "$(%s shell-init %s)"\n' "$BINARY_PATH" "$init_shell" > "$init_file"

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
