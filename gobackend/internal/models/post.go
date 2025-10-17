package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Post 表示一篇完整的文章
type Post struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsMarkdown  bool      `json:"is_markdown"`
	Tags        []string  `json:"tags"`
	Cover       string    `json:"cover"`
	IsPublished bool      `json:"is_published"`
}

// PostSummary 表示文章的摘要信息
type PostSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsMarkdown  bool      `json:"is_markdown"`
	Tags        []string  `json:"tags"`
	Cover       string    `json:"cover"`
	IsPublished bool      `json:"is_published"`
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateSlug 从标题生成slug
func GenerateSlug(title string) string {
	// 将标题转换为小写
	slug := strings.ToLower(title)
	
	// 替换非字母数字字符为连字符
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	
	// 删除开头和结尾的连字符
	slug = strings.Trim(slug, "-")
	
	return slug
}

// SavePost 保存文章（创建或更新）
func SavePost(post *Post, basePath string) error {
	postsDir := filepath.Join(basePath, "posts")
	
	// 确保目录存在
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		return err
	}
	
	// 确保有标题和slug
	if post.Title == "" {
		post.Title = fmt.Sprintf("新文章 - %s", time.Now().Format("2006-01-02"))
	}
	
	// 如果slug为空，从标题生成
	if post.Slug == "" {
		post.Slug = GenerateSlug(post.Title)
	}
	
	// 生成文件名
	var fileName string
	
	// 检查是否已有同名文件
	files, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return err
	}
	
	// 默认使用slug作为文件名
	fileName = fmt.Sprintf("%s.md", post.Slug)
	
	// 如果是更新现有文章
	if !post.CreatedAt.IsZero() {
		post.UpdatedAt = time.Now()
		
		// 尝试查找现有文件
		fileFound := false
		for _, file := range files {
			if !file.IsDir() && (strings.HasSuffix(file.Name(), ".md") || strings.HasSuffix(file.Name(), ".markdown")) {
				fileNameWithoutExt := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				// 检查文件名是否包含slug或者是日期格式+slug
				if fileNameWithoutExt == post.Slug || strings.HasSuffix(fileNameWithoutExt, "-"+post.Slug) {
					fileName = file.Name()
					fileFound = true
					break
				}
			}
		}
		
		// 如果找不到现有文件，使用slug作为文件名
		if !fileFound {
			fileName = fmt.Sprintf("%s.md", post.Slug)
		}
	} else {
		// 新文章
		post.CreatedAt = time.Now()
		post.UpdatedAt = time.Now()
	}
	
	// 构建完整的文件路径
	filePath := filepath.Join(postsDir, fileName)
	
	// 生成稳定的ID，基于文件路径
	hasher := md5.New()
	hasher.Write([]byte(filePath))
	post.ID = hex.EncodeToString(hasher.Sum(nil))
	
	// 构建Markdown内容，包括元数据
	var contentBuilder strings.Builder
	
	// 添加元数据
	contentBuilder.WriteString(fmt.Sprintf("# %s\n\n", post.Title))
	
	if post.Author != "" {
		contentBuilder.WriteString(fmt.Sprintf("作者: %s\n", post.Author))
	}
	
	if len(post.Tags) > 0 {
		contentBuilder.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(post.Tags, ", ")))
	}
	
	if post.Cover != "" {
		contentBuilder.WriteString(fmt.Sprintf("封面: %s\n", post.Cover))
	}
	
	contentBuilder.WriteString("\n")
	
	// 添加正文内容（如果内容不是以标题开始，则保留原内容）
	content := post.Content
	if strings.HasPrefix(strings.TrimSpace(content), "# ") {
		// 内容已经包含标题，移除第一个标题行及其后的元数据
		lines := strings.Split(content, "\n")
		startIndex := 0
		for i, line := range lines {
			if strings.TrimSpace(line) == "" && i > 0 {
				startIndex = i + 1
				break
			}
		}
		if startIndex > 0 && startIndex < len(lines) {
			content = strings.Join(lines[startIndex:], "\n")
		}
	}
	
	contentBuilder.WriteString(content)
	
	// 写入文件
	return ioutil.WriteFile(filePath, []byte(contentBuilder.String()), 0644)
}

