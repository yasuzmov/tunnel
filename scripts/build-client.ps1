# 编译 Flutter Windows 客户端。
# MSBuild 无法正确读取含非 ASCII 字符的路径，因此遇到中文目录时临时 subst 到盘符再编。
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$client = Join-Path $root "client"

function Test-AsciiPath([string]$Path) {
    return $Path -match '^[\x00-\x7F]+$'
}

$work = $client
$mapped = $null
if (-not (Test-AsciiPath $client)) {
    foreach ($letter in @("T", "U", "V", "W", "X")) {
        $candidate = "${letter}:"
        if (-not (Test-Path $candidate)) {
            subst $candidate $root
            $mapped = $candidate
            $work = Join-Path $candidate "client"
            break
        }
    }
    if (-not $mapped) {
        throw "没有可用盘符用于绕过中文路径，请把仓库放到纯英文目录后重试"
    }
}

try {
    Set-Location $work
    flutter pub get
    flutter build windows --release
    Write-Host "产物目录: $client\build\windows\x64\runner\Release"
} finally {
    if ($mapped) {
        subst $mapped /d
    }
}
