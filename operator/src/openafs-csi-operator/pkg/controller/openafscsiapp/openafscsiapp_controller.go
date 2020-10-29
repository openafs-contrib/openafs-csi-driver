package openafscsiapp

import (
	"context"

	openafscsiv1 "openafs-csi-operator/pkg/apis/openafscsi/v1"

	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_openafscsiapp")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new OpenafsCSIApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileOpenafsCSIApp{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("openafscsiapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OpenafsCSIApp
	err = c.Watch(&source.Kind{Type: &openafscsiv1.OpenafsCSIApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner OpenafsCSIApp

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openafscsiv1.OpenafsCSIApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openafscsiv1.OpenafsCSIApp{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &openafscsiv1.OpenafsCSIApp{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileOpenafsCSIApp implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOpenafsCSIApp{}

// ReconcileOpenafsCSIApp reconciles a OpenafsCSIApp object
type ReconcileOpenafsCSIApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a OpenafsCSIApp object and makes changes based on the state read
// and what is in the OpenafsCSIApp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOpenafsCSIApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OpenafsCSIApp")

	// Fetch the OpenafsCSIApp instance
	instance := &openafscsiv1.OpenafsCSIApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

/* Provisioner statefulset */
	provService := newProvService(instance)

        if err := controllerutil.SetControllerReference(instance, provService, r.scheme); err != nil {
           return reconcile.Result{}, err
        }

	found := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: provService.ObjectMeta.Name, Namespace: provService.ObjectMeta.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Service %v not found, so creating it\n", provService.ObjectMeta.Name)
			err = r.client.Create(context.TODO(), provService)
			if err != nil {
				reqLogger.Info("Unable to create a Service %v\n", provService.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully created service %v\n", provService.ObjectMeta.Name)
			 // Set OpenafsCSIApp instance as the owner and controller
		}
	}


	provStateFulSet := newProvStateFulSet(instance)

        if err := controllerutil.SetControllerReference(instance, provStateFulSet, r.scheme); err != nil {
           return reconcile.Result{}, err
        }

	foundSS := &appsv1.StatefulSet{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: provStateFulSet.ObjectMeta.Name,  Namespace: provStateFulSet.ObjectMeta.Namespace}, foundSS)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Statefulset %v not found, so creating it\n", provStateFulSet.ObjectMeta.Name)
			err = r.client.Create(context.TODO(), provStateFulSet)
			if err != nil {
				reqLogger.Info("Unable to create a Statefulset %v\n", provStateFulSet.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully created Statefulset %v\n", provStateFulSet.ObjectMeta.Name)
                         // Set OpenafsCSIApp instance as the owner and controller
		}
	}

/*=====================Attacher ======================*/


	attachService := newAttacherService(instance)

        if err := controllerutil.SetControllerReference(instance, attachService, r.scheme); err != nil {
           return reconcile.Result{}, err
        }

	attachServicefound := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: attachService.ObjectMeta.Name, Namespace: attachService.ObjectMeta.Namespace}, attachServicefound)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Service %v not found, so creating it\n", attachService.ObjectMeta.Name)
			err = r.client.Create(context.TODO(), attachService)
			if err != nil {
				reqLogger.Info("Unable to create a Service %v\n", attachService.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully created service %v\n", attachService.ObjectMeta.Name)
			 // Set OpenafsCSIApp instance as the owner and controller
		}
	}


	attachStateFulSet := newAttacherStateFulSet(instance)

        if err := controllerutil.SetControllerReference(instance, attachStateFulSet, r.scheme); err != nil {
           return reconcile.Result{}, err
        }

	AtfoundSS := &appsv1.StatefulSet{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: attachStateFulSet.ObjectMeta.Name,  Namespace: attachStateFulSet.ObjectMeta.Namespace}, AtfoundSS)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Statefulset %v not found, so creating it\n", provStateFulSet.ObjectMeta.Name)
			err = r.client.Create(context.TODO(), attachStateFulSet)
			if err != nil {
				reqLogger.Info("Unable to create a Statefulset %v\n", attachStateFulSet.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully created Statefulset %v\n", attachStateFulSet.ObjectMeta.Name)
                         // Set OpenafsCSIApp instance as the owner and controller
		}
	}

/*======================= Plugin ===========================================*/

	pluginDaemonSet := newPluginDaemonset(instance)
        if err := controllerutil.SetControllerReference(instance, pluginDaemonSet, r.scheme); err != nil {
           return reconcile.Result{}, err
        }

	pluginFound := &appsv1.DaemonSet{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pluginDaemonSet.ObjectMeta.Name,  Namespace: pluginDaemonSet.ObjectMeta.Namespace}, pluginFound)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Daemonset %v not found, so creating it\n", pluginDaemonSet.ObjectMeta.Name)
			err = r.client.Create(context.TODO(), pluginDaemonSet)
			if err != nil {
				reqLogger.Info("Unable to create a DaemonSet %v\n", pluginDaemonSet.ObjectMeta.Name)
				return reconcile.Result{}, err
			}
			reqLogger.Info("Successfully created DaemonSet %v\n", pluginDaemonSet.ObjectMeta.Name)
                         // Set OpenafsCSIApp instance as the owner and controller
		}
	}
	reqLogger.Info("Skip reconcile: Service %v and Statefulset %v already exists\n", provService.ObjectMeta.Name, provStateFulSet.ObjectMeta.Name)
	return reconcile.Result{}, nil
}