// GetAllPosts 获取所有文章的摘要信息
func GetAllPosts(basePath string) ([]PostSummary, error) {
	postsDir := filepath.Join(basePath, "posts")
	
	// 初始化一个空数组，确保即使没有文章也返回空数组而不是null
	var posts []PostSummary = []PostSummary{}
	
	// 检查目录是否存在
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		// 如果目录不存在，创建它
		if err := os.MkdirAll(postsDir, 0755); err != nil {
			return posts, err
		}
		return posts, nil
	}
	
	// 读取所有markdown文件
	files, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return posts, err
	}
	
	for _, file := range files {
		// 只处理markdown文件，忽略JSON文件和其他文件
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".md") || strings.HasSuffix(file.Name(), ".markdown")) {
			fileName := file.Name()
			fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			
			// 读取文件内容
			filePath := filepath.Join(postsDir, fileName)
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				continue
			}
			
			// 默认使用文件名作为标题和slug
			title := fileNameWithoutExt
			slug := fileNameWithoutExt // 直接使用文件名（不包含扩展名）作为slug
			createdAt := file.ModTime()
			
			// 尝试从文件名提取日期和标题（如果是日期-标题格式）
			datePattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
			matches := datePattern.FindStringSubmatch(fileNameWithoutExt)
			if len(matches) == 3 {
				// 如果文件名符合日期-标题格式
				dateStr := matches[1]
				titleFromFileName := matches[2]
				parsedDate, err := time.Parse("2006-01-02", dateStr)
				if err == nil {
					// 如果成功解析日期
					createdAt = parsedDate
					title = titleFromFileName
					slug = titleFromFileName // 使用标题部分作为slug
				}
			}
			
			// 尝试从内容中提取H1标题作为文章标题
			contentLines := strings.Split(string(content), "\n")
			for _, line := range contentLines[:min(5, len(contentLines))] {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "# ") {
					// 找到H1标题，使用它作为文章标题
					extractedTitle := strings.TrimSpace(strings.TrimPrefix(line, "# "))
					if extractedTitle != "" {
						title = extractedTitle
					}
					break
				}
			}
			
			// 生成稳定的ID，基于文件路径
			hasher := md5.New()
			hasher.Write([]byte(filePath))
			id := hex.EncodeToString(hasher.Sum(nil))
			
			// 创建文章摘要对象
			post := PostSummary{
				ID:          id,
				Title:       title,
				Slug:        slug,
				Author:      "", // 从Markdown内容中提取作者信息（如果有）
				CreatedAt:   createdAt,
				UpdatedAt:   file.ModTime(),
				IsMarkdown:  true,
				Tags:        []string{}, // 从Markdown内容中提取标签信息（如果有）
				Cover:       "", // 从Markdown内容中提取封面图片（如果有）
				IsPublished: true,
			}
			
			// 尝试从内容中提取更多元数据（作者、标签等）
			for _, line := range contentLines[:min(20, len(contentLines))] {
				line = strings.TrimSpace(line)
				
				// 提取作者信息
				if strings.HasPrefix(line, "作者:") || strings.HasPrefix(line, "Author:") {
					post.Author = strings.TrimSpace(strings.Split(line, ":")[1])
				}
				
				// 提取标签信息
				if strings.HasPrefix(line, "标签:") || strings.HasPrefix(line, "Tags:") {
					tagsPart := strings.TrimSpace(strings.Split(line, ":")[1])
					tags := strings.Split(tagsPart, ",")
					for i, tag := range tags {
						tags[i] = strings.TrimSpace(tag)
					}
					post.Tags = tags
				}
				
				// 提取封面图片
				if strings.HasPrefix(line, "封面:") || strings.HasPrefix(line, "Cover:") {
					post.Cover = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			
			posts = append(posts, post)
		}
	}
	
	return posts, nil
}

// GetPostByID 通过ID获取文章详情
func GetPostByID(id string, basePath string) (*Post, error) {
	// 由于我们现在完全依赖Markdown文件，而不是JSON文件
	// 我们需要先获取所有文章，然后根据ID查找
	posts, err := GetAllPosts(basePath)
	if err != nil {
		return nil, err
	}
	
	// 查找匹配ID的文章
	for _, postSummary := range posts {
		if postSummary.ID == id {
			// 找到匹配的文章，通过slug获取完整内容
			return GetPostBySlug(postSummary.Slug, basePath)
		}
	}
	
	return nil, fmt.Errorf("post with ID '%s' not found", id)
}

