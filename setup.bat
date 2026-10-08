:: curl -sL "https://monkeforge.org/cli/setup.bat | cmd

:: Output dir in AppData
set "OutputDir=%LOCALAPPDATA%\Programs\MonkeForgeCLI"

:: Create dir
if not exist "%OutputDir%" mkdir "%OutputDir%"

:: Download to MonkeForgeCLI\mforge.exe
curl -L "https://github.com/sirkingbinx/monkeforge-cli/releases/latest/download/mforge-win-x64.exe" -o "%OutputDir%\mforge.exe"

:: Add it to path
setx PATH "%PATH%;%OutputDir%"

:: Refresh PATH for current shell
for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v PATH') do set "PATH=%%B"

:: We're done
echo "`mforge` command is installed"