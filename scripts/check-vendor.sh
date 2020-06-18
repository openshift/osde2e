#!/bin/bash
#
# Check if the vendor directory needs an update.
#

ROOT_DIR="$(dirname "$0")/.."
TEMP_COMMIT="false"
ORIG_MODULES="$(mktemp)"

reset() {
	if [ "$TEMP_COMMIT" = "true" ]
	then
		rm "$ORIG_MODULES"
		(cd "$ROOT_DIR" && git reset --hard > /dev/null 2>&1)
		(cd "$ROOT_DIR" && git clean -fd > /dev/null 2>&1)
		git reset HEAD~1 > /dev/null 2>&1
	fi
}

trap reset EXIT TERM

MODULES_FILE="$ROOT_DIR/vendor/modules.txt"
ORIG_MODULES="$(mktemp)"

cp "$MODULES_FILE" "$ORIG_MODULES"

# Create a temporary commit
git add "$ROOT_DIR" 2>&1
if ! git commit -q --allow-empty --author="SD CI/CD Automation <sd-cicd@redhat.com>" -m "Temporary commit for vendor check." 2>&1
then
	echo "Unable to make a temporary commit."
	exit 1
fi
TEMP_COMMIT="true"

# Update the modules
(cd "$ROOT_DIR" && go mod vendor)

if ! cmp -s "$ORIG_MODULES" "$MODULES_FILE"
then
	echo "The vendor directory needs an update."
	exit 1
fi


