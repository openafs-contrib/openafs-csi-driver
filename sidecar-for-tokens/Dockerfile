FROM registry.access.redhat.com/ubi8:latest
MAINTAINER "Yadavendra Yadav"
WORKDIR /build/
#ARG SUBS_USER
#ARG SUBS_PASS
#RUN subscription-manager register --username ${SUBS_USER} --password ${SUBS_PASS} --auto-attach

# First update the base container to latest versions of everything
RUN yum update -y

RUN yum install -y krb5-workstation
RUN yum install -y kernel-devel kernel-headers make gcc bison flex byacc libtool 
RUN yum install -y util-linux
RUN yum install -y bash git which krb5-devel
COPY id_rsa .
RUN eval $(ssh-agent -s) && \
    ssh-add id_rsa && \
    ssh-keyscan -H github.com >> /etc/ssh/ssh_known_hosts && \
    git clone git@github.com:openafs/openafs.git
#COPY ./openafs openafs
RUN cd openafs && sh regen.sh && ./configure && make install
RUN cp openafs/lib/libafshcrypto* /lib64/
RUN cp openafs/lib/librokenafs.* /lib64/
RUN mkdir -p /usr/local/etc/openafs/
ADD rekinit.sh /rekinit.sh
RUN chmod 777 /rekinit.sh
