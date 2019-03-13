package main

import (
	"gossh/config"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	logger "github.com/hsyan2008/go-logger"
	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/common"
	"github.com/hsyan2008/hfw2/pac"
	hfwsignal "github.com/hsyan2008/hfw2/signal"
	"github.com/hsyan2008/hfw2/ssh"
)

var domainFile = filepath.Join(hfw.APPPATH, "domain.txt")

func main() {
	logger.Info("LoadConfig")
	err := config.LoadConfig()
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

	signalContext := hfwsignal.GetSignalContext()

	logger.Info("create LocalForward")
	for key, val := range config.Config.LocalForward {
		for _, v := range val.Inner {
			signalContext.WgAdd()
			go func(val config.ForwardServer, v ssh.ForwardIni) {
				defer signalContext.WgDone()
				lf, err := ssh.NewLocalForward(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-signalContext.Ctx.Done()
				defer lf.Close()
			}(val, v)
		}
		for _, val2 := range config.Config.LocalForward[key].Indirect {
			for _, v := range val2.Inner {
				signalContext.WgAdd()
				go func(val config.ForwardServer, v ssh.ForwardIni) {
					defer signalContext.WgDone()
					lf, err := ssh.NewLocalForward(val.SSHConfig, ssh.ForwardIni{})
					if err != nil {
						logger.Warn(err)
						return
					}
					err = lf.Dial(val2.SSHConfig, v)
					if err != nil {
						logger.Warn(err)
						return
					}
					<-signalContext.Ctx.Done()
					defer lf.Close()
				}(val, v)
			}
		}
	}
	logger.Info("create Proxy")
	for _, val := range config.Config.Proxy {
		for _, v := range val.Inner {
			signalContext.WgAdd()
			go func(val config.ProxyServer, v ssh.ProxyIni) {
				defer signalContext.WgDone()
				p, err := ssh.NewProxy(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-signalContext.Ctx.Done()
				defer p.Close()
			}(val, v)
		}
	}

	signalContext.WgWait()
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

//reload pac
func listenSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGFPE)
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
