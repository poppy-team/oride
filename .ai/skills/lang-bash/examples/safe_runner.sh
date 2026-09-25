#!/usr/bin/env bash
# Production-Grade Defensive Bash Runner
# Standards:
# - Strict mode: set -euo pipefail
# - Safe IFS
# - Trap-driven temporary resource cleanup
# - Strict double-quoting
# - Zero eval

set -euo pipefail
IFS=$'\n\t'

# Global execution state
readonly SCRIPT_NAME="$(basename "${0}")"
TMP_DIR=""

log_info() {
    echo "[${SCRIPT_NAME}] INFO: $*" >&2
}

log_error() {
    echo "[${SCRIPT_NAME}] ERROR: $*" >&2
}

# Trap handler for deterministic cleanup
cleanup() {
    local exit_code=$?
    trap - EXIT INT TERM HUP
    if [[ -n "${TMP_DIR}" && -d "${TMP_DIR}" ]]; then
        log_info "Cleaning up temporary directory: ${TMP_DIR}"
        rm -rf -- "${TMP_DIR}"
    fi
    exit "${exit_code}"
}
trap cleanup EXIT INT TERM HUP

show_help() {
    cat << EOF
Usage: ${SCRIPT_NAME} [OPTIONS] <TARGET_FILE>

Runs automated verification tasks safely.

Options:
  -h, --help            Show this help message
  -v, --verbose         Enable verbose diagnostic logging
  -o, --output FILE     Write results to specified file
EOF
}

main() {
    local target_file=""
    local output_file=""
    local verbose=0

    # Parse arguments safely
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--verbose)
                verbose=1
                shift
                ;;
            -o|--output)
                if [[ -z "${2:-}" ]]; then
                    log_error "Option '$1' requires an argument."
                    exit 1
                fi
                output_file="$2"
                shift 2
                ;;
            -*)
                log_error "Unknown option: $1"
                show_help >&2
                exit 1
                ;;
            *)
                if [[ -z "${target_file}" ]]; then
                    target_file="$1"
                    shift
                else
                    log_error "Unexpected additional argument: $1"
                    exit 1
                fi
                ;;
        esac
    done

    # Create bounded scratch workspace
    TMP_DIR="$(mktemp -d -t "prumo-safe-runner-XXXXXX")"
    log_info "Created scratch workspace at: ${TMP_DIR}"

    local scratch_manifest="${TMP_DIR}/manifest.txt"
    echo "task_id=sample" > "${scratch_manifest}"
    echo "timestamp=$(date -u +%s)" >> "${scratch_manifest}"

    if [[ ${verbose} -eq 1 ]]; then
        log_info "Scratch manifest content:"
        cat "${scratch_manifest}" >&2
    fi

    # Process sample data safely with arrays
    local -a items=("alpha" "beta" "gamma" "delta")
    local processed_count=0

    for item in "${items[@]}"; do
        echo "Processing: ${item}" >> "${TMP_DIR}/output.log"
        processed_count=$((processed_count + 1))
    done

    log_info "Successfully processed ${processed_count} items."

    if [[ -n "${output_file}" ]]; then
        cp -- "${TMP_DIR}/output.log" "${output_file}"
        log_info "Results saved to: ${output_file}"
    fi

    log_info "Task completed successfully."
}

main "$@"
