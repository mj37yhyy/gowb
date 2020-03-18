package gowb

import (
	"context"
	"errors"
	"fmt"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/db"
	gowbLog "github.com/mj37yhyy/gowb/pkg/log"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"github.com/mj37yhyy/gowb/pkg/web"
	"os"
	"runtime"
	"unsafe"
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
	ConfigName       string
	ConfigType       string
	Config           config.Config
	Routers          []web.Router
	AutoCreateTables []interface{}
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
			if err := doBootstrap(g, _config); err != nil {
				return err
			}
		}
	} else if unsafe.Sizeof(g.Config) > 0 {
		if err := doBootstrap(g, g.Config); err != nil {
			return err
		}
	} else {
		return errors.New("ConfigName and ConfigType is empty!")
	}
	return nil
}

func doBootstrap(g Gowb, _config config.Config) error {
	c := context.WithValue(context.Background(), "routers", g.Routers)
	c = context.WithValue(c, "config", _config)

	//初始化mysql
	err := initMysql(c, g)
	if err != nil {
		return err
	}

	//初始化日志
	err = gowbLog.InitLogger(c)
	if err != nil {
		return err
	}

	//初始化gin
	web.Bootstrap(c)
	return nil
}

func initMysql(c context.Context, g Gowb) error {
	err := db.InitMysql(c)
	if err != nil {
		return err
	}
	//建表
	for _, t := range g.AutoCreateTables {
		db.DB.AutoMigrate(t)
	}
	return nil
}
