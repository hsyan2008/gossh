package main

import (
	"errors"
	"path/filepath"
	"strings"

	hfw "github.com/hsyan2008/hfw2"
	"github.com/hsyan2008/hfw2/ssh"
)

var Config tomlConfig

func LoadConfig() (err error) {
	tomlfile := filepath.Join(hfw.APPPATH, "config.toml")
	err = hfw.TomlLoad(tomlfile, &Config)
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
	ssh.SSHConfig
	Inner map[string]ssh.ForwardIni
}

type ProxyServer struct {
	ssh.SSHConfig
	Inner map[string]ssh.ProxyIni
}
