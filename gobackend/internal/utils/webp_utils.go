package utils

import (
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"bytes"
	"sync"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// detectImageType 检测图片的真实类型
func detectImageType(path string) (string, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()
	
	// 读取文件前512字节以检测文件类型
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取文件头失败: %w", err)
	}
	
	// 检测文件类型
	contentType := http.DetectContentType(buffer)
	log.Printf("检测到文件类型: %s", contentType)
	
	// 根据MIME类型返回合适的图片格式
	switch contentType {
	case "image/jpeg":
		return "jpeg", nil
	case "image/png":
		return "png", nil
	case "image/gif":
		return "gif", nil
	case "image/webp":
		return "webp", nil
	default:
		if strings.HasPrefix(contentType, "image/") {
			// 其他图片类型
			return strings.TrimPrefix(contentType, "image/"), nil
		}
		return "", fmt.Errorf("不支持的图片类型: %s", contentType)
	}
}

// 根据EXIF方向数据调整图像方向
func correctImageOrientation(img image.Image, imgPath string) (image.Image, error) {
	// 打开文件以读取EXIF数据
	f, err := os.Open(imgPath)
	if err != nil {
		log.Printf("打开文件读取EXIF数据失败，将使用原始方向: %v", err)
		return img, nil
	}
	defer f.Close()

	// 尝试解码EXIF数据
	exifData, err := exif.Decode(f)
	if err != nil {
		// 如果无法解码EXIF（可能图片没有EXIF数据），使用原始图像
		log.Printf("解析EXIF数据失败，将使用原始方向: %v", err)
		return img, nil
	}

	// 尝试获取方向标签
	orientationTag, err := exifData.Get(exif.Orientation)
	if err != nil {
		log.Printf("获取方向信息失败，将使用原始方向: %v", err)
		return img, nil
	}

	// 获取方向值
	orientation, err := orientationTag.Int(0)
	if err != nil {
		log.Printf("解析方向值失败，将使用原始方向: %v", err)
		return img, nil
	}

	log.Printf("检测到EXIF方向: %d", orientation)

	// 根据EXIF方向值调整图像
	switch orientation {
	case 1:
		// 正常方向，不需要处理
		return img, nil
	case 2:
		// 水平翻转
		return imaging.FlipH(img), nil
	case 3:
		// 旋转180度
		return imaging.Rotate180(img), nil
	case 4:
		// 垂直翻转
		return imaging.FlipV(img), nil
	case 5:
		// 顺时针旋转90度后垂直翻转
		rotated := imaging.Rotate270(img)
		return imaging.FlipV(rotated), nil
	case 6:
		// 顺时针旋转90度
		return imaging.Rotate270(img), nil
	case 7:
		// 顺时针旋转90度后水平翻转
		rotated := imaging.Rotate270(img)
		return imaging.FlipH(rotated), nil
	case 8:
		// 顺时针旋转270度
		return imaging.Rotate90(img), nil
	default:
		// 未知方向，使用原始图像
		log.Printf("未知的EXIF方向值: %d，将使用原始方向", orientation)
		return img, nil
	}
}

