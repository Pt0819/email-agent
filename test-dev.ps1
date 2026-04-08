# Email Agent System - Development Self-Test Script
# 用于验证三端项目是否正确初始化

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Email Agent System - Self Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$PROJECT_ROOT = "D:\claude project\mail-agent"
$FAILED = 0
$PASSED = 0

function Test-Check($condition, $name, $detail = "") {
    if ($condition) {
        Write-Host "[PASS] $name" -ForegroundColor Green
        $script:PASSED++
    } else {
        Write-Host "[FAIL] $name" -ForegroundColor Red
        if ($detail) {
            Write-Host "       $detail" -ForegroundColor Yellow
        }
        $script:FAILED++
    }
}

# ============================================
# 1. Go Backend Tests
# ============================================
Write-Host "`n[Go Backend]" -ForegroundColor Yellow

$goBackend = "$PROJECT_ROOT\email-backend"

# Check directory structure
Test-Check (Test-Path "$goBackend\go.mod") "go.mod exists"
Test-Check (Test-Path "$goBackend\go.sum") "go.sum exists"
Test-Check (Test-Path "$goBackend\cmd\server\main.go") "main.go exists"
Test-Check (Test-Path "$goBackend\config\config.yaml") "config.yaml exists"
Test-Check (Test-Path "$goBackend\internal\pkg\config\config.go") "config.go exists"
Test-Check (Test-Path "$goBackend\internal\pkg\response\response.go") "response.go exists"

# Check compiled binary
$binaryExists = Test-Path "$goBackend\server.exe"
Test-Check $binaryExists "server.exe compiled"

if ($binaryExists) {
    Write-Host "       Binary size: $((Get-Item "$goBackend\server.exe").Length / 1MB) MB" -ForegroundColor Gray
}

# ============================================
# 2. Python Agent Tests
# ============================================
Write-Host "`n[Python Agent]" -ForegroundColor Yellow

$pythonAgent = "$PROJECT_ROOT\email-agent"

Test-Check (Test-Path "$pythonAgent\requirements.txt") "requirements.txt exists"
Test-Check (Test-Path "$pythonAgent\app\main.py") "main.py exists"
Test-Check (Test-Path "$pythonAgent\app\config.py") "config.py exists"
Test-Check (Test-Path "$pythonAgent\app\schemas\__init__.py") "schemas/__init__.py exists"
Test-Check (Test-Path "$pythonAgent\config\config.yaml") "config.yaml exists"

# Check Python syntax
Write-Host "`n       Checking Python syntax..." -ForegroundColor Gray
$pythonCmd = "D:\python\py3.11\python.exe"
try {
    $syntaxOutput = & $pythonCmd -m py_compile "$pythonAgent\app\main.py" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Test-Check $true "main.py syntax valid"
    } else {
        Test-Check $false "main.py syntax valid" $syntaxOutput
    }
} catch {
    Test-Check $false "Python interpreter accessible" $_.Exception.Message
}

# ============================================
# 3. React Web Tests
# ============================================
Write-Host "`n[React Web]" -ForegroundColor Yellow

$reactWeb = "$PROJECT_ROOT\email-web"

Test-Check (Test-Path "$reactWeb\package.json") "package.json exists"
Test-Check (Test-Path "$reactWeb\src\App.tsx") "App.tsx exists"
Test-Check (Test-Path "$reactWeb\src\api\client.ts") "api/client.ts exists"
Test-Check (Test-Path "$reactWeb\src\pages\EmailList.tsx") "pages/EmailList.tsx exists"
Test-Check (Test-Path "$reactWeb\tailwind.config.js") "tailwind.config.js exists"

# Check build output
$distExists = Test-Path "$reactWeb\dist"
Test-Check $distExists "dist folder exists (built)"

if ($distExists) {
    $htmlExists = Test-Path "$reactWeb\dist\index.html"
    Test-Check $htmlExists "index.html generated"
}

# ============================================
# 4. Common Directory Tests
# ============================================
Write-Host "`n[Common Directories]" -ForegroundColor Yellow

Test-Check (Test-Path "$PROJECT_ROOT\docs") "docs folder exists"
Test-Check (Test-Path "$PROJECT_ROOT\sql") "sql folder exists"
Test-Check (Test-Path "$PROJECT_ROOT\docker") "docker folder exists"
Test-Check (Test-Path "$PROJECT_ROOT\configs") "configs folder exists"

# ============================================
# Summary
# ============================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  PASSED: $PASSED" -ForegroundColor Green
Write-Host "  FAILED: $FAILED" -ForegroundColor $(if ($FAILED -gt 0) { "Red" } else { "Green" })

if ($FAILED -eq 0) {
    Write-Host "`n  All tests passed! System ready for development." -ForegroundColor Green
    Write-Host "`n  Next steps:" -ForegroundColor Cyan
    Write-Host "    1. Start MySQL: docker run -d --name email-mysql -p 3306:3306 mysql:8.0" -ForegroundColor Gray
    Write-Host "    2. Start Redis: docker run -d --name email-redis -p 6379:6379 redis:7-alpine" -ForegroundColor Gray
    Write-Host "    3. Run Go Backend: cd email-backend && go run cmd/server/main.go" -ForegroundColor Gray
    Write-Host "    4. Run Python Agent: cd email-agent && python app/main.py" -ForegroundColor Gray
    Write-Host "    5. Run Web Dev: cd email-web && npm run dev" -ForegroundColor Gray
} else {
    Write-Host "`n  Some tests failed. Please check the errors above." -ForegroundColor Red
}

Write-Host ""
exit $FAILED