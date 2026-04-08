# 本地开发环境配置指南

> 本指南帮助你搭建邮件分类Agent系统的本地开发环境

---

## 1. 环境要求总览

| 组件 | 版本要求 | 用途 | 必需性 |
|------|---------|------|--------|
| **Python** | 3.11+ | Agent端开发 | ✅ 必需 |
| **Go** | 1.21+ | 服务端开发 | ✅ 必需 |
| **Node.js** | 18+ | Web前端开发 | ✅ 必需 |
| **MySQL** | 8.0 | 数据存储 | ✅ 必需 |
| **Redis** | 7+ | 缓存/消息队列 | ✅ 必需 |
| **Docker** | 最新版 | 容器化部署 | ⚡ 推荐 |

---

## 2. Python 环境搭建

### 2.1 Windows 安装 Python 3.11

#### 方式一：官方安装包（推荐）

1. **下载安装包**
   ```
   访问: https://www.python.org/downloads/windows/
   选择: Python 3.11.x (Windows installer 64-bit)
   ```

2. **安装步骤**
   - 运行下载的安装包
   - **重要**：勾选 `Add Python to PATH`
   - 点击 `Install Now` 或自定义安装路径

3. **验证安装**
   ```powershell
   # 打开 PowerShell 或 CMD
   python --version
   # 输出: Python 3.11.x
   
   pip --version
   # 输出: pip 23.x.x from ...
   ```

#### 方式二：使用 winget 安装

```powershell
# Windows 11 自带 winget
winget install Python.Python.3.11
```

#### 方式三：使用 Conda（适合多项目管理）

1. **下载 Miniconda**
   ```
   https://docs.conda.io/en/latest/miniconda.html
   ```

2. **安装后创建项目专用环境**
   ```powershell
   # 创建虚拟环境
   conda create -n email-agent python=3.11
   
   # 激活环境
   conda activate email-agent
   
   # 验证
   python --version
   ```

### 2.2 配置 Python 虚拟环境

推荐为项目创建独立的虚拟环境，避免依赖冲突：

```powershell
# 进入项目目录
cd "D:\claude project\mail-agent\email-agent"

# 创建虚拟环境 (使用 venv)
python -m venv venv

# 激活虚拟环境
# Windows PowerShell:
.\venv\Scripts\Activate.ps1

# Windows CMD:
.\venv\Scripts\activate.bat

# 激活后，命令行会显示 (venv) 前缀
# (venv) D:\claude project\mail-agent\email-agent>
```

### 2.3 安装 Python 依赖

创建 `requirements.txt` 文件：

```txt
# email-agent/requirements.txt

# Web框架
fastapi==0.110.0
uvicorn[standard]==0.27.1

# LLM框架
langchain==0.1.16
langchain-core==0.1.42
langchain-openai==0.0.8
langchain-community==0.0.32

# 向量存储
chromadb==0.4.22

# Pydantic数据验证
pydantic==2.6.1

# Redis客户端
redis==5.0.1

# HTTP客户端
httpx==0.26.0
aiohttp==3.9.3

# 配置管理
pyyaml==6.0.1

# 日志
loguru==0.7.2

# 测试
pytest==8.0.0
pytest-asyncio==0.23.4
```

安装依赖：

```powershell
# 确保虚拟环境已激活
pip install -r requirements.txt

# 验证安装
pip list
```

### 2.4 Python IDE 推荐

- **VS Code** + Python扩展（免费，推荐）
- **PyCharm** Community Edition（免费）
- **Cursor**（AI辅助编辑器）

VS Code Python扩展安装：
```
扩展ID: ms-python.python
扩展ID: ms-python.vscode-pylance (类型检查)
```

---

## 3. Go 环境搭建

### 3.1 Windows 安装 Go 1.21+

#### 方式一：官方安装包

1. **下载安装包**
   ```
   https://go.dev/dl/
   选择: go1.21.x.windows-amd64.msi
   ```

2. **安装后验证**
   ```powershell
   go version
   # 输出: go version go1.21.x windows/amd64
   
   go env GOPATH
   # 显示Go包路径
   ```

#### 方式二：使用 winget

```powershell
winget install GoLang.Go
```

### 3.2 Go 环境配置

```powershell
# 设置环境变量（可选，默认已配置）
# GOPATH - Go项目和工作空间路径
$env:GOPATH = "D:\GoProjects"

# GOMODCACHE - 模块缓存路径
$env:GOMODCACHE = "D:\GoProjects\pkg\mod"

# 开启Go模块模式
$env:GO111MODULE = "on"
```

### 3.3 Go 项目初始化