// ConvertToWebP 将图片转换为WebP格式，自动判断图片方向并调整尺寸
// 参数imgPath可以是：
//  1. 相对于assets目录的路径，如 "imgs/13/1.jpg"
//  2. 相对于当前工作目录的路径，如 "diyu.jpg" 
//  3. 绝对路径，如 "/Users/username/path/to/image.jpg"
// 参数useWebpExt：是否使用.webp作为输出文件扩展名，默认为false，保持原扩展名
// 返回转换后的图片路径
func ConvertToWebP(imgPath string, useWebpExt ...bool) (string, error) {
	// 默认不使用.webp扩展名
	useWebp := false
	if len(useWebpExt) > 0 {
		useWebp = useWebpExt[0]
	}
	
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 构建assets目录的完整路径
	assetsDir := filepath.Join(workDir, "..", "assets")
	
	// 构建原图片的完整路径 - 检查多种情况
	var srcPath string
	
	// 判断路径类型
	if filepath.IsAbs(imgPath) {
		// 1. 绝对路径
		srcPath = imgPath
	} else if strings.HasPrefix(imgPath, "./") || strings.HasPrefix(imgPath, "../") {
		// 处理相对于当前目录的路径
		srcPath = filepath.Clean(filepath.Join(workDir, imgPath))
	} else if strings.Contains(imgPath, "/") || strings.Contains(imgPath, "\\") {
		// 2. 包含路径分隔符，假设是相对于assets目录的路径
		srcPath = filepath.Join(assetsDir, imgPath)
	} else {
		// 3. 简单文件名，假设是相对于当前工作目录
		srcPath = filepath.Join(workDir, imgPath)
	}
	
	log.Printf("处理图片路径: %s -> %s", imgPath, srcPath)
	
	// 检查源文件是否存在
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("源文件不存在: %s", srcPath)
	}
	
	// 读取源图片文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("打开源图片失败: %w", err)
	}
	defer srcFile.Close()
	
	// 首先尝试检测真实的图片类型
	imageType, typeErr := detectImageType(srcPath)
	if typeErr != nil {
		log.Printf("检测图片类型失败: %v，将尝试自动解码", typeErr)
	} else {
		log.Printf("检测到图片真实类型: %s", imageType)
	}
	
	// 将文件指针重置到开头
	if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %w", err)
	}
	
	// 先尝试使用通用解码器
	var srcImg image.Image
	var decodeErr error
	
	srcImg, format, decodeErr := image.Decode(srcFile)
	if decodeErr == nil {
		log.Printf("使用通用解码器成功解码图片，格式: %s", format)
	} else {
		log.Printf("通用解码器解码失败: %v，尝试专用解码器", decodeErr)
		
		// 重置文件指针
		srcFile.Close()
		srcFile, err = os.Open(srcPath)
		if err != nil {
			return "", fmt.Errorf("重新打开源图片失败: %w", err)
		}
		defer srcFile.Close()
		
		// 根据检测到的真实图片类型或扩展名选择专用解码器
		if imageType != "" {
			switch imageType {
			case "jpeg":
				srcImg, decodeErr = jpeg.Decode(srcFile)
			case "png":
				srcImg, decodeErr = png.Decode(srcFile)
			case "gif":
				srcImg, decodeErr = gif.Decode(srcFile)
			default:
				// 其他检测到的图片类型，再次尝试通用解码器
				if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
					return "", fmt.Errorf("重置文件指针失败: %w", err)
				}
				srcImg, _, decodeErr = image.Decode(srcFile)
			}
		} else {
			// 如果没有检测到类型，根据文件扩展名尝试
			ext := strings.ToLower(filepath.Ext(srcPath))
			switch ext {
			case ".jpg", ".jpeg":
				log.Printf("尝试使用JPEG解码器...")
				srcImg, decodeErr = jpeg.Decode(srcFile)
			case ".png":
				log.Printf("尝试使用PNG解码器...")
				srcImg, decodeErr = png.Decode(srcFile)
			case ".gif":
				log.Printf("尝试使用GIF解码器...")
				srcImg, decodeErr = gif.Decode(srcFile)
			default:
				// 最后一次尝试：读取全部内容并使用bytes.Reader尝试解码
				srcFile.Close()
				fileContent, err := os.ReadFile(srcPath)
				if err != nil {
					return "", fmt.Errorf("读取文件内容失败: %w", err)
				}
				
				reader := bytes.NewReader(fileContent)
				srcImg, _, decodeErr = image.Decode(reader)
			}
		}
		
		if decodeErr != nil {
			return "", fmt.Errorf("所有解码方法都失败，无法处理图片: %w", decodeErr)
		}
		
		log.Printf("使用专用解码器成功解码图片")
	}
	
	// 添加方向矫正逻辑
	srcImg, err = correctImageOrientation(srcImg, srcPath)
	if err != nil {
		log.Printf("图像方向矫正失败: %v, 使用未矫正的图像继续处理", err)
	} else {
		log.Printf("已根据EXIF数据完成图像方向矫正")
	}
	
	// 获取图片尺寸，自动判断是横图还是竖图
	bounds := srcImg.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()
	
	// 基于宽高比决定目标尺寸
	var maxWidth, maxHeight int
	if origWidth > origHeight {
		// 横图，使用1280×720
		maxWidth = 1280
		maxHeight = 720
		log.Printf("检测到横图 (%dx%d), 将调整为1280x720", origWidth, origHeight)
	} else {
		// 竖图，使用600×900
		maxWidth = 600
		maxHeight = 900
		log.Printf("检测到竖图 (%dx%d), 将调整为600x900", origWidth, origHeight)
	}
	
	// 计算新尺寸，保持宽高比
	var newWidth, newHeight int
	widthRatio := float64(maxWidth) / float64(origWidth)
	heightRatio := float64(maxHeight) / float64(origHeight)
	
	// 使用较小的比例，确保图片完全适合最大尺寸
	resizeRatio := widthRatio
	if heightRatio < widthRatio {
		resizeRatio = heightRatio
	}
	
	newWidth = int(float64(origWidth) * resizeRatio)
	newHeight = int(float64(origHeight) * resizeRatio)
	
	// 调整图片尺寸，保持原宽高比
	resizedImg := imaging.Resize(srcImg, newWidth, newHeight, imaging.Lanczos)
	
	// 构建输出文件的路径
	var outputPath string
	
	if useWebp {
		// 使用WebP扩展名
		baseFilename := strings.TrimSuffix(filepath.Base(imgPath), filepath.Ext(imgPath))
		
		// 确定WebP输出的目录
		var outputDir string
		
		if filepath.IsAbs(imgPath) || !strings.Contains(imgPath, "/") && !strings.Contains(imgPath, "\\") {
			// 对于绝对路径或简单文件名，保存在同一目录下
			outputDir = filepath.Dir(srcPath)
			outputPath = filepath.Join(outputDir, baseFilename + ".webp")
		} else {
			// 对于相对于assets的路径，保持原有结构
			baseDir := filepath.Dir(imgPath)
			outputPath = filepath.Join(assetsDir, baseDir, baseFilename + ".webp")
			outputDir = filepath.Dir(outputPath)
		}
		
		// 确保目标目录存在
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %w", err)
		}
	} else {
		// 保持原始扩展名，直接覆盖源文件
		outputPath = srcPath
	}
	
	// 为避免直接覆盖源文件导致问题，先创建临时文件
	tempFile := outputPath + ".tmp"
	dstFile, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer func() {
		dstFile.Close()
		// 如果出错，删除临时文件
		if err != nil {
			os.Remove(tempFile)
		}
	}()
	
	// 编码为WebP格式，质量设为80%，并且不保留元数据
	options := &webp.Options{
		Quality: 80,
		// WebP库默认会清除所有元数据和配置文件信息
	}
	err = webp.Encode(dstFile, resizedImg, options)
	if err != nil {
		return "", fmt.Errorf("编码WebP图片失败: %w", err)
	}
	
	// 关闭文件，确保写入完成
	dstFile.Close()
	
	// 重命名临时文件为目标文件名，实现原子替换
	if err := os.Rename(tempFile, outputPath); err != nil {
		return "", fmt.Errorf("替换原文件失败: %w", err)
	}
	
	if useWebp {
		log.Printf("成功将图片 %s 转换为WebP格式 %s (尺寸: %dx%d, 已去除元数据)", imgPath, outputPath, newWidth, newHeight)
	} else {
		log.Printf("成功将图片 %s 转换为WebP格式并保持原扩展名 %s (尺寸: %dx%d, 已去除元数据)", imgPath, outputPath, newWidth, newHeight)
	}
	
	// 返回处理后的图片路径
	return outputPath, nil
}

