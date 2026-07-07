#!/usr/bin/env bash

set -Eeuo pipefail

DRY_RUN=false

case "${1:-}" in
  --dry-run|-DryRun)
    DRY_RUN=true
    ;;
  "")
    ;;
  *)
    echo "Option inconnue : $1" >&2
    echo "Usage : $0 [--dry-run]" >&2
    exit 1
    ;;
esac

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

GO_API_DIR="$REPO_ROOT/go-api"
FRONT_DIR="$REPO_ROOT/collector-spa"
GO_ENV_PATH="$GO_API_DIR/.env"

declare -A GO_ENV

write_step() {
  local message="$1"
  echo
  printf '\033[36m==> %s\033[0m\n' "$message"
}

assert_command() {
  local name="$1"

  if ! command -v "$name" >/dev/null 2>&1; then
    echo "La commande '$name' est introuvable. Ajoute-la au PATH avant de lancer ce script." >&2
    exit 1
  fi
}

trim() {
  local value="$1"

  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"

  printf '%s' "$value"
}

read_dotenv() {
  local path="$1"
  local line trimmed_line key value

  if [[ ! -f "$path" ]]; then
    return
  fi

  while IFS= read -r line || [[ -n "$line" ]]; do
    trimmed_line="$(trim "$line")"

    [[ -z "$trimmed_line" ]] && continue
    [[ "$trimmed_line" == \#* ]] && continue
    [[ "$trimmed_line" != *=* ]] && continue

    key="$(trim "${trimmed_line%%=*}")"
    value="$(trim "${trimmed_line#*=}")"

    [[ -z "$key" ]] && continue

    if [[ ${#value} -ge 2 ]]; then
      if [[ "$value" == \"*\" ]] || [[ "$value" == \'*\' ]]; then
        local first_char="${value:0:1}"
        local last_char="${value: -1}"

        if [[ "$first_char" == "$last_char" ]]; then
          value="${value:1:${#value}-2}"
        fi
      fi
    fi

    GO_ENV["$key"]="$value"
  done < "$path"
}

format_command() {
  local -a parts=()
  local argument

  for argument in "$@"; do
    printf -v argument '%q' "$argument"
    parts+=("$argument")
  done

  printf '%s ' "${parts[@]}"
}

invoke_checked_command() {
  local working_directory="$1"
  local description="$2"
  shift 2

  local command_text
  command_text="$(format_command "$@")"

  if [[ "$DRY_RUN" == true ]]; then
    printf '\033[33m[dry-run] %s\033[0m\n' "$command_text"
    return
  fi

  (
    cd -- "$working_directory"
    "$@"
  ) || {
    local exit_code=$?
    echo "$description a échoué avec le code $exit_code." >&2
    exit "$exit_code"
  }
}

wait_for_database() {
  local db_user="$1"
  local db_name="$2"
  local max_attempts=30
  local attempt output exit_code

  if [[ "$DRY_RUN" == true ]]; then
    printf '\033[33m[dry-run] docker compose exec -T db pg_isready -U %q -d %q\033[0m\n' \
      "$db_user" "$db_name"
    return
  fi

  for ((attempt = 1; attempt <= max_attempts; attempt++)); do
    echo "Attente de PostgreSQL ($attempt/$max_attempts)..."

    set +e
    output="$(
      cd -- "$GO_API_DIR" &&
      docker compose exec -T db pg_isready -U "$db_user" -d "$db_name" 2>&1
    )"
    exit_code=$?
    set -e

    [[ -n "$output" ]] && printf '%s\n' "$output"

    if [[ $exit_code -eq 0 ]]; then
      return
    fi

    sleep 2
  done

  echo "PostgreSQL n'est pas prêt après $max_attempts tentatives." >&2
  exit 1
}

get_terminal_command() {
  if command -v gnome-terminal >/dev/null 2>&1; then
    printf '%s' "gnome-terminal"
  elif command -v konsole >/dev/null 2>&1; then
    printf '%s' "konsole"
  elif command -v xterm >/dev/null 2>&1; then
    printf '%s' "xterm"
  else
    return 1
  fi
}

start_dev_window() {
  local title="$1"
  local working_directory="$2"
  local command_text="$3"
  local terminal
  local shell_command

  terminal="$(get_terminal_command)" || {
    echo "Aucun émulateur de terminal compatible détecté (gnome-terminal, konsole ou xterm)." >&2
    exit 1
  }

  shell_command="cd $(printf '%q' "$working_directory"); $command_text; status=\$?; echo; echo \"Le processus s'est arrêté avec le code \$status.\"; exec bash"

  if [[ "$DRY_RUN" == true ]]; then
    printf '\033[33m[dry-run] %s --title=%q -- bash -lc %q\033[0m\n' \
      "$terminal" "$title" "$shell_command"
    return
  fi

  case "$terminal" in
    gnome-terminal)
      gnome-terminal --title="$title" -- bash -lc "$shell_command" >/dev/null 2>&1 &
      ;;
    konsole)
      konsole --new-tab -p tabtitle="$title" -e bash -lc "$shell_command" >/dev/null 2>&1 &
      ;;
    xterm)
      xterm -T "$title" -e bash -lc "$shell_command" >/dev/null 2>&1 &
      ;;
  esac
}

assert_command "docker"
assert_command "go"
assert_command "npm"

if [[ ! -d "$GO_API_DIR" ]]; then
  echo "Le dossier go-api est introuvable : $GO_API_DIR" >&2
  exit 1
fi

if [[ ! -d "$FRONT_DIR" ]]; then
  echo "Le dossier collector-spa est introuvable : $FRONT_DIR" >&2
  exit 1
fi

read_dotenv "$GO_ENV_PATH"

DB_USER="${GO_ENV[DB_USER]:-golang}"
DB_NAME="${GO_ENV[DB_NAME]:-ecommerce}"
JWT_SECRET="${GO_ENV[JWT_SECRET]:-change-this-for-real-tests}"

STRIPE_ENABLED=false
if [[ "${GO_ENV[STRIPE_ENABLED]:-}" =~ ^([Tt][Rr][Uu][Ee])$ ]]; then
  STRIPE_ENABLED=true
fi

HAS_STRIPE_CLI=false
if command -v stripe >/dev/null 2>&1; then
  HAS_STRIPE_CLI=true
fi

write_step "Reset de la base Docker locale"
invoke_checked_command \
  "$GO_API_DIR" \
  "Le reset de la base" \
  docker compose down -v

write_step "Démarrage de PostgreSQL"
invoke_checked_command \
  "$GO_API_DIR" \
  "Le démarrage de PostgreSQL" \
  docker compose up -d db

write_step "Attente de PostgreSQL"
wait_for_database "$DB_USER" "$DB_NAME"

write_step "Création des fixtures"
invoke_checked_command \
  "$GO_API_DIR" \
  "Le seed des fixtures" \
  go run . seed

write_step "Ouverture de l'API et du front"
start_dev_window \
  "collector-shop API" \
  "$GO_API_DIR" \
  "go run ."

start_dev_window \
  "collector-shop Front" \
  "$FRONT_DIR" \
  "export JWT_SECRET=$(printf '%q' "$JWT_SECRET"); npm run dev"

echo
printf '\033[32mEnvironnements lancés :\033[0m\n'
echo "  API   : http://localhost:8080"
echo "  Front : http://localhost:5173"
echo
printf '\033[90mPour arrêter la base plus tard :\033[0m\n'
printf '  cd %q\n' "$GO_API_DIR"
echo "  docker compose down"

if [[ "$STRIPE_ENABLED" == true ]]; then
  echo
  printf '\033[32mStripe demo :\033[0m\n'

  if [[ "$HAS_STRIPE_CLI" == true ]]; then
    printf '\033[90m  Lance ensuite la CLI Stripe dans un autre terminal :\033[0m\n'
    echo "  stripe listen --events checkout.session.completed,checkout.session.expired,checkout.session.async_payment_failed,checkout.session.async_payment_succeeded --forward-to localhost:8080/payments/stripe/webhook"
    echo "  Puis copie le secret whsec_... affiché dans STRIPE_WEBHOOK_SECRET et redémarre l'API."
  else
    printf '\033[33m  La Stripe CLI n'est pas détectée. Installe-la pour transférer les webhooks locaux.\033[0m\n'
  fi
fi