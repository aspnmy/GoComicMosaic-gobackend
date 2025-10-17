package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// 临时注释AVIF支持，使用标准库处理其他格式
)

// ConvertToAvif 将图片转换为AVIF格式
// 参数:
// - imgPath: 输入图片路径
// - useAvifExt: 是否使用.avif扩展名（如果为false，则保持原扩展名）
// - quality: AVIF压缩质量(0-100)
// 返回值:
// - 输出图片路径
// - 错误信息
func ConvertToAvif(imgPath string, useAvifExt bool, quality int) (string, error) {
	// 临时禁用AVIF支持，返回错误信息
	return "", fmt.Errorf("AVIF支持暂时不可用，请稍后再试")
}

// ConvertToAvifWithRatio 将图片转换为AVIF格式并保持原始宽高比
// 参数:
// - imgPath: 输入图片路径
// - maxWidth: 最大宽度(0表示自动判断)
// - maxHeight: 最大高度(0表示自动判断)
// - keepOriginal: 是否保留原始图片
// - useAvifExt: 是否使用.avif扩展名
// - quality: AVIF压缩质量(0-100)
// 返回值:
// - 输出图片路径
// - 错误信息
func ConvertToAvifWithRatio(imgPath string, maxWidth, maxHeight int, keepOriginal, useAvifExt bool, quality int) (string, error) {
	// 临时禁用AVIF支持，返回错误信息
	return "", fmt.Errorf("AVIF支持暂时不可用，请稍后再试")
}

// ConvertMultipleImagesToAvif 处理JSON列表中的多张图片，转换为AVIF格式
// 参数:
// - jsonList: JSON格式的图片路径列表
// - keepOriginal: 是否保留原始图片
// - useAvifExt: 是否使用.avif扩展名
// - concurrency: 并发处理的数量
// - quality: AVIF压缩质量
// 返回值:
// - 转换成功的图片路径列表
// - 错误信息
func ConvertMultipleImagesToAvif(jsonList string, keepOriginal, useAvifExt bool, concurrency, quality int) ([]string, error) {
	// 解析JSON列表
	var imgPaths []string
	if err := json.Unmarshal([]byte(jsonList), &imgPaths); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(imgPaths) == 0 {
		return []string{}, nil
	}

	// 限制并发数量
	if concurrency <= 0 {
		concurrency = 4
	}

	// 使用通道控制并发
	jobs := make(chan string, len(imgPaths))
	results := make(chan string, len(imgPaths))
	errs := make(chan error, len(imgPaths))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for imgPath := range jobs {
				// 尝试转换图片
				outputPath, err := ConvertToAvif(imgPath, useAvifExt, quality)
				if err != nil {
					errs <- fmt.Errorf("处理 %s 失败: %w", imgPath, err)
				} else {
					results <- outputPath
				}
			}
		}()
	}

	// 发送任务到通道
	for _, imgPath := range imgPaths {
		jobs <- imgPath
	}
	close(jobs)

	// 等待所有工作协程完成
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// 收集结果
	var successResults []string
	var combinedErr []string

	for result := range results {
		successResults = append(successResults, result)
	}

	for err := range errs {
		combinedErr = append(combinedErr, err.Error())
	}

	// 如果有错误，返回部分成功的结果和错误信息
	if len(combinedErr) > 0 {
		return successResults, errors.New(strings.Join(combinedErr, "; "))
	}

	return successResults, nil
}

// ProcessDirectoryToAvifSync 同步处理目录中的所有图片，转换为AVIF格式
// 参数:
// - dirPath: 目录路径
// - recursive: 是否递归处理子目录
// - keepOriginal: 是否保留原始图片
// - useAvifExt: 是否使用.avif扩展名
// - quality: AVIF压缩质量
// 返回值:
// - 处理的文件数量
// - 错误信息
func ProcessDirectoryToAvifSync(dirPath string, recursive bool, keepOriginal, useAvifExt bool, quality int) (int, error) {
	var count int

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是目录，直接返回
		if info.IsDir() {
			return nil
		}

		// 检查文件是否为支持的图片格式
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			// 处理图片
			_, err = ConvertToAvif(path, useAvifExt, quality)
			if err != nil {
				log.Printf("处理 %s 失败: %v", path, err)
				return nil // 继续处理其他文件
			}
			count++
			return nil
		default:
			return nil // 跳过不支持的文件格式
		}
	}

	// 遍历目录
	if recursive {
		err := filepath.Walk(dirPath, walkFunc)
		return count, err
	} else {
		// 非递归遍历
		files, err := os.ReadDir(dirPath)
		if err != nil {
			return 0, err
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			path := filepath.Join(dirPath, file.Name())
			walkFunc(path, &fileInfo{
				name:      file.Name(),
				size:      0,           // 忽略大小
				mode:      0,           // 忽略模式
				modTime:   time.Time{}, // 忽略修改时间
				isDir:     false,
				isSymlink: false,
			}, nil)
		}
		return count, nil
	}
}