```powershell
# 进入服务端目录
cd "D:\claude project\mail-agent\email-backend"

# 初始化Go模块
go mod init email-backend

# 下载依赖（创建go.mod后会自动处理）
go mod tidy
```

### 3.4 Go IDE 推荐

- **VS Code** + Go扩展
- **GoLand**（JetBrains，收费）

VS Code Go扩展：
```
扩展ID: golang.Go
```

---

## 4. Node.js 环境搭建

### 4.1 Windows 安装 Node.js 18+

#### 方式一：官方安装包

1. **下载安装包**
   ```
   https://nodejs.org/
   选择: 18.x LTS (Windows Installer 64-bit)
   ```

2. **安装后验证**
   ```powershell
   node --version
   # 输出: v18.x.x
   
   npm --version
   # 输出: 9.x.x
   ```

#### 方式二：使用 winget

```powershell
winget install OpenJS.NodeJS.LTS
```

#### 方式三：使用 nvm-windows（推荐，可管理多版本）

1. **下载 nvm-windows**
   ```
   https://github.com/coreybutler/nvm-windows/releases
   下载 nvm-setup.exe
   ```

2. **安装和使用**
   ```powershell
   # 安装Node.js 18
   nvm install 18
   
   # 使用Node.js 18
   nvm use 18
   
   # 验证
   node --version
   ```

### 4.2 Node.js 项目初始化

```powershell
# 进入前端目录
cd "D:\claude project\mail-agent\email-web"

# 使用 Vite 创建 React 项目（首次）
npm create vite@latest . -- --template react-ts

# 安装依赖
npm install

# 安装额外依赖
npm install tailwindcss postcss autoprefixer
npm install @tanstack/react-query axios
npm install lucide-react class-variance-authority clsx tailwind-merge
```

### 4.3 Node.js IDE 推荐

- **VS Code**（内置JS/TS支持）

---

## 5. 数据库环境

### 5.1 MySQL 8.0 安装

#### 方式一：官方安装包

```
下载: https://dev.mysql.com/downloads/installer/
选择: MySQL Community Server 8.0
```

#### 方式二：Docker（推荐）

```powershell
# 拉取MySQL镜像
docker pull mysql:8.0

# 启动MySQL容器
docker run -d \
  --name email-mysql \
  -e MYSQL_ROOT_PASSWORD=your_password \
  -e MYSQL_DATABASE=email_system \
  -p 3306:3306 \
  mysql:8.0

# 连接MySQL
docker exec -it email-mysql mysql -uroot -p
```

### 5.2 Redis 安装

#### 方式一：Windows版Redis

```
下载: https://github.com/microsoftarchive/redis/releases
选择: Redis-x64-7.x.msi
```

#### 方式二：Docker（推荐）

```powershell
# 拉取Redis镜像
docker pull redis:7-alpine

# 启动Redis容器
docker run -d \
  --name email-redis \
  -p 6379:6379 \
  redis:7-alpine

# 测试连接
docker exec -it email-redis redis-cli ping
# 输出: PONG
```

### 5.3 ChromaDB 安装

使用Docker运行：

```powershell
# 拉取ChromaDB镜像
docker pull ghcr.io/chroma-core/chroma:latest

# 启动ChromaDB容器
docker run -d \
  --name email-chroma \
  -p 8000:8000 \
  ghcr.io/chroma-core/chroma:latest
```

---

## 6. Docker 环境（推荐）

使用Docker可以简化本地环境配置，避免安装多个依赖。

### 6.1 Windows 安装 Docker Desktop

1. **下载 Docker Desktop**
   ```
   https://www.docker.com/products/docker-desktop/
   ```

2. **安装要求**
   - Windows 11 或 Windows 10 (版本 2004+)
   - 启用 WSL 2

3. **安装步骤**
   - 运行安装程序
   - 安装完成后重启电脑
   - 启动 Docker Desktop

4. **验证安装**
   ```powershell
   docker --version
   # 输出: Docker version 24.x.x
   
   docker-compose --version
   # 输出: Docker Compose version v2.x.x
   
   # 测试运行
   docker run hello-world
   ```

### 6.2 使用 Docker 开发

开发时可以只运行基础设施（数据库），本地运行代码：

```powershell
# 只启动数据库服务
docker-compose up mysql redis chroma -d

# 本地运行服务端代码
cd email-backend && go run cmd/server/main.go

# 本地运行Agent代码
cd email-agent && python -m uvicorn app.main:app --reload

# 本地运行前端
cd email-web && npm run dev
```

---

## 7. 开发工具推荐

### 7.1 IDE 选择

