#!/usr/bin/env bash

set -Eeuo pipefail

PROJECT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${PROJECT_DIR}/backend"
FRONTEND_DIR="${PROJECT_DIR}/frontend"

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  echo "Penggunaan: ./run-dev.sh"
  echo "Menjalankan backend Go dan frontend Vue secara bersamaan."
  exit 0
fi

if [[ $# -gt 0 ]]; then
  echo "Argumen tidak dikenal: $1" >&2
  echo "Gunakan ./run-dev.sh --help untuk bantuan." >&2
  exit 2
fi

for command_name in go npm; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "Perintah '${command_name}' tidak ditemukan." >&2
    exit 1
  fi
done

if [[ ! -f "${BACKEND_DIR}/.env" ]]; then
  echo "File backend/.env belum ada." >&2
  echo "Buat dengan: cp backend/.env.example backend/.env" >&2
  exit 1
fi

if [[ ! -f "${FRONTEND_DIR}/.env" ]]; then
  echo "File frontend/.env belum ada." >&2
  echo "Buat dengan: cp frontend/.env.example frontend/.env" >&2
  exit 1
fi

if [[ ! -d "${FRONTEND_DIR}/node_modules" ]]; then
  echo "Dependency frontend belum terpasang." >&2
  echo "Jalankan: cd frontend && npm install" >&2
  exit 1
fi

child_pids=()

cleanup() {
  local exit_status=$?
  trap - EXIT

  if [[ ${#child_pids[@]} -gt 0 ]]; then
    echo
    echo "Menghentikan backend dan frontend..."
    kill "${child_pids[@]}" 2>/dev/null || true
    wait "${child_pids[@]}" 2>/dev/null || true
  fi

  exit "${exit_status}"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "Menjalankan SIRAPI development..."
echo "Backend : http://127.0.0.1:8080"
echo "Frontend: http://localhost:5173"
echo "Tekan Ctrl+C untuk menghentikan keduanya."
echo

(
  cd "${BACKEND_DIR}"
  exec go run ./cmd/api
) &
child_pids+=("$!")

(
  cd "${FRONTEND_DIR}"
  exec npm run dev
) &
child_pids+=("$!")

set +e
wait -n "${child_pids[@]}"
exit_status=$?
set -e

if [[ ${exit_status} -eq 0 ]]; then
  echo "Salah satu proses berhenti; proses lainnya akan dihentikan."
else
  echo "Salah satu proses gagal dengan status ${exit_status}; proses lainnya akan dihentikan." >&2
fi

exit "${exit_status}"
