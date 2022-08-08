WORDIR=${PWD}

mkdir ${WORDIR}/bin

# run unit tests
cd ${WORDIR}/model/ && go test -v .
cd ${WORDIR}/serving/ && go test -v .

# build binary
cd ${WORDIR}/cmd && go build -o ${WORDIR}/bin/server