| 开发内容 | 推荐IDE | 扩展 |
|---------|---------|------|
| Python Agent | VS Code | Python, Pylance |
| Go Server | VS Code | Go |
| React Web | VS Code | ESLint, Prettier |
| 全栈开发 | VS Code | 以上全部 |

### 7.2 VS Code 推荐扩展

```json
// .vscode/extensions.json
{
  "recommendations": [
    // Python
    "ms-python.python",
    "ms-python.vscode-pylance",
    
    // Go
    "golang.Go",
    
    // JavaScript/TypeScript
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    
    // 其他
    "editorconfig.editorconfig",
    "eamodio.gitlens",
    "docker.docker",
    "redhat.vscode-yaml"
  ]
}
```

### 7.3 VS Code 调试配置

```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    // Python Agent 调试
    {
      "name": "Python: Agent",
      "type": "python",
      "request": "launch",
      "module": "uvicorn",
      "args": [
        "app.main:app",
        "--reload",
        "--port", "8001"
      ],
      "cwd": "${workspaceFolder}/email-agent",
      "env": {
        "PYTHONPATH": "${workspaceFolder}/email-agent"
      }
    },
    
    // Go Server 调试
    {
      "name": "Go: Server",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/email-backend/cmd/server",
      "cwd": "${workspaceFolder}/email-backend",
      "env": {
        "DB_HOST": "localhost",
        "DB_PORT": "3306"
      }
    },
    
    // React Web 调试
    {
      "name": "Web: Chrome",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:5173",
      "webRoot": "${workspaceFolder}/email-web/src"
    }
  ]
}
```

---

## 8. 环境验证清单

安装完成后，运行以下命令验证环境：

```powershell
# 创建验证脚本 verify-env.ps1

Write-Host "=== 环境验证 ===" -ForegroundColor Green

# Python
Write-Host "Python: " -NoNewline
python --version

# Go
Write-Host "Go: " -NoNewline
go version

# Node.js
Write-Host "Node.js: " -NoNewline
node --version

Write-Host "npm: " -NoNewline
npm --version

# Docker
Write-Host "Docker: " -NoNewline
docker --version

Write-Host "Docker Compose: " -NoNewline
docker-compose --version

# MySQL (Docker)
Write-Host "MySQL: " -NoNewline
docker ps --filter name=mysql --format "{{.Status}}"

# Redis (Docker)
Write-Host "Redis: " -NoNewline
docker ps --filter name=redis --format "{{.Status}}"

Write-Host "=== 验证完成 ===" -ForegroundColor Green
```

---

## 9. 快速开始步骤

### 步骤1：安装基础环境

```powershell
# 1. 安装 Python 3.11
winget install Python.Python.3.11

# 2. 安装 Go 1.21
winget install GoLang.Go

# 3. 安装 Node.js 18
winget install OpenJS.NodeJS.LTS

# 4. 安装 Docker Desktop
# 手动下载安装: https://docker.com/products/docker-desktop
```

### 步骤2：启动数据库服务

```powershell
# 创建 docker-compose.dev.yml 仅运行数据库
# docker-compose -f docker-compose.dev.yml up -d

docker run -d --name email-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=email_system -p 3306:3306 mysql:8.0
docker run -d --name email-redis -p 6379:6379 redis:7-alpine
docker run -d --name email-chroma -p 8000:8000 ghcr.io/chroma-core/chroma:latest
```

### 步骤3：初始化各端项目

```powershell
# Agent端
cd email-agent
python -m venv venv
.\venv\Scripts\Activate.ps1
pip install fastapi uvicorn langchain chromadb redis pyyaml loguru

# 服务端
cd email-backend
go mod init email-backend
go get github.com/gin-gonic/gin gorm.io/gorm gorm.io/driver/mysql github.com/redis/go-redis/v9

# 前端
cd email-web
npm create vite@latest . -- --template react-ts
npm install
```

---

## 10. 常见问题

### Q1: Python pip 安装速度慢？

```powershell
# 使用国内镜像
pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# 配置永久镜像
pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
```

### Q2: Go 模块下载慢？

```powershell
# 设置代理
$env:GOPROXY = "https://goproxy.cn,direct"
go mod tidy
```

### Q3: npm 安装慢？

```powershell
# 使用淘宝镜像
npm config set registry https://registry.npmmirror.com

# 或使用 pnpm（更快）
npm install -g pnpm
pnpm install
```

### Q4: Docker Desktop 启动慢？

- 确保WSL 2已启用
- 在Docker Desktop设置中减少内存限制
- 使用 `docker-compose` 替代手动docker run

---

*最后更新: 2026-04-07*