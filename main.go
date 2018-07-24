package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hsyan2008/go-logger/logger"
	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/common"
	"github.com/hsyan2008/hfw2/pac"
	"github.com/hsyan2008/hfw2/ssh"
)

var domainFile = filepath.Join(hfw.APPPATH, "domain.txt")

func main() {
	logger.Info("LoadConfig")
	err := LoadConfig()
	if err != nil {
		logger.Warn(err)
		return
	}

	logger.Info("LoadPac")
	err = pac.LoadDefault()
	if err != nil {
		logger.Warn(err)
		return
	}
	customPac()
	go listenSignal()

	logger.Info("create LocalForward")
	for _, val := range Config.LocalForward {
		for _, v := range val.Inner {
			hfw.Ctx.WgAdd()
			go func(val ForwardServer, v ssh.ForwardIni) {
				defer hfw.Ctx.WgDone()
				lf, err := ssh.NewLocalForward(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-hfw.Ctx.Shutdown
				defer lf.Close()
			}(val, v)
		}
	}
	logger.Info("create Proxy")
	for _, val := range Config.Proxy {
		for _, v := range val.Inner {
			hfw.Ctx.WgAdd()
			go func(val ProxyServer, v ssh.ProxyIni) {
				defer hfw.Ctx.WgDone()
				p, err := ssh.NewProxy(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-hfw.Ctx.Shutdown
				defer p.Close()
			}(val, v)
		}
	}

	hfw.Ctx.WgWait()
	logger.Info("Shutdown")
}

func getDomain(file string) (domain map[string]bool) {
	domain = make(map[string]bool)
	if !common.IsExist(file) {
		return
	}
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	lines := strings.Split(string(fileContent), "\n")
	for _, line := range lines {
		s := strings.Split(strings.ToLower(strings.TrimSpace(line)), ":")
		if len(s) == 2 {
			domain[s[0]] = s[1] == "true"
		}
	}

	return
}

func customPac() {
	logger.Info(domainFile)
	domain := getDomain(domainFile)
	for k, v := range domain {
		logger.Warn("pac", k, v)
		pac.Add(k, v)
	}
}

func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	for {
		<-c
		logger.Info("LoadPac")
		err := pac.Reset()
		if err != nil {
			logger.Warn(err)
		} else {
			customPac()
		}
	}
}
