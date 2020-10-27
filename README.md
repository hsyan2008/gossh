# gossh
[goproxy](https://github.com/hsyan2008/goproxy)和[gotunnel](https://github.com/hsyan2008/gotunnel)的修改完善版

## 特性  

#### 代理  
* 服务器不需要部署任何服务，只要有ssh服务和账号即可
* 支持http
* 支持https
* 支持socket5的TCP
* 支持以上所有协议over在ssh上(可选)

#### 隧道  
* 支持多个环境(每个环境单个跳板机)
* 支持多个内部机器
* 支持批量创建

## 使用方法
* 安装[golang](https://www.golang.org/)
* 执行如下命令
    
        mkdir ~/go
        GOPATH=~/go
        go get -u github.com/gossh
        cd ~/go/src/github.com/gossh
        go build            #win下有cmd窗口，使用go build -ldflags -H=windowsgui
* 编辑config.toml和domain.txt
* 启动

        ./gossh   #win下是./gossh.exe
