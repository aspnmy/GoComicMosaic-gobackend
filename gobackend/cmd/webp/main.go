package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"GoComicMosaic-gobackend/gobackend/internal/config"
	"GoComicMosaic-gobackend/gobackend/internal/models"
	"GoComicMosaic-gobackend/gobackend/internal/utils"
)

func main() {
	// 打印版本信息
	fmt.Printf("WebP工具版本: %s\n", config.Version)

	// 解析命令行参数
	imgPath := flag.String("img", "", "要转换的图片路径，可以是相对路径或绝对路径")
	dirPath := flag.String("dir", "", "要批量处理的目录路径")
	jsonList := flag.String("json", "", "图片路径的JSON列表，例如：[\"img1.jpg\",\"img2.png\"]")
	recursive := flag.Bool("recursive", true, "是否递归处理子目录")
	keepRatio := flag.Bool("ratio", true, "是否保持原始宽高比")
	maxWidth := flag.Int("w", 0, "最大宽度(0表示自动判断：横图1280px，竖图600px)")
	maxHeight := flag.Int("h", 0, "最大高度(0表示自动判断：横图720px，竖图900px)")
	keepOriginal := flag.Bool("keep", true, "是否保留原始图片")
	useWebp := flag.Bool("webp", false, "是否使用.webp扩展名（默认保持原扩展名）")
	concurrency := flag.Int("concurrency", 4, "并发处理的数量，仅批量处理时有效")
	noAsync := flag.Bool("sync", false, "使用同步模式处理（不使用并发），批量处理时有效")
	versionFlag := flag.Bool("version", false, "显示版本号")
	flag.Parse()

	// 如果只请求显示版本号，则直接退出
	if *versionFlag {
		return
	}

	// 检查是否提供了图片路径、目录路径或JSON列表
	if *imgPath == "" && *dirPath == "" && *jsonList == "" {
		fmt.Println("请提供要转换的图片路径(-img)、要批量处理的目录路径(-dir)或图片路径的JSON列表(-json)")
		flag.Usage()
		os.Exit(1)
	}

	// 初始化数据库连接，以确保可以正常使用utils包
	db, err := models.InitDB()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 处理JSON列表批量转换
	if *jsonList != "" {
		log.Printf("开始处理JSON图片列表")
		log.Printf("参数: 保留原图=%v, 使用WebP扩展名=%v, 并发数=%d",
			*keepOriginal, *useWebp, *concurrency)

		// 调用处理JSON列表的函数
		results, err := utils.ConvertMultipleImages(*jsonList, *keepOriginal, *useWebp, *concurrency)
		if err != nil {
			log.Printf("警告: 部分图片处理失败: %v", err)
		}

		// 打印输出结果
		fmt.Println("\nJSON列表转换完成!")
		fmt.Printf("成功处理图片数量: %d\n", len(results))

		if len(results) > 0 {
			fmt.Println("\n转换后的图片路径:")
			for i, path := range results {
				if i < 10 || i == len(results)-1 { // 只显示前10个和最后1个
					fmt.Printf("  %s\n", path)
				} else if i == 10 {
					fmt.Printf("  ... 以及其他 %d 张图片\n", len(results)-11)
				}
			}
		}
		return
	}

	// 处理目录批量转换
	if *dirPath != "" {
		// 验证目录是否存在
		if _, err := os.Stat(*dirPath); os.IsNotExist(err) {
			log.Fatalf("指定的目录不存在: %s", *dirPath)
		}

		// 获取绝对路径
		absPath, err := filepath.Abs(*dirPath)
		if err != nil {
			log.Fatalf("获取目录绝对路径失败: %v", err)
		}

		log.Printf("开始批量处理目录: %s", absPath)
		log.Printf("参数: 递归处理=%v, 保留原图=%v, 使用WebP扩展名=%v, 并发数=%d, 同步模式=%v",
			*recursive, *keepOriginal, *useWebp, *concurrency, *noAsync)

		var count int

		// 选择同步或异步处理方式
		if *noAsync {
			count, err = utils.ProcessDirectorySync(absPath, *recursive, *keepOriginal, *useWebp)
		} else {
			count, err = utils.BatchProcessImages(absPath, *recursive, *keepOriginal, *useWebp, *concurrency)
		}

		if err != nil {
			log.Fatalf("批量处理失败: %v", err)
		}

		// 打印输出结果
		fmt.Println("\n批量转换成功!")
		fmt.Printf("处理目录: %s\n", absPath)
		fmt.Printf("处理文件数量: %d\n", count)
		return
	}

	// 单个图片处理
	log.Printf("开始处理图片: %s", *imgPath)
	log.Printf("参数: 保持比例=%v, 最大宽度=%d, 最大高度=%d, 保留原图=%v, 使用WebP扩展名=%v",
		*keepRatio, *maxWidth, *maxHeight, *keepOriginal, *useWebp)

	// 验证图片路径是否存在
	var fullPath string

	// 检查是否为绝对路径
	if filepath.IsAbs(*imgPath) {
		fullPath = *imgPath
	} else {
		// 相对路径，先尝试当前目录
		workDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("获取工作目录失败: %v", err)
		}

		fullPath = filepath.Join(workDir, *imgPath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// 文件不存在，打印更详细的提示
		log.Printf("注意: 文件在 %s 不存在", fullPath)
		log.Printf("将尝试按原始路径处理: %s", *imgPath)
	} else {
		log.Printf("找到文件: %s", fullPath)
		// 使用完整路径替换输入路径
		*imgPath = fullPath
	}

	var outputPath string

	// 根据参数选择合适的转换方法
	if *keepRatio {
		outputPath, err = utils.ConvertToWebPWithRatio(*imgPath, *maxWidth, *maxHeight, *keepOriginal, *useWebp)
	} else {
		outputPath, err = utils.ConvertToWebP(*imgPath, *useWebp)
	}

	if err != nil {
		log.Fatalf("图片转换失败: %v", err)
	}

	// 打印输出结果
	fmt.Println("\n转换成功!")
	fmt.Printf("原始图片: %s\n", *imgPath)

	// 确定输出格式的显示方式
	if *useWebp {
		fmt.Printf("WebP图片路径: %s\n", outputPath)
	} else {
		fmt.Printf("输出图片路径: %s (内容已转换为WebP格式但保留原扩展名)\n", outputPath)
	}

	// 检查输出文件是否存在
	if _, err := os.Stat(outputPath); err == nil {
		fileInfo, err := os.Stat(outputPath)
		if err == nil {
			fmt.Printf("输出文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
		}
	}
}