// GetPostBySlug 通过slug获取文章详情
func GetPostBySlug(slug string, basePath string) (*Post, error) {
	postsDir := filepath.Join(basePath, "posts")
	
	// 读取目录中的所有文件
	files, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}
	
	// 查找匹配的Markdown文件
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".md") || strings.HasSuffix(file.Name(), ".markdown")) {
			fileName := file.Name()
			fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			
			// 如果slug匹配文件名（不包含扩展名）或者是文件名的一部分
			if fileNameWithoutExt == slug || strings.Contains(fileNameWithoutExt, slug) {
				filePath := filepath.Join(postsDir, fileName)
				content, err := ioutil.ReadFile(filePath)
				if err != nil {
					return nil, err
				}
				
				// 默认使用文件名作为标题和slug
				title := fileNameWithoutExt
				fileSlug := fileNameWithoutExt
				createdAt := file.ModTime()
				
				// 尝试从文件名提取日期和标题（如果是日期-标题格式）
				datePattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
				matches := datePattern.FindStringSubmatch(fileNameWithoutExt)
				if len(matches) == 3 {
					// 如果文件名符合日期-标题格式
					dateStr := matches[1]
					titleFromFileName := matches[2]
					parsedDate, err := time.Parse("2006-01-02", dateStr)
					if err == nil {
						// 如果成功解析日期
						createdAt = parsedDate
						title = titleFromFileName
						fileSlug = titleFromFileName
					}
				}
				
				// 尝试从内容中提取H1标题作为文章标题
				contentLines := strings.Split(string(content), "\n")
				for _, line := range contentLines[:min(5, len(contentLines))] {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "# ") {
						// 找到H1标题，使用它作为文章标题
						extractedTitle := strings.TrimSpace(strings.TrimPrefix(line, "# "))
						if extractedTitle != "" {
							title = extractedTitle
						}
						break
					}
				}
				
				// 生成稳定的ID，基于文件路径
				hasher := md5.New()
				hasher.Write([]byte(filePath))
				id := hex.EncodeToString(hasher.Sum(nil))
				
				// 创建文章对象
				post := &Post{
					ID:          id,
					Title:       title,
					Slug:        fileSlug,
					Content:     string(content),
					Author:      "", // 从内容中提取
					CreatedAt:   createdAt,
					UpdatedAt:   file.ModTime(),
					IsMarkdown:  true,
					Tags:        []string{}, // 从内容中提取
					Cover:       "", // 从内容中提取
					IsPublished: true,
				}
				
				// 尝试从内容中提取更多元数据（作者、标签等）
				for _, line := range contentLines[:min(20, len(contentLines))] {
					line = strings.TrimSpace(line)
					
					// 提取作者信息
					if strings.HasPrefix(line, "作者:") || strings.HasPrefix(line, "Author:") {
						post.Author = strings.TrimSpace(strings.Split(line, ":")[1])
					}
					
					// 提取标签信息
					if strings.HasPrefix(line, "标签:") || strings.HasPrefix(line, "Tags:") {
						tagsPart := strings.TrimSpace(strings.Split(line, ":")[1])
						tags := strings.Split(tagsPart, ",")
						for i, tag := range tags {
							tags[i] = strings.TrimSpace(tag)
						}
						post.Tags = tags
					}
					
					// 提取封面图片
					if strings.HasPrefix(line, "封面:") || strings.HasPrefix(line, "Cover:") {
						post.Cover = strings.TrimSpace(strings.Split(line, ":")[1])
					}
				}
				
				return post, nil
			}
		}
	}
	
	return nil, fmt.Errorf("post with slug '%s' not found", slug)
}

// DeletePost 删除文章
func DeletePost(id string, basePath string) error {
	// 获取所有文章
	posts, err := GetAllPosts(basePath)
	if err != nil {
		return err
	}
	
	// 查找匹配ID的文章
	var targetSlug string
	for _, post := range posts {
		if post.ID == id {
			targetSlug = post.Slug
			break
		}
	}
	
	if targetSlug == "" {
		return fmt.Errorf("post with ID '%s' not found", id)
	}
	
	// 查找并删除对应的Markdown文件
	postsDir := filepath.Join(basePath, "posts")
	files, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".md") || strings.HasSuffix(file.Name(), ".markdown")) {
			fileNameWithoutExt := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			if strings.Contains(fileNameWithoutExt, targetSlug) {
				// 删除Markdown文件
				err := os.Remove(filepath.Join(postsDir, file.Name()))
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
	
	return fmt.Errorf("markdown file for post with ID '%s' not found", id)
}

// SearchPosts 搜索文章
func SearchPosts(query string, basePath string) ([]PostSummary, error) {
	// 初始化一个空数组，确保即使没有匹配的文章也返回空数组而不是null
	var results []PostSummary = []PostSummary{}
	
	// 获取所有文章
	posts, err := GetAllPosts(basePath)
	if err != nil {
		return results, err
	}
	
	// 如果查询为空，返回所有文章
	if query == "" {
		return posts, nil
	}
	
	// 将查询转换为小写以进行不区分大小写的搜索
	query = strings.ToLower(query)
	
	// 搜索匹配的文章
	for _, post := range posts {
		// 在标题、描述、作者和标签中搜索
		if strings.Contains(strings.ToLower(post.Title), query) ||
			strings.Contains(strings.ToLower(post.Author), query) {
			results = append(results, post)
			continue
		}
		
		// 在标签中搜索
		for _, tag := range post.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, post)
				break
			}
		}
	}
	
	return results, nil
} 