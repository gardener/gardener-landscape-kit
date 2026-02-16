#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

# TODO(LucaBernstein): Remove the drop-in replacement for `parallel` once the `format.sh` script in `g/g` introduced `MODE=sequential` support (again) with https://github.com/gardener/gardener/pull/14076.
# if $MODE var is set to parallel, then export the parallel function to be used in the format.sh script, otherwise it will be ignored and commands will be executed sequentially.
if [[ "$MODE" == "sequential" ]]; then
  # Custom drop-in replacement for GNU Parallel to avoid the dependency on GNU Parallel while still allowing for parallel execution of commands.
  # Necessary, as the used `format.sh` script is also used in Renovate's `make format` post-step, where GNU Parallel is not available.
  parallel() {
    echo "Called drop-in replacement for GNU Parallel with arguments: $*"
    local cmd_args=("$@") # Store all arguments passed to the parallel function
    local input_item

    # Read each line from standard input
    while IFS= read -r input_item; do
      local -a current_command=() # Array to build the command for the current item
      local arg

      # Construct the command, replacing {} with the input_item
      for arg in "${cmd_args[@]}"; do
        if [[ "$arg" == "{}" ]]; then
          current_command+=("$input_item")
        elif [[ "$arg" == "--will-cite" ]]; then
          # Ignore --will-cite as it's specific to GNU Parallel and not functional for a custom implementation
          continue
        else
          current_command+=("$arg")
        fi
      done

      # Execute the constructed command
      "${current_command[@]}" || {
        # Basic error handling: print a message if the command fails
        echo "Error executing command for input: '$input_item': ${current_command[*]}" >&2
        # You might want to 'exit 1' here if failure should stop the script
      }
    done
  }

  export -f parallel
fi

bash $GARDENER_HACK_DIR/format.sh $@
