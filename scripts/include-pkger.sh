#!/bin/bash
#
# A script to include assets needed for osde2e testing.
#


old=$(md5sum pkged.go| awk '{print $1 }')

pkger --include "$(pwd)"/assets

new=$(md5sum pkged.go| awk '{print $1 }')

if [ "$new" == "$old" ]; then
    echo "The file pkged.go has not changed from its previous state."
    exit 0
else
    echo "The file pkged.go has been updated from its previous state."
    exit 1
fi
