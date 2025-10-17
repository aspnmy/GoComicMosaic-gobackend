# GoComicMosaic - Go 后端

这是 GoComicMosaic 影视资源共建平台的后端项目，使用 Golang 实现。该平台允许用户自由提交资源信息（标题、类型、简介、图片、资源链接），像马赛克一样，由多方贡献拼凑成完整资源。

## 功能特点
- **资源管理**：支持资源的新增、修改、查询、删除
- **资源审核**：支持资源的审核流程，包括审核、驳回
- **图片处理**：图片自动转WebP格式，支持批量处理
- **CORS代理**：内置代理功能，解决前端跨域问题
- **网站配置**：支持自定义网站信息、图标等
- **TMDB集成**：支持配置TMDB API密钥，提供TMDB数据访问接口

## 项目结构

```
gobackend/
├── README.md                             # 项目说明文档
├── cmd/                                  # 命令目录
│   ├── api/                              # API应用
│   │   └── main.go                       # API主程序
│   ├── diagnostic/                       # 诊断工具
│   │   └── main.go                       # 诊断主程序
│   ├── test/                             # 测试工具
│   │   ├── sqlite.go                     # SQLite测试
│   │   └── webp_tool_test.go             # WebP工具测试
│   └── webp/                             # WebP转换工具
│       └── main.go                       # WebP转换主程序
├── internal/                             # 内部包
│   ├── auth/                             # 认证工具
│   │   └── auth.go                       # JWT认证相关功能
│   ├── config/                           # 配置处理
│   │   └── config.go                     # 配置加载与管理
│   ├── handlers/                         # HTTP处理器
│   │   ├── auth_handlers.go              # 认证相关处理器
│   │   ├── middleware.go                 # 中间件
│   │   ├── proxy_handler.go              # CORS代理处理器
│   │   ├── resource_approval.go          # 资源审批处理器
│   │   ├── resource_handlers.go          # 资源基本操作处理器
│   │   ├── routes.go                     # 路由定义
│   │   ├── site_settings_handlers.go     # 网站设置处理器
│   │   ├── tmdb_handlers.go              # TMDB API处理器
│   │   ├── tmdb_season_handlers.go       # TMDB季集处理器
│   │   └── upload_handlers.go            # 文件上传处理器
│   ├── models/                           # 数据模型
│   │   ├── database.go                   # 数据库连接和初始化
│   │   └── models.go                     # 数据模型定义
│   └── utils/                            # 工具函数
│       ├── config.go                     # 配置工具
│       ├── image_utils.go                # 图像处理工具
│       ├── tmdb.go                       # TMDB API工具
│       ├── tmdb_seasons.go               # TMDB季集工具
│       └── webp_utils.go                 # WebP转换工具
├── go.mod                                # Go模块定义
└── go.sum                                # Go依赖版本锁定
```

## API文档

### 资源管理API

- `GET /api/resources` - 获取资源列表
- `GET /api/resources/:id` - 获取资源详情
- `POST /api/resources` - 创建新资源
- `PUT /api/resources/:id` - 更新资源
- `DELETE /api/resources/:id` - 删除资源
- `GET /api/resources/:id/supplements` - 获取资源补充内容
- `POST /api/resources/:id/supplements` - 添加资源补充内容

### 资源审核API

- `GET /api/admin/approval` - 获取待审核资源列表
- `POST /api/admin/approval/:id/approve` - 审核通过资源
- `POST /api/admin/approval/:id/reject` - 驳回资源
- `POST /api/admin/approval/:id/supplement/approve` - 审核通过资源补充
- `POST /api/admin/approval/:id/supplement/reject` - 驳回资源补充

### 用户认证API

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `GET /api/auth/verify` - 验证用户Token
- `POST /api/auth/logout` - 用户登出

### 文件上传API

- `POST /api/upload` - 上传文件
- `GET /api/files/:filename` - 获取文件

### 网站配置API

- `GET /api/site/settings` - 获取网站配置
- `POST /api/admin/site/settings` - 更新网站配置

### TMDB API相关

- `GET /api/settings/tmdb_status` - 获取当前TMDB功能状态
- `GET /api/admin/tmdb/config` - 获取当前TMDB配置
- `PUT /api/admin/tmdb/config` - 更新TMDB配置信息
- `GET /api/tmdb/search` - 搜索TMDB影视作品
- `POST /api/tmdb/create` - 从TMDB信息创建资源
- `GET /api/tmdb/seasons/:series_id` - 获取TMDB电视剧的所有季信息
- `GET /api/tmdb/seasons/:series_id/:season_number` - 获取TMDB电视剧指定季的所有集信息
- `GET /api/tmdb/seasons/:series_id/:season_number/:episode_number/images` - 获取指定剧集的剧照
- `GET /api/tmdb/seasons/:series_id/:season_number/:episode_number/credits` - 获取指定剧集的演员信息
- `GET /api/tmdb/resource/:tmdb_id` - 通过TMDB ID获取本地资源
- `PUT /api/tmdb/update-resource-id/:id/:tmdb_id` - 更新资源的TMDB ID

### 代理API

- `GET /api/proxy` - 代理请求，用于解决跨域问题

## 安装与运行

### 环境要求

- Go 1.22+
- libwebp (用于WebP图片转换)

### 配置

主要配置项：

```
PORT=8000
ASSETS_PATH="../data/assets" # 默认 '../assets'
DB_PATH="../data/resource_hub.db" # 默认gobackend目录下
TMDB_API_KEY=your_tmdb_api_key # 此处可选，也可通过管理界面配置
```

### 模块路径

该项目的模块路径已配置为：
```
github.com/aspnmy/GoComicMosaic-gobackend
```

每次构建时会自动从GitHub获取最新代码。

### 运行

开发测试运行（默认为Release模式）
```
go run cmd/api/main.go
```

开发调试运行（启用Debug模式）
```
GIN_MODE=debug go run cmd/api/main.go
```

编译
```
go build -ldflags="-w -s" -o app cmd/api/main.go
```

### 交叉编译

Mac 编译 Linux 版本：
```bash
sudo chown -R $(whoami):admin /usr/local/Homebrew
chmod u+w /usr/local/Homebrew
brew install x86_64-linux-gnu-binutils
brew tap messense/macos-cross-toolchains
brew install x86_64-unknown-linux-gnu

CC=/usr/local/Cellar/x86_64-unknown-linux-gnu/13.3.0.reinstall/bin/x86_64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags "sqlite_static" -ldflags="-w -s" -o app cmd/api/main.go
```

### 使用构建脚本

项目提供了便捷的构建脚本：
```bash
./build-backend.sh
```

该脚本会自动处理依赖安装、版本注入和多版本构建（标准版、WebP版、AVIF版）。