package main

import (
	logger "github.com/hsyan2008/go-logger"
	"github.com/hsyan2008/gossh/config"
	"github.com/hsyan2008/gossh/controllers"
	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/pac"
	hfwsignal "github.com/hsyan2008/hfw2/signal"
	"github.com/hsyan2008/hfw2/ssh"
)

func main() {
	logger.Info("LoadConfig")
	err := config.LoadConfig()
	if err != nil {
		logger.Warn(err)
		return
	}
	logger.Warn(config.Config)

	logger.Info("LoadPac")
	err = pac.LoadDefault()
	if err != nil {
		logger.Warn(err)
		return
	}

	signalContext := hfwsignal.GetSignalContext()

	logger.Info("create LocalForward")
	for key, val := range config.Config.LocalForward {
		signalContext.WgAdd()
		go func(key string, val config.ForwardServer) {
			defer signalContext.WgDone()
			for _, v := range val.Inner {
				lf, err := ssh.NewLocalForward(val.SSHConfig, v)
				if err != nil {
					logger.Warn(val.SSHConfig, err)
					return
				}
				defer lf.Close()
			}
			for _, val2 := range config.Config.LocalForward[key].Indirect {
				lf, err := ssh.NewLocalForward(val.SSHConfig, nil)
				if err != nil {
					logger.Warn(val.SSHConfig, err)
					return
				}
				defer lf.Close()
				for _, v := range val2.Inner {
					err = lf.Dial(val2.SSHConfig, v)
					if err != nil {
						logger.Warn(val2, err)
						return
					}
				}
			}
			<-signalContext.Ctx.Done()
		}(key, val)
	}
	logger.Info("create Proxy")
	for _, val := range config.Config.Proxy {
		customPac(val.DomainPac)
		for _, v := range val.Inner {
			signalContext.WgAdd()
			go func(val config.ProxyServer, v *ssh.ProxyIni) {
				defer signalContext.WgDone()
				p, err := ssh.NewProxy(val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					return
				}
				defer p.Close()
				<-signalContext.Ctx.Done()
			}(val, v)
		}
	}

	signalContext.WgWait()
	logger.Info("Shutdown")

	hfw.Handler("/pac", &controllers.Pac{})
	hfw.Config.Server.Address = ":44444"
	hfw.Run()
}

func customPac(domainPac config.DomainPac) {
	logger.Infof("%#v", domainPac)
	for _, v := range domainPac.Deny {
		logger.Warn("pac", v, false)
		pac.Add(v, false)
	}
	for _, v := range domainPac.Allow {
		logger.Warn("pac", v, true)
		pac.Add(v, true)
	}
}