func newAttacherService(cr *openafscsiv1.OpenafsCSIApp) *corev1.Service {

	return &corev1.Service {
		ObjectMeta: metav1.ObjectMeta {
			Name: cr.Spec.AttacherSpec.AttacherName,
			Namespace: cr.Spec.AttacherSpec.AttacherNameSpace,
			Labels: map[string]string {
						"app": cr.Spec.AttacherSpec.AttacherName,
				},
		},
	        Spec: corev1.ServiceSpec {
			Selector: map[string]string {
				"app": cr.Spec.AttacherSpec.AttacherName,
			},
			Ports: []corev1.ServicePort {
				{
					Name: "dummy",
					Port: 12345,
				},
			},
		},
	}
}




func newProvService(cr *openafscsiv1.OpenafsCSIApp) *corev1.Service {

	return &corev1.Service {
		ObjectMeta: metav1.ObjectMeta {
			Name: cr.Spec.ProvisionerSpec.ProvisionerName,
			Namespace: cr.Spec.ProvisionerSpec.ProvisionerNameSpace,
			Labels: map[string]string {
						"app": cr.Spec.ProvisionerSpec.ProvisionerName,
				},
		},
	        Spec: corev1.ServiceSpec {
			Selector: map[string]string {
				"app": cr.Spec.ProvisionerSpec.ProvisionerName,
			},
			Ports: []corev1.ServicePort {
				{
					Name: "dummy",
					Port: 12345,
				},
			},
		},
	}
}

