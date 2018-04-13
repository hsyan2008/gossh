package main

import (
	"github.com/hsyan2008/go-logger/logger"
	hfw "github.com/hsyan2008/hfw2"
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
	for k, v := range domain {
		pac.Add(k, v)
	}

	logger.Info("create LocalForward")
	for _, val := range Config.LocalForward {
		for _, v := range val.Inner {
			lf, err := ssh.NewLocalForward(val.SSHConfig, v)
			if err != nil {
				logger.Warn(err)
				return
			}
			defer lf.Close()
		}
	}
	logger.Info("create Proxy")
	for _, val := range Config.Proxy {
		for _, v := range val.Inner {
			p, err := ssh.NewProxy(val.SSHConfig, v)
			if err != nil {
				logger.Warn(err)
				return
			}
			defer p.Close()
		}
	}

	_ = hfw.Run()
}

var domain = map[string]bool{
	//黑名单
	"googlevideo.com": false,
	"github.com":      false,
	//白名单
	"goanimate.com":    true,
	"shutterstock.com": true,
	"google.cn":        true,
	"google.com.hk":    true,
}
