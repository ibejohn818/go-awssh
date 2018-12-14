
clean:
	rm -rf ./dist

build: clean
	mkdir ./dist
	go build -o ./dist/awssh awssh/awssh.go
	chmod +x ./dist/awssh

install:
	cp ./dist/awssh ~/bin/awssh
