#!/bin/sh
set -e

# This script creates a user with the same UID/GID as the host user
# to avoid file permission issues with mounted volumes. It also handles
# cases where the UID/GID might already exist in the container.

# Default to UID/GID of 1000 if not provided
USER_ID=${HOST_UID:-1000}
GROUP_ID=${HOST_GID:-1000}

echo ">>> Starting with UID: $USER_ID, GID: $GROUP_ID"

# --- FIX #1: Handle existing user/group ---
# Create a group and user, but don't fail if they already exist.
# Check if the group exists, if not, create it.
if ! getent group "$GROUP_ID" >/dev/null; then
    addgroup -g "$GROUP_ID" -S user
fi

# Check if the user exists, if not, create it.
if ! getent passwd "$USER_ID" >/dev/null; then
    adduser -u "$USER_ID" -S user -G user
fi

# --- FIX #2: Take ownership of the working directory ---
# This allows the new non-root user to create files (like Air's tmp dir).
# The `"."` refers to the current working directory set in the Dockerfile.
chown -R "$USER_ID:$GROUP_ID" .

# "Step down" from root and execute the main command passed to the container.
# The `"$@"` passes along all arguments from the Docker CMD or `command:`.
exec su-exec "$USER_ID" "$@"