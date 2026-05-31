Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$envFile = Join-Path $scriptDir ".env"
$generatedDir = Join-Path $scriptDir ".generated"
$casdoorDir = Join-Path $generatedDir "casdoor"
$dotblueDir = Join-Path $generatedDir "dotblue"
$generatedStart = "# >>> prepare-config generated >>>"
$generatedEnd = "# <<< prepare-config generated <<<"

function Fail($Message) {
  throw $Message
}

function Require-File($Path) {
  if (-not (Test-Path $Path)) {
    Fail "Missing required file: $Path"
  }
}

function Read-DotEnv($Path) {
  $result = @{}
  foreach ($line in Get-Content -Path $Path) {
    if ([string]::IsNullOrWhiteSpace($line)) { continue }
    if ($line.TrimStart().StartsWith("#")) { continue }
    $parts = $line -split "=", 2
    if ($parts.Count -ne 2) { continue }
    $result[$parts[0].Trim()] = $parts[1]
  }
  return $result
}

function Require-Env($EnvMap, $Key) {
  if (-not $EnvMap.ContainsKey($Key) -or [string]::IsNullOrWhiteSpace($EnvMap[$Key])) {
    Fail "Missing required value in .env: $Key"
  }
}

function New-RandomHex([int]$Bytes) {
  $data = New-Object byte[] $Bytes
  $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try {
    $rng.GetBytes($data)
  } finally {
    $rng.Dispose()
  }
  return ([System.BitConverter]::ToString($data)).Replace("-", "").ToLowerInvariant()
}

function Ensure-Directory($Path) {
  New-Item -ItemType Directory -Force -Path $Path | Out-Null
}

function New-SelfSignedPemFiles($CertPath, $KeyPath, $SubjectName) {
  if ((Test-Path $CertPath) -and (Test-Path $KeyPath)) {
    return
  }

  $canUseModernPemApi = $false
  try {
    $rsa = [System.Security.Cryptography.RSA]::Create(2048)
    $request = [System.Security.Cryptography.X509Certificates.CertificateRequest]::new(
      "CN=$SubjectName",
      $rsa,
      [System.Security.Cryptography.HashAlgorithmName]::SHA256,
      [System.Security.Cryptography.RSASignaturePadding]::Pkcs1
    )
    $request.CertificateExtensions.Add(
      [System.Security.Cryptography.X509Certificates.X509BasicConstraintsExtension]::new($false, $false, 0, $false)
    )
    $request.CertificateExtensions.Add(
      [System.Security.Cryptography.X509Certificates.X509SubjectKeyIdentifierExtension]::new($request.PublicKey, $false)
    )
    $cert = $request.CreateSelfSigned([datetimeoffset]::UtcNow.AddDays(-1), [datetimeoffset]::UtcNow.AddYears(10))
    $certMethod = $cert.GetType().GetMethod("ExportCertificatePem")
    $keyMethod = $rsa.GetType().GetMethod("ExportPkcs8PrivateKeyPem")
    $canUseModernPemApi = ($null -ne $certMethod) -and ($null -ne $keyMethod)
  } catch {
    $canUseModernPemApi = $false
  }

  if ($canUseModernPemApi) {
    $certPem = $cert.ExportCertificatePem().Trim()
    $keyPem = $rsa.ExportPkcs8PrivateKeyPem().Trim()
    Set-Content -Path $CertPath -Value $certPem
    Set-Content -Path $KeyPath -Value $keyPem
    return
  }

  $openssl = Get-Command openssl.exe -ErrorAction SilentlyContinue
  if (-not $openssl) {
    $openssl = Get-Command openssl -ErrorAction SilentlyContinue
  }
  if (-not $openssl) {
    Fail "Cannot generate PEM certificate on this PowerShell runtime. Install OpenSSL or run prepare-config.sh under WSL/Linux."
  }

  & $openssl.Source req -x509 -nodes -newkey rsa:2048 -keyout $KeyPath -out $CertPath -days 3650 -subj "/CN=$SubjectName" | Out-Null
  if ($LASTEXITCODE -ne 0) {
    Fail "OpenSSL failed to generate certificate files."
  }
}

