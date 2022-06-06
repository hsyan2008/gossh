package main

import (
	"time"

	"github.com/hsyan2008/gossh/config"
	"github.com/hsyan2008/hfw"
	"github.com/hsyan2008/hfw/pac"
	"github.com/hsyan2008/hfw/signal"
	"github.com/hsyan2008/hfw/ssh"
)

func main() {
	signalContext := signal.GetSignalContext()

	signalContext.Info("LoadConfig")
	err := config.LoadConfig()
	if err != nil {
		signalContext.Warn("LoadConfig:", err)
		return
	}

	signalContext.Info("create LocalForward")
	for key, val := range config.Config.Forward {
		go func(key string, val config.ForwardServer) {
			time.Sleep(val.Delay * time.Second)
			ctx := hfw.NewHTTPContext()
			defer ctx.Cancel()
			ssh.NewForward(ctx, val.Type, val.ForwardConfig)
		}(key, val)
	}

	signalContext.Info("create Proxy")
	for _, val := range config.Config.Proxy {
		for _, v := range val.Inner {
			if v.IsPac {
				signalContext.Info("LoadPac")
				err = pac.LoadDefault()
				if err != nil {
					signalContext.Warn("LoadPac:", err)
					signalContext.Cancel()
					return
				}
				break
			}
		}
		customPac(val.DomainPac)
		time.Sleep(val.Delay * time.Second)
		for _, v := range val.Inner {
			go func(v *ssh.ProxyIni) {
				ctx := hfw.NewHTTPContext()
				defer ctx.Cancel()
				p, err := ssh.NewProxy(ctx, val.SSHConfig, v)
				if err != nil {
					signalContext.Warn(err)
					signalContext.Cancel()
					return
				}
				defer p.Close()
			}(v)
		}
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
	signal.GetSignalContext().Infof("%#v", domainPac)
	for _, v := range domainPac.Deny {
		signal.GetSignalContext().Warn("pac", v, false)
		pac.Add(v, false)
	}
	for _, v := range domainPac.Allow {
		signal.GetSignalContext().Warn("pac", v, true)
		pac.Add(v, true)
	}
}
