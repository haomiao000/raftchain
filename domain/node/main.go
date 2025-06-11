package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/haomiao000/raftchain/domain/node/consensus"
	"github.com/haomiao000/raftchain/domain/node/rpc/raft"
	"github.com/haomiao000/raftchain/library/resource"
	"github.com/haomiao000/raftchain/library/ext"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


// Config 用于解析 raft.yaml
type Config struct {
	Nodes map[int]string `yaml:"nodes"`
}

// initDB 直接在此处实现数据库的初始化逻辑
func initDB() {
	// 1. 加载数据库配置
	var mysqlConf ext.MySQLConfig
	v := viper.New()
	v.SetConfigName("mysql")
	v.SetConfigType("yaml")
	v.AddConfigPath("../../conf") // 相对于 main.go 的路径
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read mysql.yaml: %v", err)
	}
	if err := v.Unmarshal(&mysqlConf); err != nil {
		log.Fatalf("Failed to unmarshal mysql config: %v", err)
	}

	// 2. 构造 DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConf.User,
		mysqlConf.Password,
		mysqlConf.Host,
		mysqlConf.Port,
		mysqlConf.DBName,
	)

	// 3. 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 4. 将创建好的数据库实例赋值给全局变量
	resource.GormServe = db
	log.Println("[DB] Database connection for this node has been initialized successfully.")
}

func main() {
	nodeID := flag.Int("id", -1, "Node ID")
	flag.Parse()

	if *nodeID == -1 {
		log.Fatalf("Please provide a node ID using -id flag")
	}

	// ✅ 直接调用在 main.go 中新定义的 initDB 函数
	initDB()

	// 加载 Raft 节点配置
	configFile, err := os.ReadFile("../../conf/raft.yaml")
	if err != nil {
		log.Fatalf("Failed to read raft.yaml: %v", err)
	}
	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Failed to unmarshal raft.yaml: %v", err)
	}

	selfAddr, ok := config.Nodes[*nodeID]
	if !ok {
		log.Fatalf("Node ID %d not found in config file", *nodeID)
	}

	log.Printf("Initializing Node %d at %s...", *nodeID, selfAddr)

	consensusNode := consensus.NewConsensusNode(*nodeID, config.Nodes)

	lis, err := net.Listen("tcp", selfAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", selfAddr, err)
	}
	s := grpc.NewServer()
	raft.RegisterRaftServiceServer(s, consensusNode)

	go func() {
		log.Printf("Node %d gRPC server starting to serve on %s\n", *nodeID, selfAddr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Println("Waiting for peers to start...")
	time.Sleep(3 * time.Second)
	consensusNode.ConnectToPeers()
	consensusNode.Run()

	select {}
}