// ConvertMultipleImages 根据JSON列表批量转换多张图片
// 参数:
//   - imageList: 图片路径JSON字符串，如 `["img1.jpg", "img2.png"]` 或 `["/path/to/img1.jpg", "/path/to/img2.png"]`
//   - keepOriginal: 是否保留原始图片
//   - useWebpExt: 是否使用.webp扩展名
//   - concurrency: 并发处理的数量，0表示使用默认值(4)
//
// 返回转换后的图片路径列表和可能的错误
func ConvertMultipleImages(imageList string, keepOriginal bool, useWebpExt bool, concurrency int) ([]string, error) {
	// 解析图片列表
	var imagePaths []string
	if err := json.Unmarshal([]byte(imageList), &imagePaths); err != nil {
		return nil, fmt.Errorf("解析图片列表JSON失败: %w", err)
	}
	
	if len(imagePaths) == 0 {
		return []string{}, nil
	}
	
	log.Printf("开始批量处理图片列表，图片数量: %d", len(imagePaths))
	
	// 创建并发控制
	if concurrency <= 0 {
		concurrency = 4
	}
	
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	
	// 保存转换后的路径
	resultPaths := make([]string, len(imagePaths))
	errorMessages := make([]string, len(imagePaths))
	
	// 记录开始时间
	startTime := time.Now()
	
	// 处理单个图片的函数
	processImage := func(i int, imgPath string) {
		defer wg.Done()
		defer func() { <-semaphore }() // 释放信号量
		
		log.Printf("处理图片 [%d/%d]: %s", i+1, len(imagePaths), imgPath)
		
		outputPath, err := ConvertToWebPWithRatio(imgPath, 0, 0, keepOriginal, useWebpExt)
		if err != nil {
			log.Printf("转换图片失败 [%d/%d]: %s - %v", i+1, len(imagePaths), imgPath, err)
			errorMessages[i] = fmt.Sprintf("图片 %s 转换失败: %v", imgPath, err)
			return
		}
		
		resultPaths[i] = outputPath
	}
	
	// 并发处理每个图片
	for i, imgPath := range imagePaths {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		go processImage(i, imgPath)
	}
	
	// 等待所有图片处理完成
	wg.Wait()
	
	// 记录结束时间
	elapsedTime := time.Since(startTime)
	
	// 过滤出成功转换的路径
	validPaths := make([]string, 0, len(resultPaths))
	for _, path := range resultPaths {
		if path != "" {
			validPaths = append(validPaths, path)
		}
	}
	
	// 收集错误信息
	var errors []string
	for _, errMsg := range errorMessages {
		if errMsg != "" {
			errors = append(errors, errMsg)
		}
	}
	
	log.Printf("批量处理图片列表完成。成功: %d, 失败: %d, 耗时: %v", 
		len(validPaths), len(imagePaths)-len(validPaths), elapsedTime)
	
	if len(errors) > 0 {
		return validPaths, fmt.Errorf("部分图片处理失败: %s", strings.Join(errors, "; "))
	}
	
	return validPaths, nil
}