function Update-EnvBlock($Path, $GeneratedValues) {
  $lines = @()
  if (Test-Path $Path) {
    $skip = $false
    foreach ($line in Get-Content -Path $Path) {
      if ($line -eq $generatedStart) { $skip = $true; continue }
      if ($line -eq $generatedEnd) { $skip = $false; continue }
      if (-not $skip) { $lines += $line }
    }
  }
  while ($lines.Count -gt 0 -and [string]::IsNullOrWhiteSpace($lines[-1])) {
    $lines = $lines[0..($lines.Count - 2)]
  }

  $output = @()
  if ($lines.Count -gt 0) {
    $output += $lines
    $output += ""
  }
  $output += $generatedStart
  foreach ($item in $GeneratedValues.GetEnumerator() | Sort-Object Key) {
    $output += "$($item.Key)=$($item.Value)"
  }
  $output += $generatedEnd

  Set-Content -Path $Path -Value $output
}

function Resolve-HostPath([string]$PathValue) {
  if ([System.IO.Path]::IsPathRooted($PathValue)) {
    return [System.IO.Path]::GetFullPath($PathValue)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $scriptDir $PathValue))
}

function Resolve-DockerSocketGid($EnvMap) {
  if ($EnvMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"]) {
    return $EnvMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"]
  }
  if ($EnvMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"] -ne "unix:///var/run/docker.sock") {
    return ""
  }
  $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
  if (-not $wsl) {
    Fail "Cannot detect docker socket gid automatically. Install WSL or set DOTBLUE_ENGINE_DOCKER_SOCKET_GID manually in .env."
  }
  $gid = & $wsl.Source -e sh -lc "stat -c %g /var/run/docker.sock" 2>$null
  if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($gid)) {
    Fail "Cannot detect docker socket gid from WSL. Set DOTBLUE_ENGINE_DOCKER_SOCKET_GID manually in .env."
  }
  return $gid.Trim()
}

function Get-RedirectUris($EnvMap) {
  $seen = @{}
  $uris = New-Object System.Collections.Generic.List[string]
  $defaults = @(
    "$($EnvMap["DOTBLUE_PUBLIC_URL"])/callback",
    "http://localhost:9000/callback",
    "http://127.0.0.1:9000/callback"
  )
  $extras = @()
  if ($EnvMap.ContainsKey("DOTBLUE_CASDOOR_EXTRA_REDIRECT_URIS") -and -not [string]::IsNullOrWhiteSpace($EnvMap["DOTBLUE_CASDOOR_EXTRA_REDIRECT_URIS"])) {
    $extras = $EnvMap["DOTBLUE_CASDOOR_EXTRA_REDIRECT_URIS"].Split(",")
  }
  foreach ($raw in ($defaults + $extras)) {
    $value = "$raw".Trim()
    if ([string]::IsNullOrWhiteSpace($value)) { continue }
    if ($seen.ContainsKey($value)) { continue }
    $seen[$value] = $true
    $uris.Add($value)
  }
  return $uris.ToArray()
}

Require-File $envFile
$envMap = Read-DotEnv $envFile

$requiredKeys = @(
  "CASDOOR_PUBLIC_URL",
  "CASDOOR_INTERNAL_URL",
  "CASDOOR_ORG_NAME",
  "CASDOOR_APP_NAME",
  "CASDOOR_DB_NAME",
  "CASDOOR_DB_USER",
  "CASDOOR_DB_PASSWORD",
  "DOTBLUE_PUBLIC_URL",
  "DOTBLUE_BACKEND_PUBLIC_URL",
  "DOTBLUE_DB_NAME",
  "DOTBLUE_DB_USER",
  "DOTBLUE_DB_PASSWORD",
  "DOTBLUE_ADMIN_USERNAME",
  "DOTBLUE_ADMIN_DISPLAY_NAME",
  "DOTBLUE_ADMIN_EMAIL",
  "DOTBLUE_ADMIN_PASSWORD",
  "DOTBLUE_BRAND_NAME",
  "DOTBLUE_THEME_PRIMARY"
)
foreach ($key in $requiredKeys) {
  Require-Env $envMap $key
}

