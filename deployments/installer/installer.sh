#!/bin/bash

## Configuration variables. 
NAME="xlist"
RELEASE="RELEASE"
ARCH="${ARCH:-amd64}"

## Base dirs
BIN_DIR=/usr/local/bin
ETC_DIR=/etc/luids
VAR_DIR=/var/lib/luids
CACHE_DIR=/var/cache/luids
SYSTEMD_DIR=/etc/systemd/system

## Service variables
SVC_USER=lu${NAME}
SVC_GROUP=luids

## Binaries
BINARIES="xlistd xlistc xlget"

## Download
DOWNLOAD_BASE="https://github.com/luids-io/${NAME}/releases/download"
DOWNLOAD_URI="${DOWNLOAD_BASE}/${RELEASE}/${NAME}_${RELEASE}_linux_${ARCH}.tgz"

##

die() { echo "error: $@" 1>&2 ; exit 1 ; }

## some checks
for deps in "wget" "mktemp" "getent" "useradd" "groupadd" ; do
	which $deps >/dev/null \
		|| die "$deps is required!"
done
[[ $EUID -eq 0 ]] || die "This script must be run as root"
[ -d $BIN_DIR ] || die "Binary directory $BIN_DIR doesn't exist"

## options command line
OPT_UNATTEND=0
OPT_OVERWRITE_BIN=0
while [ -n "$1" ]; do
	case "$1" in
		-u) OPT_UNATTEND=1 ;;
		-o) OPT_OVERWRITE_BIN=1 ;; 
		-h) echo -e "Options:\n\t [-u] unattend\n\t [-o] overwrite binaries\n"
		    exit 0 ;; 
 		*) die "Option $1 not recognized" ;; 
	esac
	shift
done

echo
echo "======================"
echo "- luIDS installer:"
echo "   ${NAME} ${RELEASE}"
echo "======================"
echo

show_actions() {
	echo "Warning! This script will commit the following changes to your system:"
	echo ". Download and install binaries in '${BIN_DIR}'"
	echo ". Create system group '${SVC_GROUP}'"
	echo ". Create system user '${SVC_USER}'"
	echo ". Create data dirs in '${VAR_DIR}'"
	echo ". Create cache dirs in '${CACHE_DIR}'"
	echo ". Create config dirs in '${ETC_DIR}'"
	[ -d $SYSTEMD_DIR ] && echo ". Copy systemd configurations to '${SYSTEMD_DIR}'"
	echo ""
}

if [ $OPT_UNATTEND -eq 0 ]; then
	show_actions
	read -p "Are you sure? (y/n) " -n 1 -r
	echo
	echo
	if [[ ! $REPLY =~ ^[Yy]$ ]]
	then
		die "canceled"
	fi
fi

TMP_DIR=$(mktemp -d -t ins-XXXXXX) || die "couldn't create temp"
LOG_FILE=${TMP_DIR}/installer.log

log() { echo `date +%y%m%d%H%M%S`": $@" >>$LOG_FILE ; }
step() { echo -n "* $@..." ; log "STEP: $@" ; }
step_ok() { echo " OK" ; }
step_err() { echo " ERROR" ; }
user_exists() { getent passwd $1>/dev/null ; }
group_exists() { getent group $1>/dev/null ; }

## do functions
do_download() {
	[ $# -eq 2 ] || die "${FUNCNAME}: unexpected number of params"
	local url="$1"
	local filename="$2"

	local dst="${TMP_DIR}/${filename}"
	rm -f $dst
	log "downloading $url"
	echo "$url" | grep -q "^\(http\|ftp\)"
	if [ $? -eq 0 ]; then
		wget "$url" -O $dst &>>$LOG_FILE
	else
		cp -v "$url" $dst &>>$LOG_FILE
	fi
}

do_clean_file() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local filename=$1

	local src="${TMP_DIR}/${filename}"
	log "clearing $src"    
	rm -f $src &>>$LOG_FILE
}

do_install_bin() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local binary=$1

	local src="${TMP_DIR}/${binary}"
	local dst="${BIN_DIR}/${binary}"
	[ ! -f $src ] && log "$src not found!" && return 1

	log "copying $src to $dst, chown root, chmod 755"
	{ cp $src $dst \
		&& chown root:root $dst \
		&& chmod 755 $dst
	} &>>$LOG_FILE
}