// isAnimatedGif 检查GIF是否包含多个帧（是否为动画GIF）
func isAnimatedGif(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("打开GIF文件失败: %w", err)
	}
	defer file.Close()

	// 解码整个GIF文件
	gifData, err := gif.DecodeAll(file)
	if err != nil {
		return false, fmt.Errorf("解码GIF文件失败: %w", err)
	}

	// 如果图像数量大于1，则为动画GIF
	return len(gifData.Image) > 1, nil
}

// convertAnimatedGif 使用gif2webp将动画GIF转换为动画WebP
func convertAnimatedGif(srcPath, outputPath string, quality float32) error {
	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建临时输出路径
	tempOutput := outputPath + ".tmp"

	// 检查gif2webp命令是否存在
	_, err := exec.LookPath("gif2webp")
	if err != nil {
		log.Printf("无法找到gif2webp命令: %v", err)
		log.Printf("请确保已安装libwebp-tools包，可以通过系统包管理器安装，例如：")
		log.Printf("Ubuntu/Debian: sudo apt-get install webp")
		log.Printf("CentOS/RHEL: sudo yum install libwebp-tools")
		log.Printf("MacOS: brew install webp")
		
		// 尝试使用内置的GIF解码和WebP编码逻辑进行转换
		log.Printf("尝试使用内置方法转换动画GIF...")
		return convertAnimatedGifUsingGoLibs(srcPath, outputPath, quality)
	}

	// 检查源文件是否存在且可读
	if _, err := os.Stat(srcPath); err != nil {
		return fmt.Errorf("源文件检查失败: %w", err)
	}
	
	// 获取绝对路径
	srcAbsPath, err := filepath.Abs(srcPath)
	if err != nil {
		log.Printf("获取源文件绝对路径失败: %v，使用原路径", err)
		srcAbsPath = srcPath
	}
	
	tempAbsOutput, err := filepath.Abs(tempOutput)
	if err != nil {
		log.Printf("获取临时文件绝对路径失败: %v，使用原路径", err)
		tempAbsOutput = tempOutput
	}

	// 准备gif2webp命令
	log.Printf("使用gif2webp命令转换动画GIF: %s -> %s", srcAbsPath, tempAbsOutput)
	
	// 尝试两种命令行参数组合
	success := false
	
	// 参数组合1: 标准方式
	args1 := []string{
		"-q", fmt.Sprintf("%.0f", quality),  // 质量参数
		"-mixed",                            // 混合无损/有损
		srcAbsPath,                          // 源文件路径
		"-o", tempAbsOutput,                 // 输出路径
	}
	
	log.Printf("尝试参数组合1: gif2webp %s", strings.Join(args1, " "))
	cmd1 := exec.Command("gif2webp", args1...)
	var stdout1, stderr1 bytes.Buffer
	cmd1.Stdout = &stdout1
	cmd1.Stderr = &stderr1
	
	if err := cmd1.Run(); err == nil {
		log.Printf("参数组合1执行成功")
		success = true
	} else {
		log.Printf("参数组合1执行失败: %v", err)
		log.Printf("标准输出: %s", stdout1.String())
		log.Printf("错误输出: %s", stderr1.String())
		
		// 如果第一种参数组合失败，尝试第二种
		args2 := []string{
			"-q", fmt.Sprintf("%.0f", quality),  // 质量参数
			srcAbsPath,                          // 源文件路径
			"-o", tempAbsOutput,                 // 输出路径
		}
		
		log.Printf("尝试参数组合2: gif2webp %s", strings.Join(args2, " "))
		cmd2 := exec.Command("gif2webp", args2...)
		var stdout2, stderr2 bytes.Buffer
		cmd2.Stdout = &stdout2
		cmd2.Stderr = &stderr2
		
		if err := cmd2.Run(); err == nil {
			log.Printf("参数组合2执行成功")
			success = true
		} else {
			log.Printf("参数组合2执行失败: %v", err)
			log.Printf("标准输出: %s", stdout2.String())
			log.Printf("错误输出: %s", stderr2.String())
			
			// 第三种组合：尝试直接调用系统命令
			commandStr := fmt.Sprintf("gif2webp -q %.0f '%s' -o '%s'", 
				quality, srcAbsPath, tempAbsOutput)
			log.Printf("尝试通过系统shell执行: %s", commandStr)
			
			shellCmd := exec.Command("sh", "-c", commandStr)
			var stdoutShell, stderrShell bytes.Buffer
			shellCmd.Stdout = &stdoutShell
			shellCmd.Stderr = &stderrShell
			
			if err := shellCmd.Run(); err == nil {
				log.Printf("Shell执行成功")
				success = true
			} else {
				log.Printf("Shell执行失败: %v", err)
				log.Printf("标准输出: %s", stdoutShell.String())
				log.Printf("错误输出: %s", stderrShell.String())
			}
		}
	}
	
	// 检查是否有任何一种方法成功
	if !success {
		log.Printf("所有参数组合都失败，尝试备用方法")
		// 尝试备选方法
		return convertAnimatedGifUsingGoLibs(srcPath, outputPath, quality)
	}

	// 检查临时文件是否存在且大小合理
	fileInfo, err := os.Stat(tempOutput)
	if err != nil {
		log.Printf("无法获取临时文件信息: %v", err)
		return fmt.Errorf("临时文件检查失败: %w", err)
	}
	
	log.Printf("生成的临时文件大小: %.2f KB", float64(fileInfo.Size())/1024)
	
	// 如果文件太小，可能转换不成功
	if fileInfo.Size() < 100 {
		log.Printf("警告：生成的文件过小，可能转换不完整")
	}

	// 重命名临时文件为目标文件名，实现原子替换
	log.Printf("重命名临时文件 %s -> %s", tempOutput, outputPath)
	if err := os.Rename(tempOutput, outputPath); err != nil {
		return fmt.Errorf("替换原文件失败: %w", err)
	}

	log.Printf("成功将动画GIF %s 转换为动画WebP %s", srcPath, outputPath)
	return nil
}

