export GOPATH=$GOPATH:$(pwd)
# if [ -f ./core ]; then
#   rm ./core || true
# fi
go build -ldflags "-s" -o ./shodan ./*.go;
echo "Build completed"
