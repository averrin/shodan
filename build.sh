export GOPATH=$GOPATH:$(pwd)
# if [ -f ./core ]; then
#   rm ./core || true
# fi
go build -ldflags "-s" -o ./shodan ./main.go;
echo "Build completed"
