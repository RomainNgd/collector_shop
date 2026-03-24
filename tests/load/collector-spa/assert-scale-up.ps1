param(
    [string]$Namespace = "collector-shop-prod",
    [string]$Deployment = "collector-spa",
    [string]$HpaName = "collector-spa",
    [int]$TimeoutSeconds = 240,
    [int]$PollSeconds = 5
)

$ErrorActionPreference = "Stop"

function Get-ReplicaValue {
    param(
        [string]$Kind,
        [string]$Name,
        [string]$JsonPath
    )

    $value = kubectl -n $Namespace get $Kind $Name -o "jsonpath=$JsonPath" 2>$null
    if ([string]::IsNullOrWhiteSpace($value)) {
        return 0
    }

    return [int]$value
}

$initialReady = Get-ReplicaValue -Kind "deployment" -Name $Deployment -JsonPath "{.status.readyReplicas}"
$deadline = (Get-Date).AddSeconds($TimeoutSeconds)

Write-Host "Initial ready replicas for $Deployment: $initialReady"
Write-Host "Watching for a scale-out during $TimeoutSeconds seconds..."

while ((Get-Date) -lt $deadline) {
    $readyReplicas = Get-ReplicaValue -Kind "deployment" -Name $Deployment -JsonPath "{.status.readyReplicas}"
    $desiredReplicas = Get-ReplicaValue -Kind "hpa" -Name $HpaName -JsonPath "{.status.desiredReplicas}"
    $currentReplicas = Get-ReplicaValue -Kind "hpa" -Name $HpaName -JsonPath "{.status.currentReplicas}"

    Write-Host ("[{0}] ready={1} hpaCurrent={2} hpaDesired={3}" -f (Get-Date -Format "HH:mm:ss"), $readyReplicas, $currentReplicas, $desiredReplicas)

    if ($readyReplicas -gt $initialReady) {
        Write-Host "Scale-out detected: ready replicas moved from $initialReady to $readyReplicas."
        exit 0
    }

    Start-Sleep -Seconds $PollSeconds
}

Write-Error "No extra ready pod was observed for deployment '$Deployment' within $TimeoutSeconds seconds."
exit 1
