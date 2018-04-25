package main

import (
	"io/ioutil"
	"strings"

	"github.com/hsyan2008/go-logger/logger"
	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/common"
	"github.com/hsyan2008/hfw2/pac"
	"github.com/hsyan2008/hfw2/ssh"
)

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
	domain := getDomain("domain.txt")
	for k, v := range domain {
		pac.Add(k, v)
	}

	logger.Info("create LocalForward")
	for _, val := range Config.LocalForward {
		for _, v := range val.Inner {
			go func(val ForwardServer, v ssh.ForwardIni) {
				hfw.Wg.Add(1)
				defer hfw.Wg.Done()
				lf, err := ssh.NewLocalForward(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-hfw.Shutdown
				defer lf.Close()
			}(val, v)
		}
	}
	logger.Info("create Proxy")
	for _, val := range Config.Proxy {
		for _, v := range val.Inner {
			go func(val ProxyServer, v ssh.ProxyIni) {
				hfw.Wg.Add(1)
				defer hfw.Wg.Done()
				p, err := ssh.NewProxy(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				<-hfw.Shutdown
				defer p.Close()
			}(val, v)
		}
	}

	hfw.Wg.Wait()
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
