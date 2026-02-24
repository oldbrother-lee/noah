// 从测试环境（config 中 data.db.user）读取 api 表，生成 API 种子源码 api_seed.go，供「初始化数据库」时使用。
// 生成一次后可手动维护 internal/server/source/api_seed.go。
//
// 使用方式（在项目根目录）：
//
//	go run ./tools/api-seed-gen -conf config/local.yml
//	go run ./tools/api-seed-gen -conf config/local.yml -out internal/server/source/api_seed.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"go-noah/internal/model"
	"go-noah/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	confPath := flag.String("conf", "config/local.yml", "配置文件路径，其中 data.db.user 为测试环境连接")
	outPath := flag.String("out", "internal/server/source/api_seed.go", "输出的 api 种子 Go 文件路径")
	flag.Parse()

	conf := config.NewConfig(*confPath)
	driver := conf.GetString("data.db.user.driver")
	dsn := conf.GetString("data.db.user.dsn")
	if dsn == "" {
		fmt.Fprintf(os.Stderr, "错误: data.db.user.dsn 未配置，请检查 %s\n", *confPath)
		os.Exit(1)
	}

	db, err := openDB(driver, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	var apis []model.Api
	if err := db.Find(&apis).Error; err != nil {
		fmt.Fprintf(os.Stderr, "读取 api 表失败: %v\n", err)
		os.Exit(1)
	}
	sort.Slice(apis, func(i, j int) bool {
		if apis[i].Group != apis[j].Group {
			return apis[i].Group < apis[j].Group
		}
		if apis[i].Path != apis[j].Path {
			return apis[i].Path < apis[j].Path
		}
		return apis[i].Method < apis[j].Method
	})

	outDir := filepath.Dir(*outPath)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}
	body := generateGoFile(apis)
	if err := os.WriteFile(*outPath, body, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ API 种子已生成: %s（共 %d 条）\n", *outPath, len(apis))
}

func generateGoFile(apis []model.Api) []byte {
	const header = `package source

// ApiSeedData API 管理初始化数据。由 tools/api-seed-gen 从测试环境导出生成一次，之后可手动维护本文件。
// 服务首次启动且 api 表为空时，会从此变量插入种子数据。
var ApiSeedData = []ApiSeedItem{
`
	const footer = "}\n"

	b := []byte(header)
	for _, a := range apis {
		line := fmt.Sprintf("\t{Group: %s, Name: %s, Path: %s, Method: %s},\n",
			strconv.Quote(a.Group),
			strconv.Quote(a.Name),
			strconv.Quote(a.Path),
			strconv.Quote(a.Method))
		b = append(b, line...)
	}
	b = append(b, footer...)
	return b
}

func openDB(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("不支持的 driver: %s", driver)
	}
}
