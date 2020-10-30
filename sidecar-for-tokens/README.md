## SideCar container for AFS tokens - README
    
OpenAFS CSI driver can create PV using dynamic & static provisioning. However to access the volumes inside PODs we need to have appropriate tokens. One way is to get tokens inside microservice image itself, but for that we need to add OpenAFS and krb5 binaries inside microservice image. In order to avoid having OpenAFS bins inside microservice, we can use this sidecar container, which will have all OpenAFS binaries with it and will get UID based tokens, this will avoid changing the images for Microservice.  AFS mainly supports PAG and UID based tokens, this sidecar container provides a way to get UID based tokens. 

### Building Image:

- For building image run below command. (For subscription and cloning OpenAFS repo kindly follow **step 4-5-6** in "Build and creation of Images" section of openafs-csi-driver [README](https://github.com/openafs-contrib/openafs-csi-driver/blob/master/README.md))


    ```sh
    docker build -tkinitimage:latest . 
    ```
    
    Once image is created save the image in tar format and load it on all worker nodes. (To save and load images follow **step 8-9** in "Build and creation of Images" section of openafs-csi-driver [README](https://github.com/openafs-contrib/openafs-csi-driver/blob/master/README.md)).


### Using SideCar inside a POD:

There is an example pod.yaml file. Below is a description for various section in pod.yaml

- We have our microservice with name "my-frontend" which is using centos image. Inside this application we have mounted PVC openafs-pvc on /data inside a POD. PVC openafs-pvc is created by openafs-csi-driver using dynamic or static provisioning.

- In order to access openafs-pvc we need tokens for a cell. In our example it is punetest.in.ibm.com. To get tokens we need below information
    ```sh
    username and password
	krb5.conf, CellServDB and ThisCell file for a cell 
    CellName
    ```

- For username and password we need to create a `secret`, in current example we are using secret as afs-refresh-tok and file for this is secret.yaml.

- For krb5.conf, CellServDB and ThisCell we will create a `configMap`. In our example we are using configMap as openafs-config. This configMap can be created using below command
    ```sh
	kubectl create cm openafs-config --from-file=<ThisCell location> --from-file=<CellServDB location> --from-file=<krb5.conf location>
    ```
- In our example we are using `SideCar container ticket-refresh` which is using sidecar image `kinitimage:latest`. Mount secret afs-refresh-tok, and configmap openafs-config inside sidecar container. 

- Pass below environment variables to sidecar container
    ```sh
	CELLNAME: <Name of a cell>
	REFRESH_TIME: <How often tokens should be refreshed>
    ```
Once above configuration is done start a POD. On POD start, sidecar container will get UID based tokens, since UID based tokens can be seen by all processes running with same UID, so our main microservice my-frontend can also get tokens and access AFS space.
		

