#https://linux.cn/article-10001-1.html
.PHONY: all build upx install release clean
all: build upx install
ssh_forward: build upx install_ssh_forward
build:
	@echo "go build"
	CGO_ENABLED=0 go build -a -ldflags '-s -w'
upx:
	@echo "upx -9"
	upx -9 gossh
install:
	@echo "install to /home/saxon/ssh/"
	mv gossh /home/saxon/ssh/
install_ssh_forward:
	@echo "install to /home/saxon/ssh/"
	mv gossh /home/saxon/ssh/ssh_forward
release:
	@echo "rm release/*"
	rm -rf release/*
	@echo "xgo windows"
	xgo --targets windows/* -ldflags -H=windowsgui .
	@echo "xgo linux/darwin"
	xgo --targets linux/386,linux/amd64,darwin/* .
	@echo "mv to release/"
	mv gossh-* release
clean:
	@echo "Cleaning up..."
	rm -rf gossh
