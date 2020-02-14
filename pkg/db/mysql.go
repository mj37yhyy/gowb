package db

import (
	"context"
	"fmt"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

const NETWORK = "tcp"

func InitMysql(c context.Context) error {

	// 获取配置
	conf := c.Value(constant.ConfigKey).(config.Config)
	var err error
	dsn := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?%s",
		conf.Mysql.UserName,
		conf.Mysql.Password,
		NETWORK,
		conf.Mysql.Host,
		conf.Mysql.Port,
		conf.Mysql.Database,
		conf.Mysql.Params)
	log.Println("db connecting " + dsn)
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	DB.DB().SetConnMaxLifetime(
		utils.If(conf.Mysql.ConnMaxLifetime <= 0, 100*time.Second, conf.Mysql.ConnMaxLifetime*time.Second).(time.Duration),
	)
	DB.DB().SetMaxOpenConns(
		utils.If(conf.Mysql.MaxOpenConns <= 0, 100, conf.Mysql.MaxOpenConns).(int),
	)
	DB.DB().SetMaxIdleConns(
		utils.If(conf.Mysql.MaxIdleConns <= 0, 16, conf.Mysql.MaxOpenConns).(int),
	)
	DB.SingularTable(true)
	DB.LogMode(true)
	return nil
}
