FROM registry.access.redhat.com/ubi8:latest
LABEL maintainers="Yadavendra Yadav"
LABEL description="OpenAFS CSI Plugin"
WORKDIR /root/OpenAFSBuild
#Uncomment below lines if we want to add subscription information. 
#This is needed in case subscription is not taken from host.
#ARG SUBS_USER
#ARG SUBS_PASS
#RUN subscription-manager register --username ${SUBS_USER} --password ${SUBS_PASS} --auto-attach
RUN yum install -y kernel-devel kernel-headers make gcc bison flex byacc libtool
RUN yum install -y util-linux
RUN yum install -y bash git which krb5-devel
#For cases where git clone does work, clone openafs repo on host and remove id_rsa, ssh and git clone step.
#Once clone of openafs repo is done copy it in current workspace and uncomment "COPY ./openafs openafs"
COPY id_rsa .
RUN eval $(ssh-agent -s) && \
    ssh-add id_rsa && \
    ssh-keyscan -H github.com >> /etc/ssh/ssh_known_hosts && \
    git clone git@github.com:openafs/openafs.git
RUN rm -rf id_rsa
#COPY ./openafs openafs
RUN cd openafs && sh regen.sh && ./configure && make install
RUN cp openafs/lib/libafshcrypto* /lib64/ 
RUN cp openafs/lib/librokenafs.* /lib64/
RUN mkdir -p /usr/local/etc/openafs/
COPY ./bin/openafsplugin /openafsplugin
ENTRYPOINT ["/openafsplugin"]
