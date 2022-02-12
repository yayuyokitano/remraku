set -e

docker build -t us-central1-docker.pkg.dev/rem-970606/remraku/main .
docker push us-central1-docker.pkg.dev/rem-970606/remraku/main