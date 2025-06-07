package bootstrap

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/haomiao000/raftchain/library/ext"
	"github.com/haomiao000/raftchain/library/resource"
	"github.com/haomiao000/raftchain/library/query"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func InitBootStrap(modelInit bool) {
	if modelInit {
		initGormModel()
	}
	initGorm()

}

func initGormModel() {
	loadConfig("mysql.yaml", &ext.MySQL)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		ext.MySQL.User, ext.MySQL.Password, ext.MySQL.Host, ext.MySQL.Port,
		ext.MySQL.DBName, ext.MySQL.Charset, ext.MySQL.ParseTime, ext.MySQL.Loc)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 初始化 generator
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./library/query", // 生成的代码输出目录
		ModelPkgPath: "./library/model", // 生成 model 的目录（可选）
		Mode:         gen.WithoutContext|gen.WithDefaultQuery|gen.WithQueryInterface,
	})

	g.UseDB(db)

	g.ApplyBasic(g.GenerateAllTable()...)
	g.Execute()
}

func initGorm() {
	loadConfig("mysql.yaml", &ext.MySQL)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		ext.MySQL.User, ext.MySQL.Password, ext.MySQL.Host, ext.MySQL.Port,
		ext.MySQL.DBName, ext.MySQL.Charset, ext.MySQL.ParseTime, ext.MySQL.Loc)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("MySQL 初始化失败: %v", err)
	}
	resource.GormServe = db
	query.SetDefault(db)
}

func loadConfig(fileName string, configPtr interface{}) {
	v := viper.New()
	v.SetConfigName(fileName[:len(fileName)-len(filepath.Ext(fileName))])
	v.SetConfigType("yaml")
	v.AddConfigPath("../conf")

	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("读取配置文件 %s 失败: %v", fileName, err)
	}

	err = v.Unmarshal(configPtr)
	if err != nil {
		log.Fatalf("解析配置文件 %s 失败: %v", fileName, err)
	}
}
