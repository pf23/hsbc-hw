mkdir bin

# run unit tests
cd cmd/ && go test -v .
cd ../model/ && go test -v .
cd ../serving/ && go test -v .

cd ../cmd && go build -o ../bin/server
