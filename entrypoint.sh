#!/bin/sh
# Set strict error checking
set -e

# Take ownership of the notes directory.
# This is necessary because the volume might be mounted with root ownership.
chown -R appuser:appgroup /app/notes

# Execute the main command (passed as arguments to this script) as the appuser
exec su-exec appuser "$@"