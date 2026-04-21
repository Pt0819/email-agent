@echo off
chcp 65001 >nul
echo ============================================
echo  Email Agent 系统启动脚本
echo ============================================
echo.

:: 检查MySQL
echo [1/4] 检查MySQL服务...
mysql -u root -p123456 -e "SELECT 1" >nul 2>&1
if errorlevel 1 (
    echo [警告] MySQL未运行或连接失败
    echo 请确保MySQL服务已启动
    echo.
) else (
    echo [OK] MySQL连接正常
    echo.
)

:: 检查Redis
echo [2/4] 检查Redis服务...
redis-cli ping >nul 2>&1
if errorlevel 1 (
    echo [警告] Redis未运行
    echo 请确保Redis服务已启动
    echo.
) else (
    echo [OK] Redis连接正常
    echo.
)

:: 启动Agent服务
echo [3/4] 启动Agent服务 (端口 9111)...
cd email-agent
start "Email Agent" cmd /c "python app/main.py"
cd ..
timeout /t 3 /nobreak >nul
echo [OK] Agent服务启动中...
echo.

:: 启动后端服务
echo [4/4] 启动后端服务 (端口 8080)...
cd email-backend
start "Email Backend" cmd /c "server.exe"
cd ..
timeout /t 3 /nobreak >nul
echo [OK] 后端服务启动中...
echo.

:: 检查服务状态
echo ============================================
echo  等待服务就绪...
echo ============================================
timeout /t 5 /nobreak >nul

echo.
echo 检查Agent服务...
curl -s http://localhost:9111/api/v1/health >nul 2>&1
if errorlevel 1 (
    echo [失败] Agent服务未就绪
) else (
    echo [OK] Agent服务运行正常
)

echo.
echo 检查后端服务...
curl -s http://localhost:8080/api/v1/health >nul 2>&1
if errorlevel 1 (
    echo [失败] 后端服务未就绪
) else (
    echo [OK] 后端服务运行正常
)

echo.
echo ============================================
echo  服务启动完成
echo ============================================
echo.
echo 访问地址:
echo   前端: npm run dev (在email-web目录)
echo   后端API: http://localhost:8080
echo   Agent API: http://localhost:9111
echo   Agent文档: http://localhost:9111/docs
echo.
echo 按任意键退出...
pause >nul
