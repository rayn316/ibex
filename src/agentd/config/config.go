package config

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/koding/multiconfig"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/runner"

	"github.com/ulricqin/ibex/src/pkg/httpx"
)

var (
	C    = new(Config)
	once sync.Once
)

func MustLoad(fpaths ...string) {
	once.Do(func() {
		loaders := []multiconfig.Loader{
			&multiconfig.TagLoader{},
			&multiconfig.EnvironmentLoader{},
		}

		for _, fpath := range fpaths {
			handled := false

			if strings.HasSuffix(fpath, "toml") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "conf") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "json") {
				loaders = append(loaders, &multiconfig.JSONLoader{Path: fpath})
				handled = true
			}
			if strings.HasSuffix(fpath, "yaml") {
				loaders = append(loaders, &multiconfig.YAMLLoader{Path: fpath})
				handled = true
			}

			if !handled {
				fmt.Println("config file invalid, valid file exts: .conf,.yaml,.toml,.json")
				os.Exit(1)
			}
		}

		m := multiconfig.DefaultLoader{
			Loader:    multiconfig.MultiLoader(loaders...),
			Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
		}

		if C.Heartbeat.Host == "$ip" {
			C.Heartbeat.Endpoint = fmt.Sprint(GetOutboundIP())
			if C.Heartbeat.Endpoint == "" {
				fmt.Println("ip auto got is blank")
				os.Exit(1)
			}
			fmt.Println("host.ip:", C.Heartbeat.Endpoint)
		}

		if C.MetaDir == "" {
			C.MetaDir = "./meta"
		}

		C.MetaDir = path.Join(runner.Cwd, C.MetaDir)
		file.EnsureDir(C.MetaDir)

		m.MustLoad(C)
	})
}

type Config struct {
	RunMode   string
	MetaDir   string
	Heartbeat Heartbeat
	HTTP      httpx.Config
}

type Heartbeat struct {
	Interval int64
	Servers  []string
	Host     string
	Endpoint string
}

func (c *Config) IsDebugMode() bool {
	return c.RunMode == "debug"
}

func (c *Config) GetHost() (string, error) {
	if c.Heartbeat.Host == "$ip" {
		return c.Heartbeat.Endpoint, nil
	}

	if c.Heartbeat.Host == "$hostname" {
		return os.Hostname()
	}

	return c.Heartbeat.Host, nil
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("auto get outbound ip fail:", err)
		os.Exit(1)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}