// convertAnimatedGifUsingGoLibs 使用Go库将动画GIF转换为动画WebP
// 这是一个备选方法，当外部gif2webp命令不可用时使用
func convertAnimatedGifUsingGoLibs(srcPath, outputPath string, quality float32) error {
	// 打开GIF文件
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开GIF文件失败: %w", err)
	}
	defer file.Close()

	// 解码GIF
	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		return fmt.Errorf("解码GIF文件失败: %w", err)
	}

	log.Printf("成功解码GIF文件，帧数: %d", len(gifImg.Image))

	// 确保至少有一帧
	if len(gifImg.Image) == 0 {
		return fmt.Errorf("GIF文件没有帧")
	}

	// 创建临时输出路径
	tempOutput := outputPath + ".tmp"

	// 创建临时文件
	outFile, err := os.Create(tempOutput)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer func() {
		outFile.Close()
		if err != nil {
			os.Remove(tempOutput)
		}
	}()

	// 获取第一帧
	firstFrame := gifImg.Image[0]
	
	// 准备WebP选项
	options := webp.Options{
		Lossless: false,
		Quality:  quality,
	}

	// 转换第一帧
	err = webp.Encode(outFile, firstFrame, &options)
	if err != nil {
		return fmt.Errorf("编码第一帧失败: %w", err)
	}

	log.Printf("警告：内置方法目前只能转换第一帧，结果将是静态WebP")
	log.Printf("要获得完整的动画支持，请安装并使用gif2webp命令行工具")

	// 关闭文件并重命名
	outFile.Close()
	if err := os.Rename(tempOutput, outputPath); err != nil {
		return fmt.Errorf("替换原文件失败: %w", err)
	}

	log.Printf("成功将GIF首帧转换为静态WebP %s", outputPath)
	return nil
}

