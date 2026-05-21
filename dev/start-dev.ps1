[CmdletBinding()]
param(
    [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$GoApiDir = Join-Path $RepoRoot "go-api"
$FrontDir = Join-Path $RepoRoot "collector-spa"
$GoEnvPath = Join-Path $GoApiDir ".env"

function Write-Step {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message
    )

    Write-Host ""
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Assert-Command {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name
    )

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "La commande '$Name' est introuvable. Ajoute-la au PATH avant de lancer ce script."
    }
}

function Read-DotEnv {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $values = @{}

    if (-not (Test-Path -LiteralPath $Path)) {
        return $values
    }

    foreach ($line in Get-Content -LiteralPath $Path) {
        $trimmedLine = $line.Trim()

        if ($trimmedLine.Length -eq 0 -or $trimmedLine.StartsWith("#")) {
            continue
        }

        $separatorIndex = $trimmedLine.IndexOf("=")
        if ($separatorIndex -lt 1) {
            continue
        }

        $key = $trimmedLine.Substring(0, $separatorIndex).Trim()
        $value = $trimmedLine.Substring($separatorIndex + 1).Trim()

        if ($value.Length -ge 2) {
            if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
                $value = $value.Substring(1, $value.Length - 2)
            }
        }

        $values[$key] = $value
    }

    return $values
}

function Format-Command {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,
        [Parameter(Mandatory = $true)]
        [string[]]$ArgumentList
    )

    $parts = @($FilePath)
    foreach ($argument in $ArgumentList) {
        if ($argument -match "\s") {
            $parts += '"' + $argument + '"'
            continue
        }

        $parts += $argument
    }

    return $parts -join " "
}

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,
        [Parameter(Mandatory = $true)]
        [string[]]$ArgumentList,
        [Parameter(Mandatory = $true)]
        [string]$WorkingDirectory,
        [Parameter(Mandatory = $true)]
        [string]$Description
    )

    $commandText = Format-Command -FilePath $FilePath -ArgumentList $ArgumentList

    if ($DryRun) {
        Write-Host "[dry-run] $commandText" -ForegroundColor Yellow
        return
    }

    Push-Location $WorkingDirectory
    try {
        & $FilePath @ArgumentList
        if ($LASTEXITCODE -ne 0) {
            throw "$Description a echoue avec le code $LASTEXITCODE."
        }
    }
    finally {
        Pop-Location
    }
}

function Wait-ForDatabase {
    param(
        [Parameter(Mandatory = $true)]
        [string]$DbUser,
        [Parameter(Mandatory = $true)]
        [string]$DbName
    )

    if ($DryRun) {
        Write-Host "[dry-run] docker compose exec -T db pg_isready -U $DbUser -d $DbName" -ForegroundColor Yellow
        return
    }

    $maxAttempts = 30

    for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
        Write-Host "Attente de PostgreSQL ($attempt/$maxAttempts)..."

        Push-Location $GoApiDir
        try {
            $output = & docker compose exec -T db pg_isready -U $DbUser -d $DbName 2>&1
            if ($output) {
                $output | ForEach-Object { Write-Host $_ }
            }

            if ($LASTEXITCODE -eq 0) {
                return
            }
        }
        finally {
            Pop-Location
        }

        Start-Sleep -Seconds 2
    }

    throw "PostgreSQL n'est pas pret apres $maxAttempts tentatives."
}

function Escape-SingleQuotedText {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Value
    )

    return $Value -replace "'", "''"
}

function Start-DevWindow {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Title,
        [Parameter(Mandatory = $true)]
        [string]$WorkingDirectory,
        [Parameter(Mandatory = $true)]
        [string]$CommandText
    )

    $shellPath = (Get-Process -Id $PID).Path
    $escapedTitle = Escape-SingleQuotedText -Value $Title
    $escapedWorkingDirectory = Escape-SingleQuotedText -Value $WorkingDirectory
    $inlineCommand = "`$Host.UI.RawUI.WindowTitle = '$escapedTitle'; Set-Location -LiteralPath '$escapedWorkingDirectory'; $CommandText"

    if ($DryRun) {
        Write-Host "[dry-run] Start-Process -FilePath $shellPath -ArgumentList -NoExit -Command $inlineCommand" -ForegroundColor Yellow
        return
    }

    Start-Process -FilePath $shellPath -WorkingDirectory $WorkingDirectory -ArgumentList @("-NoExit", "-Command", $inlineCommand) | Out-Null
}