func newPluginDaemonset(cr *openafscsiv1.OpenafsCSIApp) *appsv1.DaemonSet {
	
	priv := true
	propagation := corev1.MountPropagationBidirectional
	dirOrCreate := corev1.HostPathDirectoryOrCreate
	dir := corev1.HostPathDirectory
	return &appsv1.DaemonSet {
		ObjectMeta:  metav1.ObjectMeta {
			Name: cr.Spec.PluginSpec.PluginName,
			Namespace: cr.Spec.PluginSpec.PluginNameSpace,
		},
		Spec: appsv1.DaemonSetSpec {
			Selector: &metav1.LabelSelector {
				MatchLabels: map[string]string {
					"app": cr.Spec.PluginSpec.PluginName,
				},
			},
			Template: corev1.PodTemplateSpec {
				ObjectMeta: metav1.ObjectMeta {
					Labels: map[string]string {
						"app": cr.Spec.PluginSpec.PluginName,
					},
				},
				Spec: corev1.PodSpec {
					ServiceAccountName: cr.Spec.PluginSpec.PluginName,
					HostNetwork: true,
					Containers: []corev1.Container {
						{
							Name: "node-driver-registrar",
							Image: cr.Spec.PluginSpec.DriverRegistrarImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Lifecycle: &corev1.Lifecycle {
								PreStop: &corev1.Handler {
									Exec: &corev1.ExecAction {
										Command: []string {
											"/bin/sh",
											"-c",
										       "rm -rf /registration/openafs-csi/registration/openafs-csi-reg.sock",
										},
									},									
								},
							},
							Args: []string {
								"--v=5",
								"--csi-address=/csi/csi.sock",
								"--kubelet-registration-path=/var/lib/kubelet/plugins/openafs-csi/csi.sock",
							},
							SecurityContext : &corev1.SecurityContext {
								Privileged: &priv,
							},
							Env: []corev1.EnvVar {
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource {
										FieldRef: &corev1.ObjectFieldSelector {
											APIVersion: "v1",
											FieldPath: "spec.nodeName",
										},
									}, 
								},
							},
							VolumeMounts: []corev1.VolumeMount {
								{
									Name: "socket-dir",
									MountPath: "/csi",									
								},
								{
									Name: "registration-dir",
									MountPath: "/registration",
								},
							},
						},
						{
						 	Name: "openafs",
							Image: cr.Spec.PluginSpec.PluginImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string { 
								"--drivername=afscsi.openafs.org",
								"--v=5",
								"--endpoint=$(CSI_ENDPOINT)",
								"--nodeid=$(KUBE_NODE_NAME)",
							},
							Env: []corev1.EnvVar {
								{
									Name: "CSI_ENDPOINT",
									Value: "unix:///csi/csi.sock",
								},
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource {
										FieldRef: &corev1.ObjectFieldSelector {
											APIVersion: "v1",
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext {
								Privileged: &priv,
							},
							Ports: []corev1.ContainerPort {
								{
									Name: "healthz",
									ContainerPort: 9898,
									Protocol: corev1.ProtocolTCP,
								},
							},
							LivenessProbe : &corev1.Probe {
								FailureThreshold: 5,
								Handler: corev1.Handler {
									HTTPGet: &corev1.HTTPGetAction {
										Path: "/healthz",
										Port: intstr.IntOrString {
											Type:  intstr.String,
											StrVal: "healthz",
										},
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds: 3,
								PeriodSeconds: 2,
							},
							VolumeMounts: []corev1.VolumeMount {
								{
									Name: "socket-dir",
									MountPath: "/csi",
								},
								{
									Name: "mountpoint-dir",
									MountPropagation: &propagation,
									MountPath: "/var/lib/kubelet/pods",
								},
								{
									Name: "plugins-dir",
									MountPropagation: &propagation,
									MountPath: "/var/lib/kubelet/plugins",
								},
								{
									Name: "dev-dir",
									MountPath: "/dev",
								},
								{
									Name: "etc-dir",
									MountPath: "/etc",
								},
								{
									Name: "afs-dir",
									MountPath: "/afs",
								},
							        {
									Name: "afsconf-dir",
									MountPath: "/etc/configmap",
								},
							},												
						},
						{
							Name: "liveness-probe",
							Image: cr.Spec.PluginSpec.LivenessProbeImage,
				   			ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string { 
								"--csi-address=/csi/csi.sock",
								"--health-port=9898",
							},
                                                        VolumeMounts: []corev1.VolumeMount {
                                                                {
                                                                        Name: "socket-dir",
                                                                        MountPath: "/csi",
                                                                },
							},
						},
					},
					Volumes: []corev1.Volume {
						{
							Name: "socket-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/var/lib/kubelet/plugins/openafs-csi",
									Type: &dirOrCreate,
								},
							},							
						},
						{
							Name: "mountpoint-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/var/lib/kubelet/pods",
									Type: &dirOrCreate,
								},
							},							
						},
						{
							Name: "registration-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/var/lib/kubelet/plugins_registry",
									Type: &dir,
								},
							},							
						},
						{
							Name: "plugins-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/var/lib/kubelet/plugins",
									Type: &dir,
								},
							},							
						},
						{
							Name: "dev-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/dev",
									Type: &dir,
								},
							},							
						},
						{
							Name: "etc-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: "/etc",
									Type: &dir,
								},
							},							
						},
						{
							Name: "afs-dir",
							VolumeSource: corev1.VolumeSource {
								HostPath: &corev1.HostPathVolumeSource {
									Path: cr.Spec.PluginSpec.AfsMount,
									Type: &dir,
								},
							},							
						},
						{
							Name: "afsconf-dir",
							VolumeSource: corev1.VolumeSource {
								ConfigMap: &corev1.ConfigMapVolumeSource {
									LocalObjectReference: corev1.LocalObjectReference {
										Name: cr.Spec.PluginSpec.Configmap,										
									},
								},
							},							
						},
					},
				},

			},
		},
	}
}

