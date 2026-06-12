grove() {
	local output
	output="$(command grove "$@")"
	local status=$?
	if [ $status -ne 0 ]; then
		return $status
	fi
	if [ -n "$output" ]; then
		cd "$output" || return 1
	fi
}
