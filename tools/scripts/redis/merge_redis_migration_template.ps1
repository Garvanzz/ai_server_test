param(
  [Parameter(Mandatory = $true)]
  [ValidateSet('export', 'import')]
  [string]$Mode,

  [Parameter(Mandatory = $true)]
  [string]$SourceRedis,

  [Parameter(Mandatory = $true)]
  [string]$TargetRedis,

  [Parameter(Mandatory = $true)]
  [string]$SourceEntryServers,

  [Parameter(Mandatory = $true)]
  [int]$TargetLogicServer,

  [string]$OutputDir = '.\redis_merge_output'
)

$ErrorActionPreference = 'Stop'

function Write-Step([string]$Message) {
  Write-Host "[merge-redis] $Message"
}

function Get-KeyPatterns {
  return @(
    'AccountRole:*',
    'Player:*',
    'Base:*',
    'Bag:*',
    'Shop:*',
    'Hero:*',
    'Activity:*',
    'ActivityPlayer:*',
    'rank_*',
    'GuildManager:*',
    'player_guild:*',
    'guild_chat_history:*',
    'systemMailId:*',
    'dailyMail:*'
  )
}

function Ensure-Dir([string]$Path) {
  if (-not (Test-Path $Path)) {
    New-Item -ItemType Directory -Path $Path | Out-Null
  }
}

Write-Step "Mode=$Mode"
Write-Step "SourceRedis=$SourceRedis"
Write-Step "TargetRedis=$TargetRedis"
Write-Step "SourceEntryServers=$SourceEntryServers"
Write-Step "TargetLogicServer=$TargetLogicServer"

Ensure-Dir $OutputDir

$patterns = Get-KeyPatterns
$patternFile = Join-Path $OutputDir 'key_patterns.txt'
$patterns | Set-Content -Path $patternFile -Encoding UTF8
Write-Step "Key patterns written to $patternFile"

if ($Mode -eq 'export') {
  Write-Step 'This template does not automatically copy production data.'
  Write-Step 'Recommended commands:'
  foreach ($pattern in $patterns) {
    Write-Host "redis-cli -h <source-host> -p <source-port> --scan --pattern '$pattern'"
  }
  Write-Step 'Export matching keys with redis-cli --rdb or custom dump/restore flow.'
  exit 0
}

if ($Mode -eq 'import') {
  Write-Step 'Before import, confirm source and target Redis backups exist.'
  Write-Step 'Then replay the exported RDB / DUMP payloads into target Redis.'
  Write-Step 'After import, sample-verify:'
  Write-Host 'redis-cli -h <target-host> -p <target-port> GET AccountRole:<uid>:<entryServerId>'
  Write-Host 'redis-cli -h <target-host> -p <target-port> HGETALL Player:<playerId>'
  Write-Host 'redis-cli -h <target-host> -p <target-port> ZRANGE rank_arena:<actId> 0 10 WITHSCORES'
  exit 0
}
