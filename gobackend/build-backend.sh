#!/bin/bash
set -e

# Go后端构建脚本 - 用于Docker多阶段构建
# 功能：下载依赖、配置构建环境、编译Go应用、注入版本信息、版本号管理

# 递增版本号函数 - 语义化版本格式(x.y.z)
increment_version() {
  local version=$1
  # 解析版本号的各个部分
  local major=$(echo "$version" | cut -d. -f1)
  local minor=$(echo "$version" | cut -d. -f2)
  local patch=$(echo "$version" | cut -d. -f3)
  
  # 递增修订号
  patch=$((patch + 1))
  
  # 返回新的版本号
  echo "${major}.${minor}.${patch}"
}

# 设置默认参数
build_mode="production"
version="0.0.1"
version_file=""

# 尝试从版本文件读取版本号
if [ -f "../../version.txt" ]; then
  version=$(cat "../../version.txt")
  version_file="../../version.txt"
  echo "从 ../../version.txt 读取版本号: $version"
elif [ -f "./version.txt" ]; then
  version=$(cat "./version.txt")
  version_file="./version.txt"
  echo "从 ./version.txt 读取版本号: $version"
else
  # 未找到版本文件，使用默认版本并设置创建路径
  echo "警告: 未找到版本文件，使用默认版本号"
  version_file="./version.txt"
fi

# 递增版本号，为本次构建准备新版本
new_version=$(increment_version "$version")
echo "本次构建将使用新版本号: $new_version"

# 设置输出路径为../release/${version}
output_path="../release/${version}"
echo "===== 开始Go后端构建流程 ====="
echo "构建输出目录: $output_path"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --dev)
      build_mode="development"
      shift
      ;;
    --output)
      output_binary="$2"
      shift 2
      ;;
    *)
      echo "未知参数: $1"
      exit 1
      ;;
  esac
done

# 创建输出目录
output_binary="${output_path}/app"
output_dir=$(dirname "$output_binary")
mkdir -p "$output_dir"

# 设置Go环境变量
export CGO_ENABLED=1  # SQLite需要CGO
export GOOS=linux

# 下载依赖
echo "===== 下载Go依赖 ====="
go mod tidy
go mod download

# 根据构建模式设置构建标志
if [ "$build_mode" = "production" ]; then
  echo "===== 生产模式构建 ====="
  # 注入版本号信息到二进制文件中
  build_flags="-ldflags='-s -w -X 'github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/config.Version=$new_version''"
else
  echo "===== 开发模式构建 ====="
  # 开发模式也注入版本号信息
  build_flags="-tags=debug -ldflags='-X 'github.com/aspnmy/GoComicMosaic-gobackend/gobackend/internal/config.Version=$new_version''"
fi

# 编译应用
echo "===== 编译Go应用 ====="
echo "输出文件: $output_binary"

# 执行构建命令
echo "使用版本号: $new_version 构建应用"
eval "go build $build_flags -o '$output_binary' ./cmd/api"

# 复制webp工具（如果需要）
if [ -d "./cmd/webp" ]; then
  echo "===== 编译WebP工具 ====="
  webp_binary="${output_binary}_webp"
  # 为WebP工具也注入版本号
  eval "go build $build_flags -o '$webp_binary' ./cmd/webp"
  echo "WebP工具已编译: $webp_binary"
fi

# 编译AVIF工具（如果需要）
if [ -d "./cmd/avif" ]; then
  echo "===== 编译AVIF工具 ====="
  avif_binary="${output_binary}_avif"
  # 为AVIF工具也注入版本号
  eval "go build $build_flags -o '$avif_binary' ./cmd/avif"
  echo "AVIF工具已编译: $avif_binary"
fi

# 显示编译信息
echo "===== 构建信息 ====="
echo "Go版本: $(go version)"
echo "构建模式: $build_mode"
echo "版本号: $new_version"
echo "文件大小: $(du -h "$output_binary" | cut -f1)"
echo "文件权限: $(ls -la "$output_binary" | awk '{print $1}')"

# 设置执行权限
chmod +x "$output_binary"

# 更新版本文件
if [ -n "$version_file" ]; then
  echo "===== 更新版本文件 ====="
  echo "将版本号从 $version 更新为 $new_version 到文件: $version_file"
  echo "$new_version" > "$version_file"
  echo "版本文件更新成功!"
fi

echo "===== Go后端构建完成 ====="
echo "可执行文件位于: $output_binary"
echo "使用版本号: $new_version"
echo "===== 构建流程结束 ====="