// ConvertToWebPWithRatio 将图片转换为WebP格式，并保持原始宽高比调整尺寸
// 参数:
//   - imgPath: 图片路径，可以是：
//     1. 相对于assets目录的路径，如 "imgs/13/1.jpg"
//     2. 相对于当前工作目录的路径，如 "diyu.jpg" 
//     3. 绝对路径，如 "/Users/username/path/to/image.jpg"
//   - maxWidth: 最大宽度，0表示自动判断(横图1280，竖图600)
//   - maxHeight: 最大高度，0表示自动判断(横图720，竖图900)
//   - keepOriginal: 是否保留原始图片 - 当使用.webp后缀且此值为false时会删除原图
//   - useWebpExt: 是否使用.webp作为输出文件扩展名，默认为false，保持原扩展名
//
// 返回转换后的图片路径
func ConvertToWebPWithRatio(imgPath string, maxWidth, maxHeight int, keepOriginal bool, useWebpExt ...bool) (string, error) {
	// 默认不使用.webp扩展名
	useWebp := false
	if len(useWebpExt) > 0 {
		useWebp = useWebpExt[0]
	}
	
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 构建assets目录的完整路径
	assetsDir := filepath.Join(workDir, "..", "assets")
	
	// 构建原图片的完整路径 - 检查多种情况
	var srcPath string
	
	// 判断路径类型
	if filepath.IsAbs(imgPath) {
		// 1. 绝对路径
		srcPath = imgPath
	} else if strings.HasPrefix(imgPath, "./") || strings.HasPrefix(imgPath, "../") {
		// 处理相对于当前目录的路径
		srcPath = filepath.Clean(filepath.Join(workDir, imgPath))
	} else if strings.Contains(imgPath, "/") || strings.Contains(imgPath, "\\") {
		// 2. 包含路径分隔符，假设是相对于assets目录的路径
		srcPath = filepath.Join(assetsDir, imgPath)
	} else {
		// 3. 简单文件名，假设是相对于当前工作目录
		srcPath = filepath.Join(workDir, imgPath)
	}
	
	log.Printf("处理图片路径: %s -> %s", imgPath, srcPath)
	
	// 检查源文件是否存在
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("源文件不存在: %s", srcPath)
	}
	
	// 首先尝试检测真实的图片类型
	imageType, typeErr := detectImageType(srcPath)
	if typeErr != nil {
		log.Printf("检测图片类型失败: %v，将尝试自动解码", typeErr)
	} else {
		log.Printf("检测到图片真实类型: %s", imageType)
	}
	
	// 特殊处理GIF动画文件
	if imageType == "gif" || strings.ToLower(filepath.Ext(srcPath)) == ".gif" {
		// 检查是否为动画GIF
		isAnimated, err := isAnimatedGif(srcPath)
		if err != nil {
			log.Printf("检查GIF动画状态失败: %v，将作为普通图像处理", err)
		} else if isAnimated {
			log.Printf("检测到动画GIF，将使用特殊处理将其转换为动画WebP")
			
			// 确定输出路径
			var outputPath string
			var originalSaved bool = false
			
			if useWebp {
				// 使用WebP扩展名
				baseFilename := strings.TrimSuffix(filepath.Base(imgPath), filepath.Ext(imgPath))
				
				// 确定WebP输出的目录
				var outputDir string
				
				if filepath.IsAbs(imgPath) || (!strings.Contains(imgPath, "/") && !strings.Contains(imgPath, "\\")) {
					// 对于绝对路径或简单文件名，保存在同一目录下
					outputDir = filepath.Dir(srcPath)
					outputPath = filepath.Join(outputDir, baseFilename + ".webp")
				} else {
					// 对于相对于assets的路径，保持原有结构
					baseDir := filepath.Dir(imgPath)
					outputPath = filepath.Join(assetsDir, baseDir, baseFilename + ".webp")
					outputDir = filepath.Dir(outputPath)
				}
				
				// 确保目标目录存在
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					return "", fmt.Errorf("创建目标目录失败: %w", err)
				}
				
				// 当使用.webp扩展名时，原文件和新文件是不同的文件
				originalSaved = true
			} else {
				// 保持原始扩展名，直接覆盖源文件
				outputPath = srcPath
			}
			
			// 使用特殊转换函数处理动画GIF
			err = convertAnimatedGif(srcPath, outputPath, 80.0) // 使用80%的质量
			if err != nil {
				return "", fmt.Errorf("转换动画GIF失败: %w", err)
			}
			
			// 如果不需要保留原始图片，且生成了新的WebP文件，则删除原图
			if !keepOriginal && originalSaved {
				log.Printf("删除原始图片: %s", srcPath)
				if err := os.Remove(srcPath); err != nil {
					log.Printf("警告：删除原始图片失败: %v", err)
				}
			}
			
			return outputPath, nil
		}
	}
	
	// 读取源图片文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("打开源图片失败: %w", err)
	}
	defer srcFile.Close()
	
	// 将文件指针重置到开头
	if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %w", err)
	}
	
	// 先尝试使用通用解码器
	var srcImg image.Image
	var decodeErr error
	
	srcImg, format, decodeErr := image.Decode(srcFile)
	if decodeErr == nil {
		log.Printf("使用通用解码器成功解码图片，格式: %s", format)
	} else {
		log.Printf("通用解码器解码失败: %v，尝试专用解码器", decodeErr)
		
		// 重置文件指针
		srcFile.Close()
		srcFile, err = os.Open(srcPath)
		if err != nil {
			return "", fmt.Errorf("重新打开源图片失败: %w", err)
		}
		defer srcFile.Close()
		
		// 根据检测到的真实图片类型或扩展名选择专用解码器
		if imageType != "" {
			switch imageType {
			case "jpeg":
				srcImg, decodeErr = jpeg.Decode(srcFile)
			case "png":
				srcImg, decodeErr = png.Decode(srcFile)
			case "gif":
				srcImg, decodeErr = gif.Decode(srcFile)
			default:
				// 其他检测到的图片类型，再次尝试通用解码器
				if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
					return "", fmt.Errorf("重置文件指针失败: %w", err)
				}
				srcImg, _, decodeErr = image.Decode(srcFile)
			}
		} else {
			// 如果没有检测到类型，根据文件扩展名尝试
			ext := strings.ToLower(filepath.Ext(srcPath))
			switch ext {
			case ".jpg", ".jpeg":
				log.Printf("尝试使用JPEG解码器...")
				srcImg, decodeErr = jpeg.Decode(srcFile)
			case ".png":
				log.Printf("尝试使用PNG解码器...")
				srcImg, decodeErr = png.Decode(srcFile)
			case ".gif":
				log.Printf("尝试使用GIF解码器...")
				srcImg, decodeErr = gif.Decode(srcFile)
			default:
				// 最后一次尝试：读取全部内容并使用bytes.Reader尝试解码
				srcFile.Close()
				fileContent, err := os.ReadFile(srcPath)
				if err != nil {
					return "", fmt.Errorf("读取文件内容失败: %w", err)
				}
				
				reader := bytes.NewReader(fileContent)
				srcImg, _, decodeErr = image.Decode(reader)
			}
		}
		
		if decodeErr != nil {
			return "", fmt.Errorf("所有解码方法都失败，无法处理图片: %w", decodeErr)
		}
		
		log.Printf("使用专用解码器成功解码图片")
	}
	
	// 添加方向矫正逻辑
	srcImg, err = correctImageOrientation(srcImg, srcPath)
	if err != nil {
		log.Printf("图像方向矫正失败: %v, 使用未矫正的图像继续处理", err)
	} else {
		log.Printf("已根据EXIF数据完成图像方向矫正")
	}
	
	// 获取原始尺寸
	origBounds := srcImg.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()
	
	// 如果未指定尺寸，自动判断图片类型并设置合适的尺寸
	if maxWidth <= 0 || maxHeight <= 0 {
		if origWidth > origHeight {
			// 横图
			maxWidth = 1280
			maxHeight = 720
			log.Printf("检测到横图 (%dx%d), 将调整为最大1280x720", origWidth, origHeight)
		} else {
			// 竖图
			maxWidth = 600
			maxHeight = 900
			log.Printf("检测到竖图 (%dx%d), 将调整为最大600x900", origWidth, origHeight)
		}
	}
	
	// 计算新尺寸，保持宽高比
	var newWidth, newHeight int
	widthRatio := float64(maxWidth) / float64(origWidth)
	heightRatio := float64(maxHeight) / float64(origHeight)
	
	// 使用较小的比例，确保图片完全适合最大尺寸
	resizeRatio := widthRatio
	if heightRatio < widthRatio {
		resizeRatio = heightRatio
	}
	
	newWidth = int(float64(origWidth) * resizeRatio)
	newHeight = int(float64(origHeight) * resizeRatio)
	
	// 调整图片尺寸，保持原宽高比
	resizedImg := imaging.Resize(srcImg, newWidth, newHeight, imaging.Lanczos)
	
	// 确定输出路径
	var outputPath string
	var originalSaved bool = false
	
	if useWebp {
		// 使用WebP扩展名
		baseFilename := strings.TrimSuffix(filepath.Base(imgPath), filepath.Ext(imgPath))
		
		// 确定WebP输出的目录
		var outputDir string
		
		if filepath.IsAbs(imgPath) || !strings.Contains(imgPath, "/") && !strings.Contains(imgPath, "\\") {
			// 对于绝对路径或简单文件名，保存在同一目录下
			outputDir = filepath.Dir(srcPath)
			outputPath = filepath.Join(outputDir, baseFilename + ".webp")
		} else {
			// 对于相对于assets的路径，保持原有结构
			baseDir := filepath.Dir(imgPath)
			outputPath = filepath.Join(assetsDir, baseDir, baseFilename + ".webp")
			outputDir = filepath.Dir(outputPath)
		}
		
		// 确保目标目录存在
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %w", err)
		}
		
		// 当使用.webp扩展名时，原文件和新文件是不同的文件
		originalSaved = true
	} else {
		// 保持原始扩展名，直接覆盖源文件
		outputPath = srcPath
	}
	
	// 为避免直接覆盖源文件导致问题，先创建临时文件
	tempFile := outputPath + ".tmp"
	dstFile, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer func() {
		dstFile.Close()
		// 如果出错，删除临时文件
		if err != nil {
			os.Remove(tempFile)
		}
	}()
	
	// 编码为WebP格式，质量设为80%，并且不保留元数据
	options := &webp.Options{
		Quality: 80,
		// WebP库默认会清除所有元数据和配置文件信息
	}
	err = webp.Encode(dstFile, resizedImg, options)
	if err != nil {
		return "", fmt.Errorf("编码WebP图片失败: %w", err)
	}
	
	// 关闭文件，确保写入完成
	dstFile.Close()
	
	// 重命名临时文件为目标文件名，实现原子替换
	if err := os.Rename(tempFile, outputPath); err != nil {
		return "", fmt.Errorf("替换原文件失败: %w", err)
	}
	
	// 记录日志
	if useWebp {
		log.Printf("成功将图片 %s 转换为WebP格式 %s (尺寸: %dx%d, 已去除元数据)", 
			imgPath, outputPath, newWidth, newHeight)
	} else {
		log.Printf("成功将图片 %s 转换为WebP格式并保持原扩展名 %s (尺寸: %dx%d, 已去除元数据)", 
			imgPath, outputPath, newWidth, newHeight)
	}
	
	// 如果不需要保留原始图片，且生成了新的WebP文件，则删除原图
	if !keepOriginal && originalSaved {
		log.Printf("删除原始图片: %s", srcPath)
		if err := os.Remove(srcPath); err != nil {
			log.Printf("警告：删除原始图片失败: %v", err)
		}
	}
	
	// 返回处理后的图片路径
	return outputPath, nil
}