if (-not $envMap["COMPOSE_PROJECT_NAME"]) {
  $envMap["COMPOSE_PROJECT_NAME"] = "dotblue"
}
if (-not $envMap["DOTBLUE_ENGINE_HOST_DATA_PATH"]) {
  $envMap["DOTBLUE_ENGINE_HOST_DATA_PATH"] = "./.runtime/agents-host"
}
if (-not $envMap["DOTBLUE_ENGINE_MOUNT_DATA_PATH"]) {
  $envMap["DOTBLUE_ENGINE_MOUNT_DATA_PATH"] = "/runtime-data"
}
if (-not $envMap["DOTBLUE_ENGINE_RUNTIME_MODE"]) {
  $envMap["DOTBLUE_ENGINE_RUNTIME_MODE"] = "container"
}
if (-not $envMap["DOTBLUE_ENGINE_ENDPOINT_MODE"]) {
  $envMap["DOTBLUE_ENGINE_ENDPOINT_MODE"] = "docker_dns"
}
if (-not $envMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"]) {
  $envMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"] = "unix:///var/run/docker.sock"
}
if (-not $envMap["DOTBLUE_ENGINE_DOCKER_NETWORK"]) {
  $envMap["DOTBLUE_ENGINE_DOCKER_NETWORK"] = "$($envMap["COMPOSE_PROJECT_NAME"])_default"
}
$envMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"] = Resolve-DockerSocketGid $envMap
$dotblueEngineHostDataPathAbs = Resolve-HostPath $envMap["DOTBLUE_ENGINE_HOST_DATA_PATH"]

if (-not $envMap["DOTBLUE_CASDOOR_CLIENT_ID"]) {
  $envMap["DOTBLUE_CASDOOR_CLIENT_ID"] = New-RandomHex 10
}
if (-not $envMap["DOTBLUE_CASDOOR_CLIENT_SECRET"]) {
  $envMap["DOTBLUE_CASDOOR_CLIENT_SECRET"] = New-RandomHex 20
}
if (-not $envMap["DOTBLUE_CASDOOR_CERT_NAME"]) {
  $envMap["DOTBLUE_CASDOOR_CERT_NAME"] = "dotblue-jwt-" + (New-RandomHex 4)
}

Ensure-Directory $generatedDir
Ensure-Directory $casdoorDir
Ensure-Directory (Join-Path $casdoorDir "logs")
Ensure-Directory (Join-Path $casdoorDir "certs")
Ensure-Directory $dotblueDir
Ensure-Directory $dotblueEngineHostDataPathAbs

$certPath = Join-Path (Join-Path $casdoorDir "certs") "$($envMap["DOTBLUE_CASDOOR_CERT_NAME"]).pem"
$keyPath = Join-Path (Join-Path $casdoorDir "certs") "$($envMap["DOTBLUE_CASDOOR_CERT_NAME"])-key.pem"
New-SelfSignedPemFiles -CertPath $certPath -KeyPath $keyPath -SubjectName $envMap["DOTBLUE_CASDOOR_CERT_NAME"]

$certPem = (Get-Content -Path $certPath -Raw).Trim()
$keyPem = (Get-Content -Path $keyPath -Raw).Trim()
$certPemBlock = (($certPem -split "`r?`n") | ForEach-Object { "    $_" }) -join "`n"

Update-EnvBlock -Path $envFile -GeneratedValues @{
  DOTBLUE_CASDOOR_CERT_NAME = $envMap["DOTBLUE_CASDOOR_CERT_NAME"]
  DOTBLUE_CASDOOR_CLIENT_ID = $envMap["DOTBLUE_CASDOOR_CLIENT_ID"]
  DOTBLUE_CASDOOR_CLIENT_SECRET = $envMap["DOTBLUE_CASDOOR_CLIENT_SECRET"]
  DOTBLUE_ENGINE_DOCKER_SOCKET_GID = $envMap["DOTBLUE_ENGINE_DOCKER_SOCKET_GID"]
}

