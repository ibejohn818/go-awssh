
clean:
	rm -rf ./dist

build_linux:
	GOOS=linux GOARCH=amd64 go build -o ./dist/awssh awssh/awssh.go
	chmod +x ./dist/awssh
	zip ./dist/awssh-linux-amd64.zip ./dist/awssh
	rm -rf ./dist/awssh

build_osx:
	GOOS=darwin GOARCH=amd64 go build -o ./dist/awssh awssh/awssh.go
	chmod +x ./dist/awssh
	zip ./dist/awssh-darwin-amd64.zip ./dist/awssh
	rm -rf ./dist/awssh


build: clean build_linux build_osx