// BatchProcessImages 批量处理指定目录下的图片
// 参数:
//   - dirPath: 要处理的目录路径
//   - recursive: 是否递归处理子目录
//   - keepOriginal: 是否保留原始图片
//   - useWebpExt: 是否使用.webp扩展名
//   - concurrency: 并发处理的数量，0表示使用默认值(CPU核心数)
//
// 返回处理的图片数量和可能的错误
func BatchProcessImages(dirPath string, recursive, keepOriginal bool, useWebpExt bool, concurrency int) (int, error) {
	// 创建并发控制的通道
	if concurrency <= 0 {
		// 使用默认的并发数量，通常是CPU核心数
		concurrency = 4 // 设置一个合理的默认值
	}
	
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	
	// 统计处理的文件数量
	processedCount := 0
	var countMutex sync.Mutex
	
	// 记录开始时间
	startTime := time.Now()
	
	log.Printf("开始批量处理目录: %s (递归=%v, 保留原图=%v, 使用WebP扩展名=%v, 并发数=%d)",
		dirPath, recursive, keepOriginal, useWebpExt, concurrency)
	
	// 处理文件的函数
	processFile := func(filePath string) {
		defer wg.Done()
		defer func() { <-semaphore }() // 释放信号量
		
		ext := strings.ToLower(filepath.Ext(filePath))
		// 只处理常见图片格式
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			log.Printf("处理文件: %s", filePath)
			
			// 打开图片确定方向
			img, err := imaging.Open(filePath)
			if err != nil {
				log.Printf("无法打开图片 %s: %v", filePath, err)
				return
			}
			
			// 获取图片原始尺寸
			bounds := img.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()
			
			// 判断是横图还是竖图
			var maxWidth, maxHeight int
			if width > height {
				// 横图，设置为1280x720
				maxWidth = 1280
				maxHeight = 720
				log.Printf("检测到横图 %s (%dx%d), 将调整为1280x720", filePath, width, height)
			} else {
				// 竖图，设置为600x900
				maxWidth = 600
				maxHeight = 900
				log.Printf("检测到竖图 %s (%dx%d), 将调整为600x900", filePath, width, height)
			}
			
			// 转换图片
			_, err = ConvertToWebPWithRatio(filePath, maxWidth, maxHeight, keepOriginal, useWebpExt)
			if err != nil {
				log.Printf("转换图片失败 %s: %v", filePath, err)
				return
			}
			
			// 增加处理计数
			countMutex.Lock()
			processedCount++
			countMutex.Unlock()
		}
	}
	
	// 遍历文件的函数
	walkFn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("访问路径失败 %s: %v", path, err)
			return nil // 继续遍历其他文件
		}
		
		// 如果是目录但不需要递归处理，则跳过子目录
		if info.IsDir() {
			if path != dirPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		
		// 处理文件
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		go processFile(path)
		
		return nil
	}
	
	// 开始遍历目录
	err := filepath.Walk(dirPath, walkFn)
	
	// 等待所有文件处理完成
	wg.Wait()
	
	// 记录结束时间
	elapsedTime := time.Since(startTime)
	
	log.Printf("批量处理完成。共处理 %d 个文件，耗时：%v", processedCount, elapsedTime)
	
	return processedCount, err
}