$appConf = @"
appname = casdoor
httpport = 8000
runmode = prod
copyrequestbody = true
driverName = postgres
dataSourceName = "user=$($envMap["CASDOOR_DB_USER"]) password=$($envMap["CASDOOR_DB_PASSWORD"]) host=postgres port=5432 sslmode=disable dbname=$($envMap["CASDOOR_DB_NAME"])"
dbName = $($envMap["CASDOOR_DB_NAME"])
tableNamePrefix =
showSql = false
redisEndpoint =
defaultStorageProvider =
isCloudIntranet = false
authState = "casdoor"
socks5Proxy =
verificationCodeTimeout = 10
initScore = 0
logPostOnly = true
isUsernameLowered = false
origin = $($envMap["CASDOOR_PUBLIC_URL"])
originFrontend = $($envMap["CASDOOR_PUBLIC_URL"])
staticBaseUrl = "https://cdn.casbin.org"
isDemoMode = false
batchSize = 100
showGithubCorner = false
forceLanguage = ""
defaultLanguage = "en"
aiAssistantUrl = "https://ai.casbin.com"
defaultApplication = "$($envMap["CASDOOR_APP_NAME"])"
maxItemsForFlatMenu = 7
enableErrorMask = false
enableGzip = true
inactiveTimeoutMinutes =
ldapServerPort = 389
ldapsCertId = ""
ldapsServerPort = 636
radiusServerPort = 1812
radiusDefaultOrganization = "$($envMap["CASDOOR_ORG_NAME"])"
radiusSecret = "secret"
quota = {"organization": -1, "user": -1, "application": -1, "provider": -1}
logConfig = {"adapter":"file", "filename": "logs/casdoor.log", "maxdays":99999, "perm":"0770"}
initDataNewOnly = false
initDataFile = "/init_data.json"
"@
Set-Content -Path (Join-Path $casdoorDir "app.conf") -Value $appConf

