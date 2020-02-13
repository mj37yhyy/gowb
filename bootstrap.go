package gowb

import (
	"context"
	"errors"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"github.com/mj37yhyy/gowb/pkg/web"
	"os"
	"runtime"
)

type Gowb struct {
	ConfigName string
	ConfigType string
	Routers    []web.Router
}

func init() {
}

func Bootstrap(g Gowb) (err error) {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	//if !reflect.DeepEqual(g, Gowb{}) {
	if g.ConfigName != "" && g.ConfigType != "" {
		cu, err := utils.ConfigUtils{}.New(g.ConfigName, g.ConfigType)
		if err != nil {
			return err
		}
		c := context.WithValue(context.Background(), "routers", g.Routers)
		c = context.WithValue(c, "config", cu)
		web.Bootstrap(c)
	} else {
		return errors.New("ConfigName and ConfigType is empty!")
	}
	return nil
}
