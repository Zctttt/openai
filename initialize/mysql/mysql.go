package mysql

import (
	"database/sql"
	"fmt"
	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
	log "github.com/sirupsen/logrus"
	"go-chatgpt-api/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

var marketDB *gorm.DB
var chatDB *gorm.DB

// Init 注册mysql
func Init() {
	log.Println("connect mysql start~")
	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		beego.Error("mysql register driver error:", err)
	}

	username := config.GetChatMysqlConfig().Username
	password := config.GetChatMysqlConfig().Password
	host := config.GetChatMysqlConfig().Host
	port := config.GetChatMysqlConfig().Port
	database := config.GetChatMysqlConfig().Database

	createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci", database)
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(createDB)
	if err != nil {
		panic(err)
	}

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", username, password, host, port, database)

	err = orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		beego.Error("mysql register database error:", err)
	}
	//每次自动创建表，会清除表数据
	//orm.RunSyncdb("default", true, true)
	log.Println("connect mysql success~")

	conn_m := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.GetMarketMysqlConfig().Username, config.GetMarketMysqlConfig().Password, config.GetMarketMysqlConfig().Host, config.GetMarketMysqlConfig().Port, config.GetMarketMysqlConfig().Database)
	marketDB, err = gorm.Open(mysql.Open(conn_m), &gorm.Config{ //建立连接时指定打印info级别的sql
		Logger: logger.Default.LogMode(logger.Info), //配置日志级别，打印出所有的sql
	})

	conn_c := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.GetChatMysqlConfig().Username, config.GetChatMysqlConfig().Password, config.GetChatMysqlConfig().Host, config.GetChatMysqlConfig().Port, config.GetChatMysqlConfig().Database)
	chatDB, err = gorm.Open(mysql.Open(conn_c), &gorm.Config{ //建立连接时指定打印info级别的sql
		Logger: logger.Default.LogMode(logger.Info), //配置日志级别，打印出所有的sql
	})
	if err != nil {
		panic(err)
	}

}

func GetMarketDb() *gorm.DB {
	return marketDB
}

func GetChatDb() *gorm.DB {
	return chatDB
}