$casdoorInitData = @{
  organizations = @(
    @{
      owner = "admin"
      name = $envMap["CASDOOR_ORG_NAME"]
      displayName = $envMap["DOTBLUE_BRAND_NAME"]
      websiteUrl = $envMap["DOTBLUE_PUBLIC_URL"]
      favicon = "$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-favicon.svg"
      defaultApplication = $envMap["CASDOOR_APP_NAME"]
      passwordType = "bcrypt"
      passwordOptions = @("AtLeast6")
      countryCodes = @("US", "CN", "DE", "JP", "SG")
      languages = @("en", "zh")
      isProfilePublic = $true
      disableSignin = $false
    }
  )
  applications = @(
    @{
      owner = "admin"
      name = $envMap["CASDOOR_APP_NAME"]
      displayName = $envMap["DOTBLUE_BRAND_NAME"]
      logo = "$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-logo.png"
      favicon = "$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-favicon.svg"
      homepageUrl = $envMap["DOTBLUE_PUBLIC_URL"]
      organization = $envMap["CASDOOR_ORG_NAME"]
      cert = $envMap["DOTBLUE_CASDOOR_CERT_NAME"]
      defaultGroup = "admin"
      enablePassword = $true
      enableSignUp = $true
      disableSignin = $false
      enableSigninSession = $false
      clientId = $envMap["DOTBLUE_CASDOOR_CLIENT_ID"]
      clientSecret = $envMap["DOTBLUE_CASDOOR_CLIENT_SECRET"]
      signinMethods = @(@{ name = "Password"; displayName = "Password"; rule = "All" })
      signupItems = @(
        @{ name = "Username"; visible = $true; required = $true; prompted = $false; rule = "None" },
        @{ name = "Display name"; visible = $true; required = $true; prompted = $false; rule = "None" },
        @{ name = "Password"; visible = $true; required = $true; prompted = $false; rule = "None" },
        @{ name = "Confirm password"; visible = $true; required = $true; prompted = $false; rule = "None" }
      )
      grantTypes = @("authorization_code", "password", "client_credentials", "refresh_token")
      redirectUris = @(Get-RedirectUris $envMap)
      tokenFormat = "JWT"
      tokenFields = @()
      expireInHours = 168
      themeData = @{
        themeType = "default"
        colorPrimary = $envMap["DOTBLUE_THEME_PRIMARY"]
        borderRadius = 12
        isCompact = $false
        isEnabled = $true
      }
      formCss = ""
      formSideHtml = "<div style=`"display:flex;flex-direction:column;justify-content:flex-end;min-height:640px;padding:52px;background:linear-gradient(180deg, rgba(15,23,42,0.10), rgba(15,23,42,0.42)), url('$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-login-bg.png') center/cover no-repeat;color:#fff;`"><div style=`"max-width:430px;`"><img src=`"$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-logo.png`" alt=`"$($envMap["DOTBLUE_BRAND_NAME"])`" style=`"width:168px;max-width:100%;height:auto;display:block;margin-bottom:22px;`" /><div style=`"display:inline-flex;align-items:center;padding:6px 12px;border-radius:999px;background:rgba(255,255,255,0.16);font-size:12px;font-weight:600;letter-spacing:.08em;margin-bottom:18px;`">ENTERPRISE AI ASSISTANTS</div><h2 style=`"margin:0 0 14px;font-size:36px;line-height:1.1;color:#fff;`">Launch secure AI workspaces with dotblue</h2><p style=`"margin:0 0 18px;font-size:15px;line-height:1.8;color:rgba(255,255,255,0.88);`">Unified login, runtime governance, and ready-to-launch assistant experiences for product, operations, and growth teams.</p><div style=`"display:flex;flex-direction:column;gap:10px;`"><div style=`"display:flex;align-items:center;gap:10px;font-size:14px;line-height:1.5;color:rgba(255,255,255,0.90);`"><span style=`"width:8px;height:8px;border-radius:999px;background:#7dd3fc;display:inline-block;flex-shrink:0;`"></span><span>Casdoor powered branded access</span></div><div style=`"display:flex;align-items:center;gap:10px;font-size:14px;line-height:1.5;color:rgba(255,255,255,0.90);`"><span style=`"width:8px;height:8px;border-radius:999px;background:#7dd3fc;display:inline-block;flex-shrink:0;`"></span><span>Enterprise-ready assistant workspace</span></div><div style=`"display:flex;align-items:center;gap:10px;font-size:14px;line-height:1.5;color:rgba(255,255,255,0.90);`"><span style=`"width:8px;height:8px;border-radius:999px;background:#7dd3fc;display:inline-block;flex-shrink:0;`"></span><span>Deployment and runtime governance built in</span></div></div></div></div>"
      formBackgroundUrl = "$($envMap["DOTBLUE_PUBLIC_URL"])/brand/dotblue-login-bg.png"
    }
  )
  users = @(
    @{
      owner = $envMap["CASDOOR_ORG_NAME"]
      name = $envMap["DOTBLUE_ADMIN_USERNAME"]
      type = "normal-user"
      password = $envMap["DOTBLUE_ADMIN_PASSWORD"]
      displayName = $envMap["DOTBLUE_ADMIN_DISPLAY_NAME"]
      avatar = ""
      email = $envMap["DOTBLUE_ADMIN_EMAIL"]
      phone = ""
      countryCode = ""
      address = @()
      addresses = @()
      affiliation = ""
      tag = ""
      score = 2000
      ranking = 1
      isAdmin = $true
      isForbidden = $false
      isDeleted = $false
      signupApplication = $envMap["CASDOOR_APP_NAME"]
      createdIp = ""
      groups = @("admin")
    }
  )
  certs = @(
    @{
      owner = "admin"
      name = $envMap["DOTBLUE_CASDOOR_CERT_NAME"]
      displayName = "$($envMap["DOTBLUE_BRAND_NAME"]) JWT"
      scope = "JWT"
      type = "x509"
      cryptoAlgorithm = "RS256"
      bitSize = 2048
      expireInYears = 10
      certificate = $certPem
      privateKey = $keyPem
    }
  )
  groups = @(
    @{
      owner = $envMap["CASDOOR_ORG_NAME"]
      name = "admin"
      displayName = "Platform Admins"
      manager = $envMap["DOTBLUE_ADMIN_USERNAME"]
      contactEmail = $envMap["DOTBLUE_ADMIN_EMAIL"]
      type = "Virtual"
      parent_id = ""
      isTopGroup = $true
      title = ""
      key = ""
      children = @()
      isEnabled = $true
    }
  )
}
$casdoorInitData | ConvertTo-Json -Depth 12 | Set-Content -Path (Join-Path $casdoorDir "init_data.json")

$backendConfig = @"
server:
  address: ":8000"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

database:
  default:
    link: "pgsql:$($envMap["DOTBLUE_DB_USER"]):$($envMap["DOTBLUE_DB_PASSWORD"])@tcp(postgres:5432)/$($envMap["DOTBLUE_DB_NAME"])"
    debug: true

