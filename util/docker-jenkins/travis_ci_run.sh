: '
An end-to-end test of both local configurations (Chrome, FF),
and remote (Edge, Safari).

This script must pass in order for an image to be pushed
to production for test runners to pull.
'
DOCKER_DIR=$(dirname "$0")
WPTD_PATH=${WPTD_PATH:-"${DOCKER_DIR}/../.."}
BASE_IMAGE_NAME="wptd-base"
JENKINS_IMAGE_NAME="wptd-testrun-jenkins"
JENKINS_DOCKERFILE="${WPTD_PATH}/Dockerfile.jenkins"

docker build -t "${BASE_IMAGE_NAME}" "${WPTD_PATH}"
docker build -t "${JENKINS_IMAGE_NAME}" -f "${JENKINS_DOCKERFILE}" "${WPTD_PATH}"

docker run \
    -p 4445:4445 \
    --entrypoint "/bin/bash" "${JENKINS_IMAGE_NAME}" \
    /home/jenkins/wpt.fyi/util/docker-jenkins/inner/travis_ci_run.sh
