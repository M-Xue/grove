function grove
	set -l output (command grove $argv)
	set -l status $status
	if test $status -ne 0
		return $status
	end
	if test -n "$output"
		cd "$output"; or return 1
	end
end
