@echo off
rem =======================================================
rem  自动化前端部署脚本 by Git老师
rem =======================================================


rem --- 配置区 ---
set "projectPath=D:\Go-Project\GoZero-AI\client"
set "zipFileName=dist.zip"
set "remoteTarget=root@8.137.22.79:/opt/Go-Agent/client"
rem --- 配置区结束 ---

rem 清理屏幕
cls

echo.
echo +-------------------------------------------------------+
echo ^|            开始执行自动化部署脚本...                ^|
echo +-------------------------------------------------------+

echo.

rem 1. 进入项目目录
echo [步骤 1/4] 正在进入项目目录: %projectPath%
cd /d %projectPath%

echo.

rem 2. 执行 npm run build
echo [步骤 2/4] 正在执行 'npm run build'，请稍候...
call npm run build

rem 检查上一步命令是否成功
if %errorlevel% neq 0 (
    echo.
    echo [错误] 'npm run build' 执行失败！脚本已中止。
    echo.
    pause
    exit /b %errorlevel%
)
echo [成功] 'npm run build' 执行成功！

echo.

rem 3. 压缩 dist 文件夹
echo [步骤 3/4] 准备压缩 'dist' 文件夹...

rem 检查并删除旧的压缩包
if exist "%zipFileName%" (
    echo    - 发现旧的压缩包 '%zipFileName%'，正在删除...
    del "%zipFileName%"
)

rem 使用 PowerShell 的压缩功能 (这是Windows自带的最稳定方式)
powershell -Command "Compress-Archive -Path './dist' -DestinationPath '%zipFileName%'"

rem 检查压缩是否成功
if not exist "%zipFileName%" (
    echo.
    echo [错误] 压缩文件夹失败！脚本已中止。
    echo.
    pause
    exit /b 1
)
echo [成功] 'dist' 文件夹已成功压缩为 '%zipFileName%'！

echo.

rem 4. 上传到服务器
echo [步骤 4/4] 正在上传 '%zipFileName%' 到服务器...
call scp .\%zipFileName% "%remoteTarget%"

rem 检查上传是否成功
if %errorlevel% neq 0 (
    echo.
    echo [错误] 上传文件失败！请检查网络或 scp 命令。
    echo.
    pause
    exit /b %errorlevel%
)
echo [成功] 文件上传成功！

echo.
echo +-------------------------------------------------------+
echo ^|      恭喜！自动化部署脚本全部执行完毕！             ^|
echo +-------------------------------------------------------+
echo.

rem (可选) 自动删除本地的压缩包
rem del "%zipFileName%"

rem 脚本执行完毕，暂停一下，方便查看结果
pause