casdoor:
  endpoint: "$($envMap["CASDOOR_INTERNAL_URL"])"
  clientId: "$($envMap["DOTBLUE_CASDOOR_CLIENT_ID"])"
  clientSecret: "$($envMap["DOTBLUE_CASDOOR_CLIENT_SECRET"])"
  jwtSecret: |
$certPemBlock
  organizationName: "$($envMap["CASDOOR_ORG_NAME"])"
  applicationName: "$($envMap["CASDOOR_APP_NAME"])"
  bootstrap:
    endpoint: ""
    clientId: ""
    clientSecret: ""
    jwtSecret: ""

setup:
  initDataPath: ""

logger:
  level: "all"
  stdout: true

debug:
  sse: true

im:
  asyncTurn: true

redis:
  address: "redis:6379"
  password: ""
  db: 0
  keyPrefix: "dot"

session:
  ownerTTL: "30s"
  fenceTTL: "2m"
  gateTTL: "2m"
  stateTTL: "2h"

worker:
  id: "compose-all-in-one"
  embedded: true
  metaTTL: "30s"
  heartbeatInterval: "10s"
  inboxTTL: "2h"
  claimBlock: "2s"

dataplane:
  requestStateRunningTTL: "30m"
  requestStateFinalTTL: "1h"
  streamMaxLen: 5000

engine:
  dataBasePath: "$dotblueEngineHostDataPathAbs"
  dataMountPath: "$($envMap["DOTBLUE_ENGINE_MOUNT_DATA_PATH"])"
  containerPort: 8642
  runtimeMode: "$($envMap["DOTBLUE_ENGINE_RUNTIME_MODE"])"
  endpointMode: "$($envMap["DOTBLUE_ENGINE_ENDPOINT_MODE"])"
  dockerEndpoint: "$($envMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"])"
  dockerNetwork: "$($envMap["DOTBLUE_ENGINE_DOCKER_NETWORK"])"
"@
Set-Content -Path (Join-Path $dotblueDir "config.yaml") -Value $backendConfig

$dotblueInitData = @{
  version = 1
  syncCasdoor = $false
  organization = @{
    name = $envMap["CASDOOR_ORG_NAME"]
    displayName = $envMap["DOTBLUE_BRAND_NAME"]
  }
  admin = @{
    username = $envMap["DOTBLUE_ADMIN_USERNAME"]
    displayName = $envMap["DOTBLUE_ADMIN_DISPLAY_NAME"]
    email = $envMap["DOTBLUE_ADMIN_EMAIL"]
    password = $envMap["DOTBLUE_ADMIN_PASSWORD"]
  }
  platform = @{
    dataBasePath = $dotblueEngineHostDataPathAbs
    dataMountPath = $envMap["DOTBLUE_ENGINE_MOUNT_DATA_PATH"]
    containerPort = 8642
    runtimeMode = $envMap["DOTBLUE_ENGINE_RUNTIME_MODE"]
    endpointMode = $envMap["DOTBLUE_ENGINE_ENDPOINT_MODE"]
    dockerEndpoint = $envMap["DOTBLUE_ENGINE_DOCKER_ENDPOINT"]
    dockerNetwork = $envMap["DOTBLUE_ENGINE_DOCKER_NETWORK"]
  }
}
if ($envMap["DOTBLUE_LLM_API_KEY"]) {
  $dotblueInitData["provider"] = @{
    type = $envMap["DOTBLUE_LLM_PROVIDER_TYPE"]
    apiBase = $envMap["DOTBLUE_LLM_API_BASE"]
    apiKey = $envMap["DOTBLUE_LLM_API_KEY"]
    model = $envMap["DOTBLUE_LLM_MODEL"]
  }
}
$dotblueInitData | ConvertTo-Json -Depth 10 | Set-Content -Path (Join-Path $dotblueDir "init_data.json")

Write-Host "Generated files:"
Write-Host "  - $casdoorDir/app.conf"
Write-Host "  - $casdoorDir/init_data.json"
Write-Host "  - $dotblueDir/config.yaml"
Write-Host "  - $dotblueDir/init_data.json"
Write-Host ""
Write-Host "Updated generated values in $envFile"
