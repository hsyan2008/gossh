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

	//检查所有local里是否有相同的bind，检查每个remote里的inner是否有相同的bind
	tmp := make(map[string]bool)
	for key, val := range Config.LocalForward {
		for k, v := range val.Inner {
			if !strings.Contains(v.Bind, ":") {
				v.Bind = ":" + v.Bind
				Config.LocalForward[key].Inner[k] = v
			}
			if _, ok := tmp[v.Bind]; ok {
				return errors.New("duplicate LocalForward bind")
			}
			tmp[v.Bind] = true
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

	//检查每个remote里的inner是否有相同的bind
	for key, val := range Config.RemoteForward {
		tmp = make(map[string]bool)
		for k, v := range val.Inner {
			if !strings.Contains(v.Bind, ":") {
				v.Bind = ":" + v.Bind
				Config.RemoteForward[key].Inner[k] = v
			}
			if _, ok := tmp[v.Bind]; ok {
				return errors.New("duplicate RemoteForward bind")
			}
			tmp[v.Bind] = true
		}
	}

	return
}

type tomlConfig struct {
	LocalForward  map[string]ForwardServer
	RemoteForward map[string]ForwardServer
	Proxy         map[string]ProxyServer
}

type ForwardServer struct {
	// Forward
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
