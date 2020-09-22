#!/bin/bash

set -x

rm -rf /usr/lib/instantclient_12_2

rm /usr/lib/libclntsh.so
rm /usr/lib/libocci.so
rm /usr/lib/libociei.so
rm /usr/lib/libnnz12.so

rm -rf /usr/lib/pkg-config/oci8.pc
