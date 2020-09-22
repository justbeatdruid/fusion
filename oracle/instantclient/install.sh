#!/bin/bash

set -ex

if test -d instantclient_12_2;then rm -rf instantclient_12_2;fi

unzip instantclient-basic-linux.x64-12.2.0.1.0.zip
unzip instantclient-sdk-linux.x64-12.2.0.1.0.zip
mv instantclient_12_2 /usr/lib

ln -s /usr/lib/instantclient_12_2/libclntsh.so.12.1 /usr/lib/libclntsh.so
ln -s /usr/lib/instantclient_12_2/libocci.so.12.1 /usr/lib/libocci.so
ln -s /usr/lib/instantclient_12_2/libociei.so /usr/lib/libociei.so
ln -s /usr/lib/instantclient_12_2/libnnz12.so /usr/lib/libnnz12.so

if ! test -d /usr/lib/pkg-config;then mkdir /usr/lib/pkg-config;fi

cat <<EOF> /usr/lib/pkg-config/oci8.pc
# Package Information for pkg-config

prefix=/usr/lib/instantclient_12_2
exec_prefix=\${prefix}
libdir=\${exec_prefix}
includedir=${prefix}/sdk/include/

glib_genmarshal=glib-genmarshal
gobject_query=gobject-query
glib_mkenums=glib-mkenums

Name: oci8
Description: oci8 library
Version: 12.2
Libs: -L\${libdir} -lclntsh
Cflags: -I\${includedir}
EOF

# add to /etc/profile
#
export ORACLE_HOME=/usr/lib/instantclient_12_2
export LD_LIBRARY_PATH=$ORACLE_HOME
export PKG_CONFIG_PATH=/usr/lib/pkg-config
#
