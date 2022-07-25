set -u
set -o privileged
set -o pipefail
set -o errtrace

if [ -z "${BASH_VERSION:=}" ] ; then
    echo "$0 depends on bash-specific features. Please install bash v4 or newer."
    exit 1
fi

_ERR_FUNCS=()
_err_cascade () {
    err_status=$1
    trap - ERR
    local i
    for (( i = ${#_ERR_FUNCS[@]} - 1 ; i >= 0 ; i-- )) ; do
	${_ERR_FUNCS[$i]}
    done

    exit $err_status
}
trap '_err_cascade $?' ERR

aterr () {
    _ERR_FUNCS+=("$@")
}

_err_report () {
    cat 1>&2 <<EOF
problem executing commands:
EOF
    local i=0
    while caller $i 1>&2; do
	((i += 1))
    done
}

aterr '_err_report'

_EXIT_FUNCS=()
_exit_cascade () {
    local i
    for (( i = ${#_EXIT_FUNCS[@]} - 1 ; i >= 0 ; i-- )) ; do
	${_EXIT_FUNCS[$i]}
    done
}
trap '_exit_cascade' EXIT

atexit () {
    _EXIT_FUNCS+=("$@")
}

__CLEANUP_FILES=()
__CLEANUP_DIRS=()

__cleanup () {
    [[ ${#__CLEANUP_FILES[@]} -gt 0 ]] && rm -f "${__CLEANUP_FILES[@]}"
    rm -rf "${__CLEANUP_DIRS[@]}"
}

atexit __cleanup

__TESTING_FUNCS=()

testing::register () {
    __TESTING_FUNCS+=("$*")
}

testing::run () {
    local count=${#__TESTING_FUNCS[@]}
    local res
    echo 1..$count

    for (( i = 0; i < $count ; i++ )) ; do
	local testnum=$(( $i + 1 ))
	if res="$(${__TESTING_FUNCS[$i]})" ; then
	    echo "ok $testnum - $res"
	else
	    echo "not ok $testnum - $res"
	fi
    done
}

testing::desc () {
    echo "$@"
}

testing::todo () {
    testing::desc "# TODO $*"
}

testing::skip () {
    testing::desc "# SKIP $*"
}

testing::skip-all () {
    echo "1..0 # skip $*"
    exit 0
}

testing::bail () {
    echo "Bail out! $*"
}

# Start off with logging going to stderr.
_LOGFD=2
log::cmd () {
    local ret
    ret=0
    log::debug "running command: $*"
    "$@" || ret=$?
    log::debug "return value was $ret"
    return $ret
}

log::debug () {
    $DEBUG || return 0
    echo "DEBUG: $*" >&${_LOGFD}
    true
}

log::warning () {
    echo "WARNING: $*" >&${_LOGFD}
}

log::error () {
    echo "ERROR: $*" >&${_LOGFD}
}

log::fatal () {
    echo "FATAL: $*" >&${_LOGFD}
    exit 1
}

log::info () {
    echo "INFO: $*" >&${_LOGFD}
}

TMPDIR=$(mktemp -d) || log::fatal "could not create directory for temporary files"
export TMPDIR
__CLEANUP_DIRS+=("$TMPDIR")

temp::file () {
    mktemp -p "$TMPDIR" "$@"
}

temp::dir () {
    mktemp -d -p "$TMPDIR" "$@"
}
