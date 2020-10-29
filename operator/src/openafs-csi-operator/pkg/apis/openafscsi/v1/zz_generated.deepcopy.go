// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenAFSAttacherSpec) DeepCopyInto(out *OpenAFSAttacherSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenAFSAttacherSpec.
func (in *OpenAFSAttacherSpec) DeepCopy() *OpenAFSAttacherSpec {
	if in == nil {
		return nil
	}
	out := new(OpenAFSAttacherSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenAFSPluginSpec) DeepCopyInto(out *OpenAFSPluginSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenAFSPluginSpec.
func (in *OpenAFSPluginSpec) DeepCopy() *OpenAFSPluginSpec {
	if in == nil {
		return nil
	}
	out := new(OpenAFSPluginSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenAFSProvSpec) DeepCopyInto(out *OpenAFSProvSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenAFSProvSpec.
func (in *OpenAFSProvSpec) DeepCopy() *OpenAFSProvSpec {
	if in == nil {
		return nil
	}
	out := new(OpenAFSProvSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenafsCSIApp) DeepCopyInto(out *OpenafsCSIApp) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenafsCSIApp.
func (in *OpenafsCSIApp) DeepCopy() *OpenafsCSIApp {
	if in == nil {
		return nil
	}
	out := new(OpenafsCSIApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenafsCSIApp) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenafsCSIAppList) DeepCopyInto(out *OpenafsCSIAppList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpenafsCSIApp, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenafsCSIAppList.
func (in *OpenafsCSIAppList) DeepCopy() *OpenafsCSIAppList {
	if in == nil {
		return nil
	}
	out := new(OpenafsCSIAppList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenafsCSIAppList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenafsCSIAppSpec) DeepCopyInto(out *OpenafsCSIAppSpec) {
	*out = *in
	out.ProvisionerSpec = in.ProvisionerSpec
	out.AttacherSpec = in.AttacherSpec
	out.PluginSpec = in.PluginSpec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenafsCSIAppSpec.
func (in *OpenafsCSIAppSpec) DeepCopy() *OpenafsCSIAppSpec {
	if in == nil {
		return nil
	}
	out := new(OpenafsCSIAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenafsCSIAppStatus) DeepCopyInto(out *OpenafsCSIAppStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenafsCSIAppStatus.
func (in *OpenafsCSIAppStatus) DeepCopy() *OpenafsCSIAppStatus {
	if in == nil {
		return nil
	}
	out := new(OpenafsCSIAppStatus)
	in.DeepCopyInto(out)
	return out
}
