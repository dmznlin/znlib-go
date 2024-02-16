:: 切换到当前路径
cd %~dp0

SET dist=dist\terminal
:: 创建目录
md %dist%

:: 复制有效文件
copy *.html %dist% /y
copy *.js %dist% /y

for /d %%i in (.\*) do (
  xcopy %%i\*.css %dist%    /s/y/i
  xcopy %%i\*.html %dist% /s/y/i
  xcopy %%i\*.js %dist% /s/y/i
)

:: 显示结果
pause