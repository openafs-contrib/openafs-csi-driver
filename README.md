## OpenAFS Container Storage Interface (CSI) Driver - README

### A. About Project
`OpenAFS Container Storage Interface` (CSI) driver allows OpenAFS to 
be used as persistent storage for stateful application running in Kubernetes 
clusters. Through this CSI Driver, Kubernetes persistent volumes (PVs) can 
be provisioned from OpenAFS. Thus, containers can be used with stateful 
microservices, such as database applications (MongoDB, PostgreSQL etc), web 
servers (nginx, apache), or any number of other containerized applications 
needing provisioned storage.

##### Supported Features of the CSI driver 
- `Openafs-csi-operator` ==> It takes care of starting/stopping of driver and reconcile it during faliures.
- `Static provisioning` ==>  Ability to use existing volumes/directories as persistent volumes.
- `Dynamic provisioning` ==> Ability to create persistent volume dynamically.

### B. Pre-requisites to run OpenAFS-CSI

- OpenAFS client should be installed on all worker nodes.

    - For OpenShift environment
        - In OpenShift environment all worker nodes are RHCOS which is kind of immutable OS.
        To install OpenAFS client on worker nodes we can use kmod-via-container framework,
        kindly follow steps mentioned at [kvc-openafs-kmod](https://github.com/openafs-contrib/kvc-openafs-kmod)

-   Supported Container Orchastrators (CO)
    ```sh
    Kubernetes >= 1.18
    OpenShift = 4.3  
    ```
currently we have tested driver on x86_64 

### C. Build and creation of Images

##### Pre-requisite:
- golang version > 1.14 
- operator-sdk version = v0.15.2
    - download operator-sdk [here](https://github.com/operator-framework/operator-sdk/releases/download/v0.15.2/operator-sdk-v0.15.2-x86_64-linux-gnu) 
    - mv operator-sdk-v0.15.2-x86_64-linux-gnu /usr/local/bin/operator-sdk
    - chmod +x /usr/local/bin/operator-sdk

##### Driver Image: 

1. Consider the workspace e.g. /vicepa/CSI/src (we can select our own workspace), export GOROOT and GOPATH as below
    ```sh
	export GOROOT=/usr/local/go (installed go path)
    export GOPATH=/vicepa/CSI/
    ```
2. cd to /vicepa/CSI/src and clone openafs-csi-driver repository
    ```sh
	cd /vicepa/CSI/src
	git clone git@github.com:openafs-contrib/openafs-csi-driver.git
    ```    

3. cd to /vicepa/CSI/src/openafs-csi-driver and run `make` command. This will build openafscsi plugin binary.


4. To create a docker image we need RHEL subscription (since we use registry.access.redhat.com/ubi8 image).

    - In OpenShift environment, subscription is taken from host itself, however in case it is not taken from host we need to uncomment below lines in Dockerfile. 
        ```dockerfile
        #ARG SUBS_USER
        #ARG SUBS_PASS
        #RUN subscription-manager register --username ${SUBS_USER} --password ${SUBS_PASS} --auto-attach
        ```

    - On systems where subscription is not taken from host and if we are using ubi8 image then we need to uncomment below lines in Dockerfile
        ```dockerfile
        #ARG SUBS_USER
        #ARG SUBS_PASS
        #RUN subscription-manager register --username ${SUBS_USER} --password ${SUBS_PASS} --auto-attach
        ```
        If we are using subscription-manager then we need to pass SUBS_USER and SUBS_PASS as build-args to docker build as mentioned in **step 7**

5. Inside Dockerfile, we clone OpenAFS master branch, so we need to get ssh private key from host. 
   Hence copy ~/.ssh/id_rsa to /vicepa/CSI/src/openafs-csi-driver. 

6. In case **git clone** for openafs repository does not work, follow below steps
	-  clone the openafs repository ([git@github.com:openafs/openafs.git](https://github.com/openafs/openafs)) on host
	-  copy openafs master repository source to /vicepa/CSI/src/openafs-csi-driver
	-  In Dockerfile, comment lines for git clone, id_rsa and ssh.  
	-  uncomment 'COPY openafs openafs'

7. Now build a docker image
- With subscription-manager inside container
   
    ```sh
    docker build --build-arg SUBS_USER=<username> --build-arg SUBS_PASS=<password> -topenafcsiplugin:latest .
    ```
    
- Without subscription-manager:
    ```
    docker build -topenafcsiplugin:latest .
    ```
8. Once `openafcsiplugin:latest` image is created, save the image in tar format and copy it on all worker nodes
    ```
    docker save openafcsiplugin:latest -o openafcsiplugin.tar
    ```
9. On all worker nodes load the image as below
    ```
    docker load < openafcsiplugin.tar
    ```
    
##### Operator Image:

1. Go inside operator directory *openafs-csi-driver/operator/src/openafs-csi-operator*

2. export GOPATH to openafs-csi-driver/operator, in our example it will be
    ```sh
    export GOPATH=/vicepa/CSI/src/openafs-csi-driver/operator
    ```
3. Build k8s api for operator 
    ```sh
    operator-sdk generate k8s
    ```
4. Build openafs operator image. Below step will create `openafsoperator:latest` image.
    ```sh
    operator-sdk build openafsoperator 
    ```
5. Save `openafsoperator:latest` image, copy and load it on all worker nodes using docker save/load commands as mentioned in **step 8-9** of driver image.



### D. Install and Deploy the OpenAFS CSI driver and Operator

1. Once images for driver and operator are installed on all worker nodes, go to *openafs-csi-driver/deploy* directory. 


2. To install a plugin, we need to have **ThisCell**, **CellServDB** and **krb5.conf** cell files for which driver will take tokens and create volumes.

    - For OpenShift environment
        make below changes before running setup.sh
        
        - In *openafs-csi-driver/deploy/CSI-Deploy/deploy/operator.yaml* file, change image to
	        ```sh
	        image: localhost/openafsoperator:latest
	        ```
        - In *openafs-csi-driver/deploy/CSI-Deploy/deploy/crds/openafscsi_cr.yaml* file, change pluginImage to
	        ```sh
	        pluginImage: localhost/openafscsiplugin:latest
	        ```
	    - In *openafs-csi-driver/deploy/CSI-Deploy/deploy/crds/openafscsi_cr.yaml* file,  change afsMount to
	        ```sh
	        afsMount: /var/afs
	        ```
	 *Note:* In case of using the registry, use the registry location for plugin and operator images
	 
   Run setup.sh as below,
    ```sh
    ./setup.sh -c <ThisCell location> -d <CellServDB location> -k <krb5.conf location> -n openafs
    ```
    Here,  '**-n openafs**' is for installing driver in "openafs" namespace.
3. Verify by checking pods in "openafs" namespace.
   (For successful installation, make sure attacher and provisioner statefulsets are READY.)
    ```
    kubectl -n openafs  get statefulset
    ```
   Also make sure OpenAFS CSI Driver daemonset is READY and it has started plugin on all woker nodes.
    ```sh
    kubectl -n openafs get daemonset 
    ```
4. For uninstalling a driver run below command
    ```sh
    ./setup.sh -n openafs -u
    ```
    `OpenAFS CSI Driver is now Installed and Running`.
    
### E. OpenAFS storage provisioning using OpenAFS CSI Driver 
There are two types of storage provisionings-
- **Dynamic Provisioning**
    > Dynamic provisioning is used to dynamically provisionthe storage backend volume based on the storageClass.

    Let's create secret and storageclass,
    
    **Creating Secret =>**
    > We need secret to create volume on AFS cell mentioned in storageclass yaml. Mainly we need to provide 'username/password' for cell where volume will be created and mount path where volume should be mounted. 

- Sample yaml file to create secret
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: <secret_name>     #This name will be used in storageclass yaml.
    data:
    #user of punetest cell 
      <cell name>_user: <base64 encoded username> 
      <cell name>_password: <base64 encoded password>
    #You can add more users in same format.
    ```
    Below is one example :
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: afs_prov_secret     #This name will be used in storageclass yaml.
    data:
    #user of punetest cell
      punetest.in.ibm.com_user: YWRtaW4K        #Provide base64 encoded user
      punetest.in.ibm.com_password: YWRtaW4K    #Provide base64 encoded password
    #You can add more users in same format.
    ```
    
    **Extra info:**
    Get base64 encoded username and password for cell
    ```sh
    echo <username/password> |base64
    ```
    Now, create secret using kubectl command
    ```sh
    kubectl create -f <secret yaml>
    ```
    
    **Note:** Naming of data field in secret should be maintained as
    ```
      <cell name>_user
      <cell name>_password
    ```

    **Storageclass ==>**
    >Storageclass defines what type of backend volume should be created by dynamic provisioning.

    Following parameter are supported by OpenAFS CSI Driver
    |Parameters | Description |
    |-----|------|
    | *cellname: |   AFS cell where AFS volume will be created. |
    |*server: |     AFS fileserver where volume will be created.|
    |*partition: |  vice partition where volume will be created.|
    |*volumepath: | AFS path with anyuser lookup permission [Ex."system:anyuser rl"]|
    |acl: |         List of ACLs to be set on dynamically provisioned volume. [Ex. system:anyuser rl smith write pat:friends rl].|
    |*csi.storage.k8s.io/provisioner-secret-name: | Name of the secret|
    |*csi.storage.k8s.io/provisioner-secret-namespace: | Namespace for secret.|
    '*' parameter is compulsary.  

  **Note:** 
        - 'volumepath' should be accessible from all worker nodes  without any tokens. New volumes will be mounted in this path.
        - For this version of plugin 'secret-namespace' should be kept as 'default'.

  For sample, kindly refer "examples/csi-storageclass.yaml" for storageClass 
  For PersistentVolumeClaim and POD using PVC refer "examples/csi-pvc.yaml", "examples/csi-app.yaml".


- **Static Provisioning:**
    > In static provisioning, the backend storage volumes and PVs are created by the administrator. Static provisioning can be used to provision a already created AFS volume.
    
    > Lets explain this with example. Consider we have user volume "usr.smith" in cell "samplecell.in.ibm.com". Mount point for this volume can be like "/afs/samplecell.in.ibm.com/usr/smith". Now to access this volume we need user "smith" tokens.
    
    > We need one path in AFS space which can be accessed without tokens, normally mount with "system:anyuser rl" access. Consider that path is "/afs/samplecell.in.ibm.com/GlobalPath". 
    
    > So as a admin we will create a mount for "usr.smith" volume inside "/afs/samplecell.in.ibm.com/GlobalPath", something as "/afs/samplecell.in.ibm.com/GlobalPath/smithVolume".  Do note without tokens we can see path "/afs/samplecell.in.ibm.com/GlobalPath/smithVolume", but to access "/afs/samplecell.in.ibm.com/GlobalPath/smithVolume" we need user "smith" tokens. 

    **With above setup admin will create PersistentVolume yaml file as below:**


    ```yaml
    apiVersion: v1
    kind: PersistentVolume
    metadata:
        name: openafs-static-pv
    spec:
      capacity:
        storage: 1Gi
      accessModes:
    - ReadWriteMany
      csi:
        driver: openafs.csi.ibm.com
        volumeHandle: "/afs/samplecell.in.ibm.com/GlobalPath/smithVolume"
    ```
    
    Now admin will create PersistentVolume using 
    ```sh
    kubectl create -f <PersistentVolume yaml file>
    ```

    With above steps "openafs-static-pv" PersistentVolume should get created. Now we can create a PersistentVolumeClaim to bound to above PersistentVolume and later we can use PersistentVolumeClaim inside a POD.

    For sample refer Sample/static-pv.yml, Sample/static-pvc.yml

    **Using PersistentVolumeClaim Inside POD:**

    There are sample for POD using PersistentVolumeClaim inside Sample/csi-app.yaml. Do note CSI Plugin will mount a volume inside a POD, but to access a volume Container Application need to get tokens, which has to be done by the Application.