do_setcap_net_admin() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local binary=$1

	local fpath="${BIN_DIR}/${binary}"
	[ ! -f $fpath ] && log "$fpath not found!" && return 1

	log "set net_admin capability to $fpath"
	setcap CAP_NET_ADMIN=+eip $fpath &>>$LOG_FILE
}


do_setcap_bind() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local binary=$1

	local fpath="${BIN_DIR}/${binary}"
	[ ! -f $fpath ] && log "$fpath not found!" && return 1

	log "set bind capability to $fpath"
	setcap CAP_NET_BIND_SERVICE=+eip $fpath &>>$LOG_FILE
}

do_unpackage() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local tgzfile=$1
	
	local src="${TMP_DIR}/${tgzfile}"
	[ ! -f $src ] && log "${FUNCNAME}: $src not found!" && return 1

	log "unpackaging $tgzfile"
	tar -zxvf $src -C $TMP_DIR &>>$LOG_FILE
}

do_create_dir() {
	[ $# -ge 1 ] || die "${FUNCNAME}: unexpected number of params"
	local datadir=$1
	local datagrp=root
	if [ $# -ge 2 ]; then
		datagrp=$2
	fi
	local perm=755
	if [ $# -ge 3 ]; then
		perm=$3
	fi

	[ -d $datadir ] && log "$datadir found!" && return 1
	group_exists $datagrp || { log "group $datagrp doesn't exists" && return 1 ; }

	log "creating dir $datadir, chgrp to $datagrp, chmod $perm"
	{ mkdir -p $datadir \
		&& chown root:$datagrp $datadir \
		&& chmod $perm $datadir
	} &>>$LOG_FILE
}

do_create_sysgroup() {
	[ $# -eq 1 ] || die "${FUNCNAME}: unexpected number of params"
	local ngroup=$1

	group_exists $ngroup && log "group $ngroup already exists" && return 1

	log "groupadd $ngroup with params"
	groupadd -r $ngroup &>>$LOG_FILE
}

do_create_sysuser() {
	[ $# -ge 2 ] || die "${FUNCNAME}: unexpected number of params"
	local nuser=$1
	local nhome=$2
	local ngroup=""
	if [ $# -ge 3 ]; then
		ngroup="$3"
	fi

	user_exists $nuser && log "user $nuser already exists" && return 1
	if [ "$ngroup" == "" ]; then
		log "useradd $nuser as system user"
		useradd -s /usr/sbin/nologin -r -M -d "$nhome" $nuser &>>$LOG_FILE
	else
		log "useradd $nuser as system user with group $ngroup"
		useradd -s /usr/sbin/nologin -r -M -d "$nhome" -g $ngroup $nuser &>>$LOG_FILE
	fi
}

## steps
install_binaries() {
	step "Downloading and installing binaries"

	if [ $OPT_OVERWRITE_BIN -eq 0 ]; then
		for binary in $BINARIES; do
			if [ -f ${BIN_DIR}/$binary ]; then
				log "${BIN_DIR}/${binary} already exists, skip download"
				step_ok
				return 0
			fi
		done
	fi

	## download
	do_download "$DOWNLOAD_URI" ${NAME}_linux.tgz
	[ $? -ne 0 ] && step_err && return 1
	do_unpackage ${NAME}_linux.tgz
	[ $? -ne 0 ] && step_err && return 1
	do_clean_file ${NAME}_linux.tgz

	## deploy binaries
	for binary in $BINARIES; do
		do_install_bin $binary
		[ $? -ne 0 ] && step_err && return 1
        	do_clean_file $binary
	done

	step_ok
}

install_helper_scripts() {
	step "Installing helper scripts"

	log "creating ${BIN_DIR}/xlget_check"
	{ cat > ${BIN_DIR}/xlget_check <<EOF
#!/bin/bash

XLGET=${BIN_DIR}/xlget
EXECUSER=${SVC_USER}
CONFIG=${ETC_DIR}/${NAME}/xlget.toml

die() { echo "error: \$@" 1>&2 ; exit 1; }

# some checks...
[[ \$EUID -eq 0 ]] || die "This script must be run as root"
[ -f \$XLGET ] || die "\$XLGET doesn't exists!"
getent passwd \$EXECUSER >/dev/null
[ \$? -eq 0 ] || die "User \$EXECUSER doesn't exist!"

su \$EXECUSER -p -c "\$XLGET --dry-run --config \$CONFIG"
EOF
	chown root:root ${BIN_DIR}/xlget_check
	chmod 755 ${BIN_DIR}/xlget_check
	} &>>$LOG_FILE
	[ $? -ne 0 ] && step_err && return 1

	log "creating ${BIN_DIR}/xlget_update"
	{ cat > ${BIN_DIR}/xlget_update <<EOF
#!/bin/bash

XLGET=${BIN_DIR}/xlget
EXECUSER=${SVC_USER}
CONFIG=${ETC_DIR}/${NAME}/xlget.toml

die() { echo "error: \$@" 1>&2 ; exit 1; }

# some checks...
[[ \$EUID -eq 0 ]] || die "This script must be run as root"
[ -f \$XLGET ] || die "\$XLGET doesn't exists!"
getent passwd \$EXECUSER >/dev/null
[ \$? -eq 0 ] || die "User \$EXECUSER doesn't exist!" 

su \$EXECUSER -p -c "\$XLGET --config \$CONFIG"
EOF
	chown root:root ${BIN_DIR}/xlget_update
	chmod 755 ${BIN_DIR}/xlget_update
	} &>>$LOG_FILE
	[ $? -ne 0 ] && step_err && return 1

	step_ok
}

create_system_group() {
	step "Creating system group"

	group_exists $SVC_GROUP \
		&& log "group $SVC_GROUP already exists" && step_ok && return 0
	do_create_sysgroup $SVC_GROUP
	[ $? -ne 0 ] && step_err && return 1
	
	step_ok
}

create_system_user() {
	step "Creating system user"

	user_exists $SVC_USER \
		&& log "user $SVC_USER already exists" && step_ok && return 0
	do_create_sysuser $SVC_USER $VAR_DIR $SVC_GROUP
	[ $? -ne 0 ] && step_err && return 1

	step_ok
}

create_data_dir() {
	step "Creating data dirs"

	if [ ! -d $VAR_DIR ]; then
		do_create_dir $VAR_DIR
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$VAR_DIR already exists"
	fi

	if [ ! -d $VAR_DIR/$NAME ]; then
		do_create_dir $VAR_DIR/$NAME $SVC_GROUP 1770
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$VAR_DIR/$NAME already exists"
	fi

	step_ok
}

create_cache_dir() {
	step "Creating cache dirs"

	if [ ! -d $CACHE_DIR ]; then
		do_create_dir $CACHE_DIR
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$CACHE_DIR already exists"
	fi

	if [ ! -d $CACHE_DIR/$NAME ]; then
		do_create_dir $CACHE_DIR/$NAME $SVC_GROUP 1770
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$CACHE_DIR/$NAME already exists"
	fi

	step_ok
}

create_base_config() {
	step "Creating base config"

	## create dirs
	if [ ! -d $ETC_DIR ]; then
		do_create_dir $ETC_DIR
		[ $? -ne 0 ] && step_err && return 1

		local ssldir="${ETC_DIR}/ssl"
		do_create_dir $ssldir/certs
		[ $? -ne 0 ] && step_err && return 1
		do_create_dir $ssldir/private $SVC_GROUP 750
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR already exists"
	fi

	## create files
	if [ ! -f $ETC_DIR/services.json ]; then
		log "creating $ETC_DIR/services.json"
		echo "[ ]" > $ETC_DIR/services.json
	else
		log "$ETC_DIR/services.json already exists"
	fi

	step_ok
}

create_service_config() {
	step "Creating service config"

	## create dirs
	if [ ! -d $ETC_DIR/$NAME ]; then
		do_create_dir $ETC_DIR/$NAME
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/$NAME already exists"
	fi

	## create files
	if [ ! -f $ETC_DIR/$NAME/xlistd.toml ]; then
		log "creating $ETC_DIR/$NAME/xlistd.toml"
		{ cat > $ETC_DIR/$NAME/xlistd.toml <<EOF
[xlist]
certsdir   = "${ETC_DIR}/ssl"
datadir    = "${VAR_DIR}/${NAME}"

[xlist.service]
files      = [ "${ETC_DIR}/${NAME}/services.json" ]

EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/$NAME/xlistd.toml already exists"
	fi

	if [ ! -f $ETC_DIR/$NAME/services.json ]; then
		log "creating $ETC_DIR/$NAME/services.json"
		echo '[{"id":"root","class":"mock"}]' > $ETC_DIR/$NAME/services.json 2>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/$NAME/services.json already exists"
	fi

	if [ ! -f $ETC_DIR/$NAME/xlget.toml ]; then
		log "creating $ETC_DIR/$NAME/xlget.toml"
		{ cat > $ETC_DIR/$NAME/xlget.toml <<EOF
[xlget]
outputdir = "${VAR_DIR}/${NAME}"
cachedir  = "${CACHE_DIR}/${NAME}"

[xlget.source]
files     = [ "${ETC_DIR}/${NAME}/sources.json" ]

EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/$NAME/xlget.toml already exists"
	fi

	if [ ! -f $ETC_DIR/$NAME/sources.json ]; then
		log "creating $ETC_DIR/$NAME/sources.json"
		echo '[ ]' > $ETC_DIR/$NAME/sources.json 2>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$ETC_DIR/$NAME/sources.json already exists"
	fi

	step_ok
}

install_systemd_services() {
	step "Installing systemd services"
	if [ ! -f $SYSTEMD_DIR/luids-xlistd.service ]; then
		log "creating $SYSTEMD_DIR/luids-xlistd.service"
		{ cat > $SYSTEMD_DIR/luids-xlistd.service <<EOF
[Unit]
Description=xlistd service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=1
User=$SVC_USER
ExecStart=$BIN_DIR/xlistd --config $ETC_DIR/$NAME/xlistd.toml

[Install]
WantedBy=multi-user.target
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$SYSTEMD_DIR/luids-xlistd.service already exists"
	fi

	if [ ! -f $SYSTEMD_DIR/luids-xlistd@.service ]; then
		log "creating $SYSTEMD_DIR/luids-xlistd@.service"
		{ cat > $SYSTEMD_DIR/luids-xlistd@.service <<EOF
[Unit]
Description=xlistd service per-config file
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=1
User=$SVC_USER
ExecStart=$BIN_DIR/xlistd --config $ETC_DIR/$NAME/%i.toml

[Install]
WantedBy=multi-user.target
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$SYSTEMD_DIR/luids-xlistd@.service already exists"
	fi

	if [ ! -f $SYSTEMD_DIR/luids-xlget.service ]; then
		log "creating $SYSTEMD_DIR/luids-xlget.service"
		{ cat > $SYSTEMD_DIR/luids-xlget.service <<EOF
[Unit]
Description=xlget service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=1
User=$SVC_USER
Environment=GOMAXPROCS=2
ExecStart=$BIN_DIR/xlget --auto --config $ETC_DIR/$NAME/xlget.toml

[Install]
WantedBy=multi-user.target
EOF
		} &>>$LOG_FILE
		[ $? -ne 0 ] && step_err && return 1
	else
		log "$SYSTEMD_DIR/luids-xlget.service already exists"
	fi

	step_ok
}

## main process
install_binaries || die "Show $LOG_FILE"
create_system_group || die "Show $LOG_FILE"
create_system_user || die "Show $LOG_FILE"
create_data_dir || die "Show $LOG_FILE"
create_cache_dir || die "Show $LOG_FILE"
create_base_config || die "Show $LOG_FILE"
create_service_config || die "Show $LOG_FILE"
[ -d $SYSTEMD_DIR ] && { install_systemd_services || die "Show $LOG_FILE for details." ; }
install_helper_scripts || die "Show $LOG_FILE"

echo
echo "Installation success!. You can see $LOG_FILE for details."