// BatchProcessImagesToAvif 并发处理目录中的所有图片，转换为AVIF格式
// 参数:
// - dirPath: 目录路径
// - recursive: 是否递归处理子目录
// - keepOriginal: 是否保留原始图片
// - useAvifExt: 是否使用.avif扩展名
// - concurrency: 并发处理的数量
// - quality: AVIF压缩质量
// 返回值:
// - 处理的文件数量
// - 错误信息
func BatchProcessImagesToAvif(dirPath string, recursive bool, keepOriginal, useAvifExt bool, concurrency, quality int) (int, error) {
	// 收集所有图片路径
	var imgPaths []string

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是目录，直接返回
		if info.IsDir() {
			return nil
		}

		// 检查文件是否为支持的图片格式
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			imgPaths = append(imgPaths, path)
		}
		return nil
	}

	// 遍历目录收集图片
	var err error
	if recursive {
		err = filepath.Walk(dirPath, walkFunc)
	} else {
		// 非递归遍历
		files, readErr := os.ReadDir(dirPath)
		if readErr != nil {
			return 0, readErr
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			path := filepath.Join(dirPath, file.Name())
			walkFunc(path, &fileInfo{
				name:      file.Name(),
				size:      0,
				mode:      0,
				modTime:   time.Time{},
				isDir:     false,
				isSymlink: false,
			}, nil)
		}
	}

	if err != nil {
		return 0, err
	}

	// 并发处理收集到的图片
	var wg sync.WaitGroup
	var countMutex sync.Mutex
	var count int

	// 限制并发数量
	if concurrency <= 0 {
		concurrency = 4
	}

	// 创建任务通道
	jobs := make(chan string, len(imgPaths))

	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for imgPath := range jobs {
				// 处理图片
				_, err := ConvertToAvif(imgPath, useAvifExt, quality)
				if err != nil {
					log.Printf("处理 %s 失败: %v", imgPath, err)
					continue
				}

				// 更新计数
				countMutex.Lock()
				count++
				countMutex.Unlock()
			}
		}()
	}

	// 发送任务
	for _, imgPath := range imgPaths {
		jobs <- imgPath
	}
	close(jobs)

	// 等待所有工作协程完成
	wg.Wait()

	return count, nil
}

// 辅助函数: 计算目标尺寸，保持宽高比
func calculateTargetSize(originalWidth, originalHeight, maxWidth, maxHeight int) (int, int) {
	// 如果未指定最大尺寸，则使用默认值
	if maxWidth <= 0 || maxHeight <= 0 {
		// 根据图片方向设置默认尺寸
		if originalWidth > originalHeight {
			// 横图
			maxWidth = 1280
			maxHeight = 720
		} else {
			// 竖图
			maxWidth = 600
			maxHeight = 900
		}
	}

	// 计算缩放比例
	widthRatio := float64(maxWidth) / float64(originalWidth)
	heightRatio := float64(maxHeight) / float64(originalHeight)

	// 使用较小的缩放比例，以确保图片完全适应目标尺寸
	scaleRatio := widthRatio
	if heightRatio < widthRatio {
		scaleRatio = heightRatio
	}

	// 如果图片已经小于目标尺寸，则保持原始尺寸
	if scaleRatio > 1.0 {
		scaleRatio = 1.0
	}

	// 计算新的尺寸
	newWidth := int(float64(originalWidth) * scaleRatio)
	newHeight := int(float64(originalHeight) * scaleRatio)

	return newWidth, newHeight
}

// 辅助函数: 获取图片类型
func getImageType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	// 移除点号
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	return ext
}

// 为非递归遍历提供的简单os.FileInfo实现
// 因为os.ReadDir返回的是os.DirEntry，需要转换为os.FileInfo

type fileInfo struct {
	name      string
	size      int64
	mode      os.FileMode
	modTime   time.Time
	isDir     bool
	isSymlink bool
}

func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) Mode() os.FileMode  { return f.mode }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) IsDir() bool        { return f.isDir }
func (f *fileInfo) Sys() interface{}   { return nil }
func (f *fileInfo) IsSymlink() bool    { return f.isSymlink }