Assert-Command -Name "docker"
Assert-Command -Name "go"
Assert-Command -Name "npm"

if (-not (Test-Path -LiteralPath $GoApiDir)) {
    throw "Le dossier go-api est introuvable: $GoApiDir"
}

if (-not (Test-Path -LiteralPath $FrontDir)) {
    throw "Le dossier collector-spa est introuvable: $FrontDir"
}

$goEnv = Read-DotEnv -Path $GoEnvPath
$dbUser = if ($goEnv.ContainsKey("DB_USER")) { $goEnv["DB_USER"] } else { "golang" }
$dbName = if ($goEnv.ContainsKey("DB_NAME")) { $goEnv["DB_NAME"] } else { "ecommerce" }
$jwtSecret = if ($goEnv.ContainsKey("JWT_SECRET")) { $goEnv["JWT_SECRET"] } else { "change-this-for-real-tests" }
$stripeEnabled = $goEnv.ContainsKey("STRIPE_ENABLED") -and $goEnv["STRIPE_ENABLED"].ToLowerInvariant() -eq "true"
$hasStripeCli = [bool](Get-Command stripe -ErrorAction SilentlyContinue)

Write-Step -Message "Reset de la base Docker locale"
Invoke-CheckedCommand -FilePath "docker" -ArgumentList @("compose", "down", "-v") -WorkingDirectory $GoApiDir -Description "Le reset de la base"

Write-Step -Message "Demarrage de PostgreSQL"
Invoke-CheckedCommand -FilePath "docker" -ArgumentList @("compose", "up", "-d", "db") -WorkingDirectory $GoApiDir -Description "Le demarrage de PostgreSQL"

Write-Step -Message "Attente de PostgreSQL"
Wait-ForDatabase -DbUser $dbUser -DbName $dbName

Write-Step -Message "Creation des fixtures"
Invoke-CheckedCommand -FilePath "go" -ArgumentList @("run", ".", "seed") -WorkingDirectory $GoApiDir -Description "Le seed des fixtures"

Write-Step -Message "Ouverture de l'API et du front"
Start-DevWindow -Title "collector-shop API" -WorkingDirectory $GoApiDir -CommandText "go run ."
$escapedJwtSecret = Escape-SingleQuotedText -Value $jwtSecret
Start-DevWindow -Title "collector-shop Front" -WorkingDirectory $FrontDir -CommandText "`$env:JWT_SECRET = '$escapedJwtSecret'; npm run dev"

Write-Host ""
Write-Host "Environnements lances:" -ForegroundColor Green
Write-Host "  API   : http://localhost:8080"
Write-Host "  Front : http://localhost:5173"
Write-Host ""
Write-Host "Pour arreter la base plus tard:" -ForegroundColor DarkGray
Write-Host "  Set-Location '$GoApiDir'"
Write-Host "  docker compose down"

if ($stripeEnabled) {
    Write-Host ""
    Write-Host "Stripe demo:" -ForegroundColor Green
    if ($hasStripeCli) {
        Write-Host "  Lance ensuite la CLI Stripe dans un autre terminal:" -ForegroundColor DarkGray
        Write-Host "  stripe listen --events checkout.session.completed,checkout.session.expired,checkout.session.async_payment_failed,checkout.session.async_payment_succeeded --forward-to localhost:8080/payments/stripe/webhook"
        Write-Host "  Puis copie le secret whsec_... affiche dans STRIPE_WEBHOOK_SECRET et redemarre l'API."
    }
    else {
        Write-Host "  La Stripe CLI n'est pas detectee. Installe-la pour forwarder les webhooks locaux." -ForegroundColor Yellow
    }
}
