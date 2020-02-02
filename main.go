package main

import (
	"os"
	"time"

	logger "github.com/hsyan2008/go-logger"
	"github.com/hsyan2008/gossh/config"
	"github.com/hsyan2008/hfw"
	"github.com/hsyan2008/hfw/pac"
	hfwsignal "github.com/hsyan2008/hfw/signal"
	"github.com/hsyan2008/hfw/ssh"
)

func main() {
	logger.Info("LoadConfig")
	err := config.LoadConfig()
	if err != nil {
		logger.Warn("LoadConfig:", err)
		return
	}
	logger.Warn(config.Config)

	signalContext := hfwsignal.GetSignalContext()

	logger.Info("create LocalForward")
	for key, val := range config.Config.Forward {
		signalContext.WgAdd()
		go func(key string, val config.ForwardServer) {
			defer signalContext.WgDone()
			time.Sleep(val.Delay * time.Second)
			for _, v := range val.Inner {
				lf, err := ssh.NewForward(hfw.NewHTTPContext(), val.Type, val.SSHConfig, v)
				if err != nil {
					logger.Warn(val.SSHConfig, err)
					os.Exit(2)
					return
				}
				defer lf.Close()
			}
			for _, val2 := range config.Config.Forward[key].Indirect {
				lf, err := ssh.NewForward(hfw.NewHTTPContext(), val.Type, val.SSHConfig, nil)
				if err != nil {
					logger.Warn(val.SSHConfig, err)
					os.Exit(2)
					return
				}
				defer lf.Close()
				for _, v := range val2.Inner {
					err = lf.Dial(val2.SSHConfig, v)
					if err != nil {
						logger.Warn(val2, err)
						os.Exit(2)
						return
					}
				}
			}
			<-signalContext.Ctx.Done()
		}(key, val)
	}
	logger.Info("create Proxy")
	for _, val := range config.Config.Proxy {
		for _, v := range val.Inner {
			if v.IsPac {
				logger.Info("LoadPac")
				err = pac.LoadDefault()
				if err != nil {
					logger.Warn("LoadPac:", err)
					return
				}
				break
			}
		}
		customPac(val.DomainPac)
		signalContext.WgAdd()
		go func(val config.ProxyServer) {
			defer signalContext.WgDone()
			time.Sleep(val.Delay * time.Second)
			for _, v := range val.Inner {
				p, err := ssh.NewProxy(hfw.NewHTTPContext(), val.SSHConfig, v)
				if err != nil {
					logger.Warn(err)
					os.Exit(2)
					return
				}
				defer p.Close()
			}
			<-signalContext.Ctx.Done()
		}(val)
	}

	// go func() {
	// 	hfw.Config.Server.Address = ":44444"
	// 	hfw.Config.Route.DefaultController = "index"
	// 	hfw.Config.Route.DefaultAction = "index"
	// 	hfw.Handler("/pac", &controllers.Pac{})
	// }()

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
