#!/usr/bin/env bash
# Equivalent bash de assert-scale-up.ps1 : verifie qu'une replique supplementaire
# du front devient prete pendant un tir de charge.
set -euo pipefail

NAMESPACE="${NAMESPACE:-collector-shop-prod}"
DEPLOYMENT="${DEPLOYMENT:-collector-spa}"
HPA_NAME="${HPA_NAME:-collector-spa}"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-240}"
POLL_SECONDS="${POLL_SECONDS:-5}"

get_replica_value() {
    local kind="$1" name="$2" json_path="$3" value
    value="$(kubectl -n "$NAMESPACE" get "$kind" "$name" -o "jsonpath=$json_path" 2>/dev/null || true)"
    if [[ -z "$value" ]]; then
        echo 0
    else
        echo "$value"
    fi
}

initial_ready="$(get_replica_value deployment "$DEPLOYMENT" '{.status.readyReplicas}')"
deadline=$(( $(date +%s) + TIMEOUT_SECONDS ))

echo "Initial ready replicas for $DEPLOYMENT: $initial_ready"
echo "Watching for a scale-out during $TIMEOUT_SECONDS seconds..."

while (( $(date +%s) < deadline )); do
    ready_replicas="$(get_replica_value deployment "$DEPLOYMENT" '{.status.readyReplicas}')"
    desired_replicas="$(get_replica_value hpa "$HPA_NAME" '{.status.desiredReplicas}')"
    current_replicas="$(get_replica_value hpa "$HPA_NAME" '{.status.currentReplicas}')"

    printf '[%s] ready=%s hpaCurrent=%s hpaDesired=%s\n' \
        "$(date +%H:%M:%S)" "$ready_replicas" "$current_replicas" "$desired_replicas"

    if (( ready_replicas > initial_ready )); then
        echo "Scale-out detected: ready replicas moved from $initial_ready to $ready_replicas."
        exit 0
    fi

    sleep "$POLL_SECONDS"
done

echo "No extra ready pod was observed for deployment '$DEPLOYMENT' within $TIMEOUT_SECONDS seconds." >&2
exit 1
