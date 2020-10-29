#!/usr/bin/env bash

oc_env=0

function usage() {

	echo "$1 -c <ThisCell loc> -d <CellServDB loc> -k <krb5.conf loc> -n <Namespace> [-u] <Uninstall> [-i] <Install>"
	exit 0	
}

function Uninstall() {
	Namespace=$1
	kubectl -n $Namespace delete -f CSI-Deploy/deploy/crds/openafscsi_cr.yaml
	kubectl -n $Namespace delete -f CSI-Deploy/deploy/operator.yaml
	kubectl -n $Namespace delete -f CSI-Deploy/deploy/crds/openafscsi_crd.yaml
	cd CSI-Deploy/RBAC/ && ls |xargs -n1 kubectl -n $Namespace delete -f
	cd ../../
	kubectl -n $Namespace delete cm afsconfig
	kubectl delete ns $Namespace

}


function ChangeNS() {

	for ent in `ls CSI-Deploy/RBAC/`; do
		sed -i "s/<NAMESPACE>/${1}/g" CSI-Deploy/RBAC/$ent
	done
	sed -i "s/<NAMESPACE>/${1}/g"  CSI-Deploy/deploy/crds/openafscsi_cr.yaml
	sed -i "s/<NAMESPACE>/${1}/g"  CSI-Deploy/deploy/crds/openafscsi_cr.yaml
	sed -i "s/<NAMESPACE>/${1}/g"  CSI-Deploy/deploy/crds/openafscsi_cr.yaml


}

function Install() {

	Namespace=$1
	kubectl create ns $Namespace
	kubectl -n $Namespace create cm afsconfig --from-file=$ThisCell --from-file=$CellServDB --from-file=$KRB5
	cd CSI-Deploy/RBAC/ && ls |xargs -n1 kubectl -n $Namespace create -f
	cd ../../ 
        if [ $oc_env -eq 1 ]; then
	        oc adm policy add-scc-to-user privileged system:serviceaccount:$Namespace:openafs-csi-node
	        oc adm policy add-scc-to-user privileged system:serviceaccount:$Namespace:openafs-csi-attacher
	        oc adm policy add-scc-to-user privileged system:serviceaccount:$Namespace:openafs-csi-provisioner
        fi
	kubectl -n $Namespace create -f CSI-Deploy/deploy/crds/openafscsi_crd.yaml
	kubectl -n $Namespace  create -f CSI-Deploy/deploy/operator.yaml
	kubectl -n $Namespace create -f CSI-Deploy/deploy/crds/openafscsi_cr.yaml
}


ThisCell=""
CellServDB=""
KRB5=""
Install=0
Uninstall=0
Namespace=""
while getopts "c:d:k:iun:h" OPTIONS; do

	case $OPTIONS in
		c) ThisCell=$OPTARG;;
		d) CellServDB=$OPTARG;;
		k) KRB5=$OPTARG;;
		i) Install=1;;
		u) Uninstall=1;;
		n) Namespace=$OPTARG;;
		h) usage $0;;
		
	esac

done

[[ -z $Namespace ]] && Namespace="default"

if [ $Uninstall -eq 1 ]; then
    Uninstall $Namespace
    exit 0
fi


[[ -z $ThisCell ]] && usage $0
[[ -z $CellServDB ]] && usage $0
[[ -z $KRB5 ]] && usage $0

oc version &>/dev/null
if [ $? -eq 0 ]; then
	oc_env=1
fi

if [ $Install -eq 1 ]; then
        ChangeNS $Namespace
	Install $Namespace
else
        ChangeNS $Namespace
	Uninstall $Namespace
	Install $Namespace
fi
