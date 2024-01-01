package cli

import (
	"fmt"
	"strings"
)

// Got without changes from here https://github.com/hashicorp/terraform/commit/efe78b2910c5b1f2292f0a16d990b2ad92352feb#diff-2873f79a86c0d8b3335cd7731b0ecf7dd4301eb19a82ef7a1cba7589b5252261R469
// because it's requited to mimic terraform cli behavior if -chdir arg.
// TODO: It's required to be reworked
func extractChdirOption(args []string) (string, []string, error) {
	if len(args) == 0 {
		return "", args, nil
	}

	const argName = "-chdir"
	const argPrefix = argName + "="
	var argValue string
	var argPos int

	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			// Because the chdir option is a subcommand-agnostic one, we require
			// it to appear before any subcommand argument, so if we find a
			// non-option before we find -chdir then we are finished.
			break
		}
		if arg == argName || arg == argPrefix {
			return "", args, fmt.Errorf("must include an equals sign followed by a directory path, like -chdir=example")
		}
		if strings.HasPrefix(arg, argPrefix) {
			argPos = i
			argValue = arg[len(argPrefix):]
		}
	}

	// When we fall out here, we'll have populated argValue with a non-empty
	// string if the -chdir=... option was present and valid, or left it
	// empty if it wasn't present.
	if argValue == "" {
		return "", args, nil
	}

	// If we did find the option then we'll need to produce a new args that
	// doesn't include it anymore.
	if argPos == 0 {
		// Easy case: we can just slice off the front
		return argValue, args[1:], nil
	}
	// Otherwise we need to construct a new array and copy to it.
	newArgs := make([]string, len(args)-1)
	copy(newArgs, args[:argPos])
	copy(newArgs[argPos:], args[argPos+1:])
	return argValue, newArgs, nil
}
