package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/config"

	"github.com/google/uuid"
)

// CalculateFileHash 计算文件内容的SHA-256哈希值，并将文件指针重置到开头
func CalculateFileHash(file io.ReadSeeker) (string, error) {
	hasher := sha256.New()

	// 通过4KB的块读取文件，这样可以处理大文件
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if n > 0 {
			hasher.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("读取文件失败: %w", err)
		}
	}

	// 计算哈希值
	hash := hex.EncodeToString(hasher.Sum(nil))

	// 重置文件指针到开头
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %w", err)
	}

	return hash, nil
}

// SaveUploadedFile 保存上传的文件到uploads目录，并返回相对路径
func SaveUploadedFile(file io.Reader, filename string) (string, error) {
	// 检查并创建上传目录
	uploadDir, err := ensureUploadDir()
	if err != nil {
		return "", err
	}

	// 生成唯一文件名
	uniqueFilename := generateUniqueFilename(filename)
	filePath := filepath.Join(uploadDir, uniqueFilename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回相对于服务器的路径
	return filepath.Join("/assets/uploads", time.Now().Format("20060102"), uniqueFilename), nil
}

// MoveApprovedImages 移动已批准的图片到资源目录
func MoveApprovedImages(resourceID int, imagePaths []string) ([]string, error) {
	if len(imagePaths) == 0 {
		return []string{}, nil
	}

	// 获取资源目录
	assetsDir := config.GetAssetsDir()

	// 创建目标目录
	imgsDir := filepath.Join(assetsDir, "imgs", fmt.Sprintf("%d", resourceID))
	if err := os.MkdirAll(imgsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 移动每个图片
	newPaths := make([]string, 0, len(imagePaths))
	for _, imgPath := range imagePaths {
		if imgPath == "" {
			continue
		}

		// 提取文件名
		filename := filepath.Base(imgPath)

		// 源文件路径
		sourcePath := filepath.Join(assetsDir, imgPath[7:]) // 去掉前面的"/assets"

		// 目标文件路径
		destPath := filepath.Join(imgsDir, filename)

		// 检查源文件是否存在
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			log.Printf("源文件不存在: %s", sourcePath)
			newPaths = append(newPaths, imgPath) // 保留原路径
			continue
		}

		// 移动文件 (先复制后删除)
		if err := moveFile(sourcePath, destPath); err != nil {
			log.Printf("移动图片失败: %s -> %s, 错误: %v", sourcePath, destPath, err)
			newPaths = append(newPaths, imgPath) // 保留原路径
			continue
		}

		log.Printf("成功移动图片: %s -> %s", sourcePath, destPath)

		// 更新路径
		newPath := fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename)
		newPaths = append(newPaths, newPath)
	}

	return newPaths, nil
}

// MoveApprovedImage 移动单个已批准的图片
func MoveApprovedImage(resourceID int, imagePath string) (string, error) {
	if imagePath == "" {
		return "", nil
	}

	// 获取资源目录
	assetsDir := config.GetAssetsDir()

	// 创建目标目录
	imgsDir := filepath.Join(assetsDir, "imgs", fmt.Sprintf("%d", resourceID))
	if err := os.MkdirAll(imgsDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 提取文件名
	filename := filepath.Base(imagePath)

	// 源文件路径
	sourcePath := filepath.Join(assetsDir, imagePath[7:]) // 去掉前面的"/assets"

	// 目标文件路径
	destPath := filepath.Join(imgsDir, filename)

	// 检查源文件是否存在
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Printf("源文件不存在: %s", sourcePath)
		return imagePath, nil // 保留原路径
	}

	// 移动文件 (先复制后删除)
	if err := moveFile(sourcePath, destPath); err != nil {
		log.Printf("移动图片失败: %s -> %s, 错误: %v", sourcePath, destPath, err)
		return imagePath, nil // 保留原路径
	}

	log.Printf("成功移动图片: %s -> %s", sourcePath, destPath)

	// 返回新路径
	return fmt.Sprintf("/assets/imgs/%d/%s", resourceID, filename), nil
}

// ConvertImageToPreferredFormat 根据配置的图片格式将图片转换为对应的格式
// 参数:
// - imgPath: 输入图片路径
// - useFormatExt: 是否使用对应格式的扩展名（如果为false，则保持原扩展名）
// - quality: 图片压缩质量(0-100)，仅对AVIF有效，WebP使用默认质量
// 返回值:
// - 输出图片路径
// - 错误信息
func ConvertImageToPreferredFormat(imgPath string, useFormatExt bool, quality int) (string, error) {
	// 获取配置的图片格式
	imageFormat := config.GetImageFormat()
	log.Printf("使用图片格式配置: %s 处理图片: %s", imageFormat, imgPath)
	
	switch imageFormat {
	case "avif":
		// 使用AVIF格式转换
		return ConvertToAvif(imgPath, useFormatExt, quality)
	case "webp":
		// 使用WebP格式转换
		return ConvertToWebP(imgPath, useFormatExt)
	default:
		// 默认使用WebP
		log.Printf("未知的图片格式配置: %s，使用默认格式WebP", imageFormat)
		return ConvertToWebP(imgPath, useFormatExt)
	}
}

// moveFile 移动文件（复制后删除原文件）
func moveFile(src, dst string) error {
	// 检查源文件是否存在
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", src)
	}

	// 检查目标路径
	if _, err := os.Stat(dst); err == nil {
		// 如果目标文件已存在，先删除它
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("删除已存在的目标文件失败: %w", err)
		}
	}

	// 首先复制文件
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 确认目标文件已成功创建
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return fmt.Errorf("目标文件创建失败: %s", dst)
	}

	// 然后删除源文件
	if err := os.Remove(src); err != nil {
		log.Printf("警告：无法删除源文件 %s: %v，将重试", src, err)

		// 重试删除
		time.Sleep(100 * time.Millisecond)
		if err := os.Remove(src); err != nil {
			log.Printf("警告：第二次尝试删除源文件 %s 失败: %v", src, err)
		} else {
			log.Printf("第二次尝试删除源文件 %s 成功", src)
		}
	} else {
		log.Printf("源文件 %s 已成功删除", src)
	}

	return nil
}

// 辅助函数

// ensureUploadDir 确保上传目录存在
func ensureUploadDir() (string, error) {
	// 获取资源目录
	assetsDir := config.GetAssetsDir()

	// 按日期创建上传目录
	dateDir := time.Now().Format("20060102")
	uploadDir := filepath.Join(assetsDir, "uploads", dateDir)

	// 确保目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	return uploadDir, nil
}

// generateUniqueFilename 生成唯一的文件名
func generateUniqueFilename(originalFilename string) string {
	// 获取文件扩展名
	ext := filepath.Ext(originalFilename)

	// 生成UUID
	uuid := uuid.New().String()

	// 生成唯一文件名
	return fmt.Sprintf("%s%s", uuid, ext)
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	// 复制内容
	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	return nil
}
