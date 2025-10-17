package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	// 全局配置变量
	DbPath    string
	AssetsDir string
	AssetPath string // 保留原有变量以保持兼容性
	Version   string = "dev" // 版本号，由构建脚本注入，默认为dev
)

// 初始化配置
func init() {
	// 初始化数据库路径
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		DbPath = envPath
		log.Printf("使用环境变量指定的数据库路径: %s", DbPath)
	} else {
		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			log.Printf("获取工作目录失败: %v，使用默认路径", err)
			workDir = "."
		}
		
		// 使用默认数据库文件路径
		DbPath = filepath.Join(workDir, "resource_hub.db")
		log.Printf("使用默认数据库路径: %s", DbPath)
	}
	
	// 初始化资源目录
	if envPath := os.Getenv("ASSETS_PATH"); envPath != "" {
		AssetsDir = envPath
		AssetPath = envPath // 保持兼容性
		log.Printf("使用环境变量指定的资源目录: %s", AssetsDir)
	} else {
		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			log.Printf("获取工作目录失败: %v，使用默认路径", err)
			workDir = "."
		}
		
		// 使用默认资源目录路径
		AssetsDir = filepath.Join(workDir, "..", "assets")
		AssetPath = AssetsDir // 保持兼容性
		log.Printf("使用默认资源目录: %s", AssetsDir)
	}
	
	// 确保目录存在
	ensureDirExists(filepath.Dir(DbPath))
	ensureDirExists(AssetsDir)
	ensureDirExists(filepath.Join(AssetsDir, "uploads"))
	ensureDirExists(filepath.Join(AssetsDir, "imgs"))
	ensureDirExists(filepath.Join(AssetsDir, "public"))
}

// 确保目录存在
func ensureDirExists(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("创建目录失败 %s: %v", dir, err)
	}
}

// GetAssetsDir 获取资源目录路径
func GetAssetsDir() string {
	return AssetsDir
}

// GetDbPath 获取数据库路径
func GetDbPath() string {
	return DbPath
}

// GetVersion 获取应用版本号
// 返回当前应用的版本号，由构建脚本在编译时注入
func GetVersion() string {
	return Version
}