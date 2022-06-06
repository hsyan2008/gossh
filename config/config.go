package config

import (
	"errors"
	"fmt"
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
			v.Bind, err = completeAddr(v.Bind)
			if err != nil {
				return
			}
			Config.Forward[key].Inner[k] = v
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
			v.Bind, err = completeAddr(v.Bind)
			if err != nil {
				return
			}
			Config.Proxy[key].Inner[k] = v
			if _, ok := tmp[v.Bind]; ok {
				return errors.New("duplicate Proxy bind")
			}
			tmp[v.Bind] = true
		}
	}

	return
}

//addr不处理，只处理bind，bind只能是本机，所以ip必须是127.0.0.1/0.0.0.0
func completeAddr(str string) (string, error) {
	if len(str) == 0 {
		return str, errors.New("err addr")
	}
	tmp := strings.Split(str, ":")
	if len(tmp) == 1 {
		return fmt.Sprintf("0.0.0.0:%s", tmp[0]), nil
	}

	if tmp[0] == "" {
		tmp[0] = "0.0.0.0"
	}

	return strings.Join(tmp, ":"), nil
}

type tomlConfig struct {
	Forward map[string]ForwardServer
	Proxy   map[string]ProxyServer
}

type ForwardServer struct {
	Type ssh.ForwardType

	Delay time.Duration

	ssh.ForwardConfig
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
