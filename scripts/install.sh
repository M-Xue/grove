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
        config_path="$HOME/.zshrc"
        ;;
    bash|sh)
        config_path="$HOME/.bashrc"
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
