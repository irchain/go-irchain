#!/bin/sh\

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
root="$PWD"
workspace="$PWD/build/_workspace"
hucdir="$workspace/src/github.com/happyuc-project"
if [ ! -L "$hucdir/happyuc-go" ]; then
    mkdir -p "$hucdir"
    cd "$hucdir"
    ln -s "$root" happyuc-go
    cd "$root"
fi

# Set up the environment to use the workspace.
export GOPATH="$workspace"

# Run the command inside the workspace.
cd "$hucdir/happyuc-go"

# Launch the arguments with the configured environment.
exec "$@"
