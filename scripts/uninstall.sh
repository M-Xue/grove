#!/bin/sh

# Removes everything install.sh created: the binary, the grove data directory
# (wrapper init file), the source line in shell startup files, and grove's
# worktree cache. Honors the same BIN_DIR and XDG_DATA_HOME overrides as
# install.sh, so export them here if you set them at install time.

set -eu

BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
BINARY_PATH="$BIN_DIR/grove"

PLATFORM="$(uname -s)"
case "$PLATFORM" in
    Darwin|Linux)
        ;;
    *)
        printf '%s\n' "unsupported platform: $PLATFORM" >&2
        exit 1
        ;;
esac

data_dir="${XDG_DATA_HOME:-$HOME/.local/share}/grove"
init_file="$data_dir/init.sh"

# Mirrors Go's os.UserCacheDir, which the cache package builds on.
if [ "$PLATFORM" = "Darwin" ]; then
    cache_dir="$HOME/Library/Caches/grove"
else
    cache_dir="${XDG_CACHE_HOME:-$HOME/.cache}/grove"
fi

source_line="[ -f \"$init_file\" ] && . \"$init_file\""

# remove_line_if_present filters the exact source line out of a startup file.
# It writes through the original file rather than replacing it, so symlinked
# dotfiles and file permissions survive.
remove_line_if_present() {
    target_file=$1
    line=$2

    if [ ! -f "$target_file" ]; then
        return
    fi
    if ! grep -Fqx "$line" "$target_file"; then
        return
    fi
    tmp_file="$target_file.grove-uninstall.tmp"
    grep -Fvx "$line" "$target_file" > "$tmp_file" || true
    cat "$tmp_file" > "$target_file"
    rm -f "$tmp_file"
    printf '%s\n' "removed grove source line from $target_file"
}

if [ -e "$BINARY_PATH" ]; then
    rm -f "$BINARY_PATH"
    printf '%s\n' "removed $BINARY_PATH"
fi
if [ -d "$data_dir" ]; then
    rm -rf "$data_dir"
    printf '%s\n' "removed $data_dir"
fi
if [ -d "$cache_dir" ]; then
    rm -rf "$cache_dir"
    printf '%s\n' "removed $cache_dir"
fi

# Scrub every startup file install.sh may have written to, regardless of which
# shell was targeted at install time.
for config_path in \
    "$HOME/.zshrc" \
    "$HOME/.bashrc" \
    "$HOME/.bash_profile" \
    "$HOME/.bash_login" \
    "$HOME/.profile"; do
    remove_line_if_present "$config_path" "$source_line"
done

# A leftover reference means an install used a different XDG_DATA_HOME than the
# current environment; the exact-match removal above can't see it.
for config_path in \
    "$HOME/.zshrc" \
    "$HOME/.bashrc" \
    "$HOME/.bash_profile" \
    "$HOME/.bash_login" \
    "$HOME/.profile"; do
    if [ -f "$config_path" ] && grep -q "grove/init\.sh" "$config_path"; then
        printf '%s\n' "note: $config_path still references a grove init.sh from a different install location; remove that line manually" >&2
    fi
done

printf '%s\n' "grove uninstalled; reload your shell (or open a new session) so the wrapper is no longer defined"
