package config

import (
	"errors"
	"strings"
	"time"

	"github.com/hsyan2008/hfw"
	"github.com/hsyan2008/hfw/configs"
	"github.com/hsyan2008/hfw/ssh"
)

var Config tomlConfig

func LoadConfig() (err error) {
	//初始化log写入文件
	_ = hfw.Init()

	err = configs.Load(&Config)
	if err != nil {
		return err
	}

	if len(Config.Forward) == 0 && len(Config.Proxy) == 0 {
		return errors.New("config is nil")
	}

	//检查所有local里是否有相同的bind，检查每个remote里的inner是否有相同的bind
	//remote的bind是在远端，不需要加到本地判断
	tmp := make(map[string]bool)
	tmpRemote := make(map[string]bool)
	for key, val := range Config.Forward {
		for k, v := range val.Inner {
			if !strings.Contains(v.Bind, ":") {
				v.Bind = ":" + v.Bind
				Config.Forward[key].Inner[k] = v
			}
			if val.Type == ssh.LOCAL {
				if _, ok := tmp[v.Bind]; ok {
					return errors.New("duplicate LocalForward bind")
				}
				tmp[v.Bind] = true
			} else if val.Type == ssh.REMOTE {
				if _, ok := tmpRemote[v.Bind]; ok {
					return errors.New("duplicate RemoteForward bind")
				}
				tmpRemote[v.Bind] = true
			} else {
				return errors.New("error forward type")
			}
		}
	}

	//检查所有proxy里的inner是否有相同的bind或和local相同
	for key, val := range Config.Proxy {
		for k, v := range val.Inner {
			if !strings.Contains(v.Bind, ":") {
				v.Bind = ":" + v.Bind
				Config.Proxy[key].Inner[k] = v
			}
			if _, ok := tmp[v.Bind]; ok {
				return errors.New("duplicate Proxy bind")
			}
			tmp[v.Bind] = true
		}
	}

	return
}

type tomlConfig struct {
	Forward map[string]ForwardServer
	Proxy   map[string]ProxyServer
}

type ForwardServer struct {
	Type ssh.ForwardType

	ssh.SSHConfig
	Delay time.Duration
	Inner map[string]*ssh.ForwardIni

	//二次登陆ssh
	Indirect map[string]ForwardIndirect
}

type ForwardIndirect struct {
	ssh.SSHConfig
	Inner map[string]*ssh.ForwardIni
}

type ProxyServer struct {
	ssh.SSHConfig
	Delay     time.Duration
	Inner     map[string]*ssh.ProxyIni
	DomainPac DomainPac
}

//is_pac设置下才有效
type DomainPac struct {
	//不允许访问
	Deny []string
	//允许访问
	Allow []string
}
