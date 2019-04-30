GIT_REV="$(git rev-parse --short=7 HEAD)"
IMAGE_BUILD="${IMAGE_BUILD:-local/uhc-check:${GIT_REV}}"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function build_image() {
	docker build -t "${IMAGE_BUILD}" ${DIR} -f ${DIR}/Dockerfile
}
