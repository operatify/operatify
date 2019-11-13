DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
VERSION=0.1.0

docker build -t resource-operator-factory-dev-container:$VERSION -f $DIR/Dockerfile $DIR/..
