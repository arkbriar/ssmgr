#!/bin/bash

set -e

if [ -z $(which openssl) ]; then
    echo "Error: openssl not found! Install it before running this script."
    exit 1
fi

if [ \( "$(uname -s)" = "Darwin" -a "$(which getopt)" = "/usr/bin/getopt" \) ]; then
    echo "Install GNU getopt first: brew install gnu-getopt"
    exit 3
fi

SHORT_OPTS=i:,n:,d:,o:,c:,s:,v,h
LONG_OPTS=ip:,numbits:,days:,output:,ca:,serial:,verbose,help

usage() {
    echo """Usage: $0
    -i/--ip <ip, default 127.0.0.1>
    -o/--output <output dir, default .>
    -n/--numbits <bits, default 2048>
    -d/--days <days, default 365>
    -c/--ca <ca dir(storing ca.pem & ca.key), default .>
    -s/--serial <ca serial file, default \"\">
    -v/--verbose"""
}

# params parsed from opts, initialize with their defaults values

IP=127.0.0.1
OUT_DIR=.
NUMBITS=2048
DAYS=365
CA_DIR=.
CA_SERIAL=
VERBOSE=0

echo_params() {
    echo """Generating self signed certificates with the following params:
    echo 
    ip: $IP
    output dir: $OUT_DIR
    numbits: $NUMBITS
    days: $DAYS
    ca dir: $CA_DIR"""
}

parse_params() {
    # -temporarily store output to be able to check for errors
    # -activate advanced mode getopt quoting e.g. via “--options”
    # -pass arguments only via   -- "$@"   to separate them correctly
    PARSED=$(getopt --options $SHORT_OPTS --longoptions $LONG_OPTS --name "$0" -- "$@")
    if [[ $? -ne 0 ]]; then
        exit 2
    fi

    # use eval with "$PARSED" to properly handle the quoting
    eval set -- "$PARSED"

    while true; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=1
                shift
                ;;
            -i|--ip)
                IP=$2
                shift 2
                ;;
            -o|--output)
                OUT_DIR=$2
                shift 2
                ;;
            -n|--numbits)
                NUMBITS=$2
                shift 2
                ;;
            -d|--days)
                DAYS=$2
                shift 2
                ;;
            -c|--ca)
                CA_DIR=$2
                shift 2
                ;;
            --) # break here
                shift
                break
                ;;
            *)
                echo "Error: unrecognized option $1"
                exit 1
        esac
    done
}

detect_ca_files() {
    ca_dir=$1
    if [ ! -d $ca_dir ]; then
        echo "Error: $ca_dir is not a directory."
        exit 1
    fi
    if [ ! -e $ca_dir/ca.pem ]; then
        echo "Error: missing file $ca_dir/ca.pem"
        exit 1
    fi
    if [ ! -e $ca_dir/ca.key ]; then
        echo "Error: missing file $ca_dir/ca.key"
        exit 1
    fi
}

generate_key() {
    numbits=$1
    outfile=$2

    echo "Generating server key ($numbits bits), saving to $outfile"
    echo

    openssl genrsa -out $outfile $numbits
}

sign_key() {
    ip=$1
    days=$2
    key_file=$3
    ca_cert=$4
    ca_key=$5
    outfile=$6
    ca_serial_opt=$7

    echo "Generating server cert ($days days, $ip), saving to $outfile"
    echo

    tmpdir=/tmp/ssl_gen_$(head -c 40 /dev/urandom | base64 -i - | sed "s/\\//_/g")
    mkdir -p $tmpdir
    echo subjectAltName = IP:$ip > $tmpdir/extfile.cnf
    openssl req -new -key $key_file -out $tmpdir/server.csr
    # openssl x509 -req -days $days -in $tmpdir/server.csr -CA $ca_cert -CAkey $ca_key <-CAserial srlfile> -out $outfile -extfile $tmpdir/extfile.cnf
    openssl x509 -req -days $days -in $tmpdir/server.csr -CA $ca_cert -CAkey $ca_key $ca_serial_opt -out $outfile -extfile $tmpdir/extfile.cnf
    rm -r $tmpdir
}

# main part

parse_params $@

detect_ca_files $CA_DIR

mkdir -p $OUT_DIR

if [ 1 -eq $VERBOSE ]; then
    echo_params
    set -vx
fi

# generate server key
IP_U=$(echo $IP | sed 's/\./_/g')
SERVER_KEY=$OUT_DIR/server_$IP_U.key
SERVER_CERT=$OUT_DIR/server_$IP_U.pem

generate_key $NUMBITS $SERVER_KEY
if [ $? -ne 0 ]; then
    echo "Failed"
    exit 1
fi

# check serial file, and use $CA_DIR/ca.srl if serial file is not specified or does
# not exists

CA_SERIAL_OPT=-CAcreateserial
if [ \(! -e $CA_SERIAL\) ]; then
    echo "Warn: serial file does not exists"
    echo
elif [ ! -z $CA_SERIAL ]; then
    CA_SERIAL_OPT=-CAserial $CA_SERIAL
fi

if [ "$CA_SERIAL_OPT" = "-CAcreateserial" ]; then
    if [ -e $CA_DIR/ca.srl ]; then
        echo "Info: using serial file $CA_DIR/ca.srl"

        CA_SERIAL_OPT="-CAserial $CA_DIR/ca.srl"
    elif [ -e .srl ]; then
        echo "Info: using serial file .srl"

        CA_SERIAL_OPT="-CAserial .srl"
    fi
fi


# sign server key

echo
sign_key $IP $DAYS $SERVER_KEY $CA_DIR/ca.pem $CA_DIR/ca.key $SERVER_CERT "$CA_SERIAL_OPT"
if [ $? -ne 0 ]; then
    echo "Failed"
    exit 1
fi

echo
echo "Generated, saved in $OUT_DIR"

exit 0
