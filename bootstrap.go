package gowb

import (
	"context"
	"errors"
	"fmt"
	"github.com/mj37yhyy/gowb/pkg/config"
	gowbLog "github.com/mj37yhyy/gowb/pkg/log"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"github.com/mj37yhyy/gowb/pkg/web"
	"os"
	"runtime"
)

const logo = `
 _____   _____   _          __  _____  
/  ___| /  _  \ | |        / / |  _  \ 
| |     | | | | | |  __   / /  | |_| | 
| |  _  | | | | | | /  | / /   |  _  { 
| |_| | | |_| | | |/   |/ /    | |_| | 
\_____/ \_____/ |___/|___/     |_____/ 
`

type Gowb struct {
	ConfigName string
	ConfigType string
	Routers    []web.Router
}

func Bootstrap(g Gowb) (err error) {
	fmt.Println(logo)
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	//if !reflect.DeepEqual(g, Gowb{}) {
	if g.ConfigName != "" && g.ConfigType != "" {
		cu, err := utils.NewConfig(g.ConfigName, g.ConfigType)
		if err != nil {
			return err
		}
		// 解析并处理yaml
		var _config config.Config
		if err := cu.Unmarshal(&_config); err != nil {
			return err
		} else {
			c := context.WithValue(context.Background(), "routers", g.Routers)
			c = context.WithValue(c, "config", _config)
			gowbLog.InitLogger(c)
			web.Bootstrap(c)
		}
	} else {
		return errors.New("ConfigName and ConfigType is empty!")
	}
	return nil
}
