Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envFile = Join-Path $scriptDir ".env"
$generatedConfigFile = Join-Path $scriptDir ".generated\dotblue\config.yaml"

function Fail([string]$Message) {
  throw $Message
}

function Load-EnvFile([string]$Path) {
  $map = @{}
  foreach ($line in Get-Content -Path $Path) {
    $trimmed = $line.TrimEnd("`r")
    if ([string]::IsNullOrWhiteSpace($trimmed) -or $trimmed.TrimStart().StartsWith("#")) {
      continue
    }
    $idx = $trimmed.IndexOf("=")
    if ($idx -lt 0) {
      continue
    }
    $key = $trimmed.Substring(0, $idx)
    $value = $trimmed.Substring($idx + 1)
    $map[$key] = $value
  }
  return $map
}

function Write-Section([string]$Title) {
  Write-Host ""
  Write-Host "== $Title =="
}

function Write-Check([string]$Label, [string]$Expected, [string]$Actual) {
  if ($Expected -eq $Actual) {
    Write-Host "[ok] $Label`: $Actual"
  }
  else {
    Write-Host "[warn] $Label`: expected '$Expected', got '$Actual'"
  }
}

function Write-Value([string]$Label, [string]$Value) {
  if ([string]::IsNullOrWhiteSpace($Value)) {
    Write-Host "[warn] $Label`: empty"
  }
  else {
    Write-Host "[ok] $Label`: $Value"
  }
}

function Invoke-Docker {
  param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Args
  )

  $docker = Get-Command docker -ErrorAction SilentlyContinue
  if ($docker) {
    return & $docker.Source @Args
  }

  $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
  if (-not $wsl) {
    Fail "docker command is unavailable and wsl.exe fallback was not found"
  }

  return & $wsl.Source -e docker @Args
}

function Read-GeneratedEngineValue([string]$Key) {
  if (-not (Test-Path -Path $generatedConfigFile)) {
    return ""
  }

  foreach ($line in Get-Content -Path $generatedConfigFile) {
    $trimmed = $line.Trim()
    if ($trimmed.StartsWith("${Key}:")) {
      $value = $trimmed.Substring($Key.Length + 1).Trim()
      return $value.Trim('"')
    }
  }
  return ""
}

if (-not (Test-Path -Path $envFile)) {
  Fail "missing $envFile. Copy .env.example to .env first."
}

Push-Location $scriptDir
try {
  $envMap = Load-EnvFile $envFile

  $backendCid = Invoke-Docker compose ps -q dotblue
  if ([string]::IsNullOrWhiteSpace($backendCid)) {
    Fail "dotblue container is not created"
  }
  $backendCid = $backendCid.Trim()

  $backendUser = Invoke-Docker inspect $backendCid --format '{{.Config.User}}'
  $backendState = Invoke-Docker inspect $backendCid --format '{{.State.Status}}'
  $backendNetworks = Invoke-Docker inspect $backendCid --format '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}} {{end}}'
  $backendSocketMeta = Invoke-Docker exec $backendCid stat -c "%g %a %U %G" /var/run/docker.sock 2>$null
  $backendIdentity = Invoke-Docker exec $backendCid id 2>$null

  $hostSocketGid = ""
  $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
  if ($wsl) {
    $hostSocketGid = & $wsl.Source -e sh -lc "if [ -S /var/run/docker.sock ]; then stat -c %g /var/run/docker.sock; fi" 2>$null
    if ($LASTEXITCODE -ne 0) {
      $hostSocketGid = ""
    }
  }

  $effectiveDockerNetwork = Read-GeneratedEngineValue "dockerNetwork"

  Write-Section "Config"
  Write-Value "runtime mode" $envMap["DOTBLUE_ENGINE_RUNTIME_MODE"]
  Write-Value "endpoint mode" $envMap["DOTBLUE_ENGINE_ENDPOINT_MODE"]
  Write-Value "docker endpoint" $envMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"]
  Write-Value "docker network (effective)" ($(if ($effectiveDockerNetwork) { $effectiveDockerNetwork } else { $envMap["DOTBLUE_ENGINE_DOCKER_NETWORK"] }))
  Write-Value "host data path" $envMap["DOTBLUE_ENGINE_HOST_DATA_PATH"]
  Write-Value "mount data path" $envMap["DOTBLUE_ENGINE_MOUNT_DATA_PATH"]

  Write-Section "Docker Socket"
  if ([string]::IsNullOrWhiteSpace($hostSocketGid)) {
    Write-Host "[warn] host docker.sock gid: unavailable from current shell"
  }
  else {
    Write-Check "host docker.sock gid vs env" $hostSocketGid.Trim() $envMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"]
  }
  Write-Value "dotblue user" $backendUser
  Write-Value "dotblue state" $backendState
  Write-Value "dotblue networks" $backendNetworks
  Write-Value "dotblue socket stat" ([string]$backendSocketMeta).Trim()
  Write-Value "dotblue id" ([string]$backendIdentity).Trim()

  $expectedGid = $envMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"]
  if (-not [string]::IsNullOrWhiteSpace($expectedGid)) {
    if (($backendIdentity | Out-String) -match "(^|[^\d])$([regex]::Escape($expectedGid))\(") {
      Write-Host "[ok] dotblue identity contains docker socket gid"
    }
    else {
      Write-Host "[warn] dotblue identity missing docker socket gid $expectedGid"
    }
  }

  Write-Section "Compose"
  Invoke-Docker compose ps

  Write-Section "Hermes Containers"
  $hermes = Invoke-Docker ps --format '{{.Names}}' --filter 'name=^hermes_'
  if ([string]::IsNullOrWhiteSpace($hermes)) {
    Write-Host "(none running)"
  }
  else {
    Invoke-Docker ps --format '{{.Names}} | {{.Status}} | {{.Ports}}' --filter 'name=^hermes_'
  }
}
finally {
  Pop-Location
}
