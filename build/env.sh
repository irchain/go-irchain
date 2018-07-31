#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
root="$PWD"
workspace="$PWD/build/_workspace"
ircdir="$workspace/src/github.com/irchain"
if [ ! -L "$ircdir/go-irchain" ]; then
    mkdir -p "$ircdir"
    cd "$ircdir"
    ln -s "$root" go-irchain
    cd "$root"
fi

# Set up the environment to use the workspace.
export GOPATH="$workspace"

# Run the command inside the workspace.
cd "$ircdir/go-irchain"

# Launch the arguments with the configured environment.
exec "$@"
