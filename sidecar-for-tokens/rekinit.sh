#!/usr/bin/env bash

TimeToSleep=${REFRESH_TIME}

username=$(cat /etc/afs/username)
password=$(cat /etc/afs/password)

CellName=${CELLNAME}
if [[ -z ${CellName} ]]; then
    CellName=$(cat /usr/local/etc/openafs/)
fi
cp /etc/config/ThisCell /usr/local/etc/openafs/
cp /etc/config/CellServDB /usr/local/etc/openafs/
cp /etc/config/krb5.conf /etc/
u_cell=$(echo $CellName | tr '[:lower:]' '[:upper:]')
[[ -z ${TimeToSleep} ]] && TimeToSleep=86400

export KRB5CCNAME=DIR:/tmp
while [ true ]; do

        echo ${password} | kinit ${username}@${u_cell}
        klist
	aklog -c ${CellName} -k ${u_cell}
        sleep ${TimeToSleep}
done