func newAttacherStateFulSet(cr *openafscsiv1.OpenafsCSIApp) *appsv1.StatefulSet {

	var priv bool
	appLabel := map[string]string {
			"app": cr.Spec.AttacherSpec.AttacherName,
		    }
	priv = true
	hostType := corev1.HostPathDirectoryOrCreate
	var replica int32
	replica = 1

	return &appsv1.StatefulSet {
		ObjectMeta: metav1.ObjectMeta {
			Name: cr.Spec.AttacherSpec.AttacherName,
			Namespace: cr.Spec.AttacherSpec.AttacherNameSpace,
		},
		Spec: appsv1.StatefulSetSpec {
			ServiceName: cr.Spec.AttacherSpec.AttacherName,
			Replicas: &replica,
			Selector: &metav1.LabelSelector {
				MatchLabels: appLabel,
			},
			Template: corev1.PodTemplateSpec {
				ObjectMeta: metav1.ObjectMeta {
					Labels: appLabel,
				},
				Spec: corev1.PodSpec {
					Affinity: &corev1.Affinity {
						PodAffinity: &corev1.PodAffinity {
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm {
								{
									LabelSelector: &metav1.LabelSelector {
										MatchExpressions: []metav1.LabelSelectorRequirement {
											{
												Key: "app",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string {
													cr.Spec.PluginSpec.PluginName,
												},
											},
										},
									},
									TopologyKey:  "kubernetes.io/hostname",
								},
							},
						},
					},
                                        ServiceAccountName: cr.Spec.AttacherSpec.AttacherName,
					Containers: []corev1.Container {
						{
							Name: cr.Spec.AttacherSpec.AttacherName,
							Image: cr.Spec.AttacherSpec.AttacherImageName,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string {
									"-v=5",
									"--csi-address=/csi/csi.sock",
							},
							SecurityContext: &corev1.SecurityContext {
								Privileged: &priv,				
							},
							VolumeMounts: []corev1.VolumeMount {
								{
									Name: "socket-dir",
									MountPath: "/csi",
								},
							},
						},
					},
					Volumes: []corev1.Volume {
						 {
								Name: "socket-dir",
								VolumeSource: corev1.VolumeSource {
									HostPath: &corev1.HostPathVolumeSource {
										Path: "/var/lib/kubelet/plugins/openafs-csi",
										Type: &hostType,
									},
								},
						},
					},
				},
			},
		},
	}
}


func newProvStateFulSet(cr *openafscsiv1.OpenafsCSIApp) *appsv1.StatefulSet {

	var priv bool
	appLabel := map[string]string {
			"app": cr.Spec.ProvisionerSpec.ProvisionerName,
		    }
	priv = true
	hostType := corev1.HostPathDirectoryOrCreate
	var replica int32
	replica = 1

	return &appsv1.StatefulSet {
		ObjectMeta: metav1.ObjectMeta {
			Name: cr.Spec.ProvisionerSpec.ProvisionerName,
			Namespace: cr.Spec.ProvisionerSpec.ProvisionerNameSpace,
		},
		Spec: appsv1.StatefulSetSpec {
			ServiceName: cr.Spec.ProvisionerSpec.ProvisionerName,
			Replicas: &replica,
			Selector: &metav1.LabelSelector {
				MatchLabels: appLabel,
			},
			Template: corev1.PodTemplateSpec {
				ObjectMeta: metav1.ObjectMeta {
					Labels: appLabel,
				},
				Spec: corev1.PodSpec {
					Affinity: &corev1.Affinity {
						PodAffinity: &corev1.PodAffinity {
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm {
								{
									LabelSelector: &metav1.LabelSelector {
										MatchExpressions: []metav1.LabelSelectorRequirement {
											{
												Key: "app",
												Operator: metav1.LabelSelectorOpIn,
												Values: []string {
													cr.Spec.PluginSpec.PluginName,
												},
											},
										},
									},
									TopologyKey:  "kubernetes.io/hostname",
								},
							},
						},
					},
                                        ServiceAccountName: cr.Spec.ProvisionerSpec.ProvisionerName,
					Containers: []corev1.Container {
						{
							Name: "csi-provisioner",
							Image: cr.Spec.ProvisionerSpec.ProvisionerImageName,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string {
									"-v=5",
									"--csi-address=/csi/csi.sock",
									"--feature-gates=Topology=true",
  									"--timeout=360s",
							},
							SecurityContext: &corev1.SecurityContext {
								Privileged: &priv,				
							},
							VolumeMounts: []corev1.VolumeMount {
								{
									Name: "socket-dir",
									MountPath: "/csi",
								},
							},
						},
					},
					Volumes: []corev1.Volume {
						 {
								Name: "socket-dir",
								VolumeSource: corev1.VolumeSource {
									HostPath: &corev1.HostPathVolumeSource {
										Path: "/var/lib/kubelet/plugins/openafs-csi",
										Type: &hostType,
									},
								},
						},
					},
				},
			},
		},
	}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *openafscsiv1.OpenafsCSIApp) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