// ProcessDirectorySync 同步处理指定目录下的图片（不使用并发）
// 这是一个更简单的版本，适用于需要按顺序处理的情况
func ProcessDirectorySync(dirPath string, recursive, keepOriginal bool, useWebpExt bool) (int, error) {
	processedCount := 0
	startTime := time.Now()
	
	log.Printf("开始同步处理目录: %s (递归=%v, 保留原图=%v, 使用WebP扩展名=%v)",
		dirPath, recursive, keepOriginal, useWebpExt)
	
	// 遍历文件的函数
	walkFn := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("访问路径失败 %s: %v", path, err)
			return nil // 继续遍历其他文件
		}
		
		// 如果是目录但不需要递归处理，则跳过子目录
		if info.IsDir() {
			if path != dirPath && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		// 只处理常见图片格式
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" {
			log.Printf("处理文件: %s", path)
			
			// 打开图片确定方向
			img, err := imaging.Open(path)
			if err != nil {
				log.Printf("无法打开图片 %s: %v", path, err)
				return nil
			}
			
			// 获取图片原始尺寸
			bounds := img.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()
			
			// 判断是横图还是竖图
			var maxWidth, maxHeight int
			if width > height {
				// 横图，设置为1280x720
				maxWidth = 1280
				maxHeight = 720
				log.Printf("检测到横图 %s (%dx%d), 将调整为1280x720", path, width, height)
			} else {
				// 竖图，设置为600x900
				maxWidth = 600
				maxHeight = 900
				log.Printf("检测到竖图 %s (%dx%d), 将调整为600x900", path, width, height)
			}
			
			// 转换图片
			_, err = ConvertToWebPWithRatio(path, maxWidth, maxHeight, keepOriginal, useWebpExt)
			if err != nil {
				log.Printf("转换图片失败 %s: %v", path, err)
				return nil
			}
			
			processedCount++
		}
		
		return nil
	}
	
	// 开始遍历目录
	err := filepath.Walk(dirPath, walkFn)
	
	// 记录结束时间
	elapsedTime := time.Since(startTime)
	
	log.Printf("同步处理完成。共处理 %d 个文件，耗时：%v", processedCount, elapsedTime)
	
	return processedCount, err
} 