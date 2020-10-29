/**
 * Copyright 2020 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


package openafs
import (
	"strconv"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	utilexec "k8s.io/utils/exec"
	"io/ioutil"
	"github.com/golang/glog"
        "google.golang.org/grpc/codes"
        "google.golang.org/grpc/status"
)
type openAFS struct {
        name              string
        nodeID            string
        version           string
        endpoint          string
        ids *identityServer
        ns  *nodeServer
        cs  *controllerServer
}
type openAFSVolume struct {
        VolName       string     `json:"volName"`
        VolID         string     `json:"volID"`
        VolSize       int64      `json:"volSize"`
        VolPath       string     `json:"volPath"`
        ServerIp      string     `json: "serverIP"`
        Partition     string     `json: "partition"`
        CellName      string     `json: "cellName"`
	MntCellName   string     `json: "mntCellName"`
	Secret        map[string]string `json: "secret"`
	KlogCmd	      string      `json: "klogCmd"`
	Acl           string      `json: "acl"`
}
type cellDetails struct {
	UserName	string `json:"userName"`
	Password 	string `json:"password"`
	Complete	bool   `json:"complete"`
}
var (
	openAFSVolumes   map[string]openAFSVolume
	openafsPAGShPath = "pagsh"
	vosPath		 = "vos"
	fsPath	         = "fs"
        klogPath         = "klog.krb5"
	pagshCmd         = fmt.Sprintf("%s %s", openafsPAGShPath, "-c")
	CmThisCell	 = "/etc/configmap/ThisCell"
	CmCellServDB	 = "/etc/configmap/CellServDB"
	CmKrb5Conf 	 = "/etc/configmap/krb5.conf"
	ConfigPath	 = "/usr/local/etc/openafs/"
)
func copyFile(sourceFile string, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
        if err != nil {
		glog.Infof("Unable to open %v for reading\n", sourceFile)
                return err
        }
        err = ioutil.WriteFile(destinationFile, input, 0644)
        if err != nil {
		glog.Infof("Unable to open %v for Writing\n", destinationFile)
                return err
        }	
	return nil
}
func copyConfigFiles() error {
	thisCellLoc := fmt.Sprintf("%s/%s", ConfigPath, "ThisCell")
	err := copyFile(CmThisCell, thisCellLoc)
	if err != nil {
		return err
	}
	cellServDBLoc := fmt.Sprintf("%s/%s", ConfigPath, "CellServDB")
	err = copyFile(CmCellServDB, cellServDBLoc)
	if err != nil {
		return err
	}
	err = copyFile(CmKrb5Conf, "/etc/krb5.conf")

	if err != nil {
		return err
	}
	return nil
}
func getOpenAFSVolumeOptions(volOptions map[string]string) (*openAFSVolume, error) {
	glog.Infof("Value of parameters [%v]", volOptions)
	openAFSVol := &openAFSVolume{}
        if _, ok := volOptions["cellname"]; !ok {
		return nil, status.Error(codes.InvalidArgument, "cellname missing in request")
	} else {
		openAFSVol.CellName = volOptions["cellname"]
	}
	if _, ok := volOptions["partition"]; !ok {
		return nil, status.Error(codes.InvalidArgument, "partition missing in request")
	} else {
		openAFSVol.Partition = volOptions["partition"]
	}
	if _, ok := volOptions["volumepath"]; !ok {
		return nil, status.Error(codes.InvalidArgument, "volumepath missing in request")
	} else {
		openAFSVol.VolPath = volOptions["volumepath"]
	} 
	if _, ok := volOptions["server"]; !ok {
		return nil, status.Error(codes.InvalidArgument, "server information missing in request")
	} else {
		openAFSVol.ServerIp = volOptions["server"]
	}
	openAFSVol.Acl = volOptions["acl"]
	glog.Infof("OpenAFS Volume Options [%v]", openAFSVol)
	return openAFSVol, nil
}
func NewOpenAFSDriver(driverName, nodeID, endpoint string, version string) (*openAFS, error) {
	if driverName == "" {
		return nil, errors.New("no driver name provided")
	}
	if nodeID == "" {
		return nil, errors.New("no node id provided")
	}
	if endpoint == "" {
		return nil, errors.New("no driver endpoint provided")
	}
	err := copyConfigFiles()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to update config %v", err))
	}
	glog.Infof("Driver: %v ", driverName)
	glog.Infof("Version: %s", version)
	copyConfigFiles()
	return &openAFS{
		name:              driverName,
		version:           version,
		nodeID:            nodeID,
		endpoint:          endpoint,
	}, nil
}

func (op *openAFS) Run() {
	op.ids = NewIdentityServer(op.name, op.version)
	op.ns = NewNodeServer(op.nodeID)
	op.cs = NewControllerServer(op.nodeID)
	s := NewNonBlockingGRPCServer()
	s.Start(op.endpoint, op.ids, op.cs, op.ns)
	s.Wait()
}
func getCellMap(secretMap map[string]string) (map[string]*cellDetails,error) {
	cellMap := make(map[string]*cellDetails)
	var cellDet *cellDetails
	var cellName string
        for key, value := range secretMap {
	   splitkey := strings.Split(key, "_")
	   cellName = strings.ToUpper(splitkey[0])
	   if _, ok := cellMap[cellName]; !ok {
              cellDet = new(cellDetails)
           } else {
              cellDet = cellMap[cellName]
           }
           if (strings.Contains(strings.ToLower(key), strings.ToLower(cellName+"_user"))) {
                   value = strings.TrimSuffix(value, "\n")
	           cellDet.UserName = value
           }
           if (strings.Contains(strings.ToLower(key), strings.ToLower(cellName+"_password"))) {
                   value = strings.TrimSuffix(value, "\n")
	           cellDet.Password = value
           }
	   if ((cellDet.UserName != "") && (cellDet.Password != "")) {
		   cellDet.Complete = true
	   }
	   cellMap[cellName] = cellDet 
	}
	return cellMap, nil
}

func executeCmd( cmdStr string) (string, error) {
	d1 := []byte(cmdStr)
	err := ioutil.WriteFile("/tmp/test.sh", d1, 0744)
	executor := utilexec.New()
	out, err := executor.Command("/bin/bash", "/tmp/test.sh").CombinedOutput()
	glog.Infof("Output for a command is : [%v]", string(out))
	return string(out), err
}

func deleteOpenAFSVolume(volId string, klog_cmd string) error {
	split := strings.Split(volId, "#")
	volName := split[0]
	MntPath := split[1]
	basePath := filepath.Base(MntPath)
	split2 := strings.Split(basePath, "-")
	provCell := split2[0]
	Partition := split2[1]
	serverIp := split2[2]
        fsRmmCmd := fmt.Sprintf("%s \"%s;%s rmm %s\"", pagshCmd, klog_cmd, fsPath, MntPath)
        outputBytes, err := executeCmd(fsRmmCmd)
        if err != nil {
                if (strings.Contains(outputBytes, "doesn't exist")) {
                        glog.Infof("Looks fs rmm was already done. Go ahead: [%v]", outputBytes)
                } else {
                        return err
                }
        }
        vosRemoveCmd := fmt.Sprintf("%s \"%s; %s remove %s %s %s -c %s\"", pagshCmd, klog_cmd, vosPath, serverIp, Partition, volName, provCell)
        outputBytes, err = executeCmd(vosRemoveCmd)
        if (err != nil) {
                if (strings.Contains(outputBytes, "no such entry")) {
                        glog.Infof("Volume %v is already deleted", volName)
                } else {
                        return err
                }
        }
	return nil
}

func createOpenAFSVolume(openAFSVol *openAFSVolume) error {
	klog_cmd := openAFSVol.KlogCmd 			
	vosCreateCmd := fmt.Sprintf("%s \"%s;%s %s %s %s %s %s %s %s %s\"", pagshCmd, klog_cmd, vosPath, "create", openAFSVol.ServerIp, openAFSVol.Partition, openAFSVol.VolID, "-max",  strconv.FormatInt(openAFSVol.VolSize, 10), "-c", openAFSVol.CellName) 
	out, err := executeCmd(vosCreateCmd)
        if err != nil {
	   if (strings.Contains(out, "already exists")) {
		glog.Infof("Looks volume %v was created, lets continue", openAFSVol.VolID)
	   } else {
		  vCreateMsg := fmt.Sprintf("%s %s %s %s %s %s %s %s %s\n", vosPath, "create", openAFSVol.ServerIp, openAFSVol.Partition, openAFSVol.VolID, "-max",  strconv.FormatInt(openAFSVol.VolSize, 10), "-c", openAFSVol.CellName)
		  failreqmap[openAFSVol.VolName] = openAFSVol.VolID
		  return status.Errorf(codes.Internal, "failed to create volume %v:%v: %v", vCreateMsg, out, err)
	   }
        }
	 mntName := fmt.Sprintf("%s-%s-%s-%s", openAFSVol.CellName, openAFSVol.Partition, openAFSVol.ServerIp, openAFSVol.VolID)
	fsMkmCmd := fmt.Sprintf("%s \"%s;%s %s %s/%s %s %s %s\"", pagshCmd, klog_cmd, fsPath, "mkm", openAFSVol.VolPath, mntName, openAFSVol.VolID, "-c", openAFSVol.CellName)
        out, err = executeCmd(fsMkmCmd)
        if err != nil {
           if (strings.Contains(out, "File exists")) {
                glog.Infof("Looks Mount %v was created, lets continue", openAFSVol.VolID)
           } else {
		   fsMkmErrMsg := fmt.Sprintf("%s %s %s/%s %s %s %s\n", fsPath, "mkm", openAFSVol.VolPath, mntName, openAFSVol.VolID, "-c", openAFSVol.CellName)
		   failreqmap[openAFSVol.VolName] = openAFSVol.VolID
		   return status.Errorf(codes.Internal, "failed to mount volume %v:%v: %v", fsMkmErrMsg, out, err)
           }
        }
        mntPath := filepath.Join(openAFSVol.VolPath, mntName)
	if (openAFSVol.Acl != "") {
		setAclCmd := fmt.Sprintf("%s \"%s; %s setacl -dir %s -acl %s\"", pagshCmd, klog_cmd, fsPath, mntPath, openAFSVol.Acl)
	        out, err = executeCmd(setAclCmd)
	        if err != nil {
		   setAclErrMsg := fmt.Sprintf("%s setacl -dir %s -acl %s", fsPath, mntPath, openAFSVol.Acl)
    		   failreqmap[openAFSVol.VolName] = openAFSVol.VolID
                   return status.Errorf(codes.Internal, "failed to setacl volume %v:%v: %v", setAclErrMsg, out, err)
	        }
	}
	openAFSVol.VolID = fmt.Sprintf("%s#%s", openAFSVol.VolID, mntPath)
        return nil
}
