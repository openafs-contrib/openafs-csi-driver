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
	"fmt"
	"strconv"
        "math/rand"
	"time"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/container-storage-interface/spec/lib/go/csi"
)
var reqmap = make(map[string]string)
var failreqmap = make(map[string]string)
type controllerServer struct {
	caps   []*csi.ControllerServiceCapability
	nodeID string
	
}
func NewControllerServer( nodeID string) *controllerServer {
	return &controllerServer{
		caps: getControllerServiceCapabilities(
			[]csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			}),
		nodeID: nodeID,
	}
}
func (cs *controllerServer) CreateKlogCommand(cellM map[string]*cellDetails) (string, error) {
	first := true	
	outStr := ""
	for cellName, mapval := range cellM {
		if (first) {
			outStr = fmt.Sprintf("echo %s | %s %s@%s -k %s -c %s", mapval.Password, klogPath, mapval.UserName, cellName, cellName, cellName)
		} else {
			outStr = outStr + fmt.Sprintf(" && echo %s |%s %s@%s -k %s -c %s", mapval.Password, klogPath, mapval.UserName, cellName, cellName, cellName)
		}
		first = false
	}
	return outStr, nil
}
func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	_, volpresent := reqmap[req.GetName()]
	if (volpresent) {
		return nil, status.Error(codes.Aborted, fmt.Sprintf("Volume creation already in process : %v", req.GetName()))
	}
	reqmap[req.GetName()] = req.GetName()
	defer delete(reqmap, req.GetName())
	secretsVal := req.GetSecrets()
	if secretsVal == nil {
		return nil, status.Error(codes.InvalidArgument, "secret missing in volume request")
	}
	capacity := int64(req.GetCapacityRange().GetRequiredBytes())

	openAFSVol, err := getOpenAFSVolumeOptions(req.GetParameters())
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())
	volumeID := strconv.Itoa(rand.Intn(2097152))
	openAFSVol.Secret = secretsVal
        openAFSVol.VolID = fmt.Sprintf("pvc%s",volumeID);

	if _, ok := failreqmap[req.GetName()]; ok {
		openAFSVol.VolID = failreqmap[req.GetName()]
		delete(failreqmap, req.GetName())
	}
	openAFSVol.VolName = req.GetName()
	openAFSVol.VolSize = capacity	
	cellM, _ := getCellMap(openAFSVol.Secret)
        klog_cmd, _ := cs.CreateKlogCommand(cellM)
	openAFSVol.KlogCmd = klog_cmd
	err = createOpenAFSVolume(openAFSVol)
	if err != nil {
		return nil, err
	}
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:           openAFSVol.VolID,
			CapacityBytes:      req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext:      req.GetParameters(),
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid delete volume req: %v", req)
		return nil, err
	}
        volId := req.GetVolumeId()

	if req.GetSecrets() == nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Volume delete request has not secret"))
	}
        cellM, _ := getCellMap(req.GetSecrets())
	klog_cmd, _ := cs.CreateKlogCommand(cellM)
        _, deletingVol:= reqmap[volId]
        if (deletingVol) {
                return nil, status.Error(codes.Aborted, fmt.Sprintf("Volume deletion already in process : %v", volId))
        }
        reqmap[volId] = volId
        defer delete(reqmap, volId)
	err := deleteOpenAFSVolume(volId, klog_cmd)
	if err != nil {
		return nil, err
	}
	glog.V(4).Infof("volume %v successfully deleted", volId)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.caps,
	}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}
	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, req.VolumeId)
	}


	for _, cap := range req.GetVolumeCapabilities() {
		if cap.GetMount() == nil && cap.GetBlock() == nil {
			return nil, status.Error(codes.InvalidArgument, "cannot have both mount and block access type be undefined")
		}

		// A real driver would check the capabilities of the given volume with
		// the set of requested capabilities.
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}, nil
}

func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (cs *controllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (cs *controllerServer) validateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}
	for _, cap := range cs.caps {
		if c == cap.GetRpc().GetType() {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported capability %s", c)
}
func getControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) []*csi.ControllerServiceCapability {
	var csc []*csi.ControllerServiceCapability

	for _, cap := range cl {
		glog.Infof("Enabling controller service capability: %v", cap.String())
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}
	return csc
}
