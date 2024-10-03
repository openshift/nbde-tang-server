/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	daemonsv1alpha1 "github.com/openshift/nbde-tang-server/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Check if really necessary
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Finalizer for tang server
const DEFAULT_TANG_FINALIZER = "finalizer.daemons.tangserver.redhat.com"

// Default recheck of keys when no active keys exit
const DEFAULT_RECONCILE_TIMER_NO_ACTIVE_KEYS = 5 // seconds

// Var to count active keys retrying counter
var activeKeyRetries uint32 = 0

// TangServerReconciler reconciles a TangServer object
type TangServerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// contains returns true if a string is found on a slice
func contains(hayjack []string, needle string) bool {
	for _, n := range hayjack {
		if n == needle {
			return true
		}
	}
	return false
}

// isInstanceMarkedToBeDeleted checks if deletion has been initialized for tang server
func isInstanceMarkedToBeDeleted(tangserver *daemonsv1alpha1.TangServer) bool {
	return tangserver.GetDeletionTimestamp() != nil
}

// dumpToErrFile allows dumping string to error file
func dumpToErrFile(msg string) {
	f, err := os.OpenFile("/tmp/tangserver-error", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(msg); err != nil {
		panic(err)
	}
}

// getSHA256 returns a random SHA256 number
func getSHA256() string {
	data := make([]byte, 10)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(data))
	return sha
}

// updateUID allows to set a UID for those cases where it is not set (i.e.:running on test infra)
func updateUID(cr *daemonsv1alpha1.TangServer, req ctrl.Request) {
	// Ugly hack to update UID for test to run appropriately
	if req.NamespacedName.Name == daemonsv1alpha1.DefaultTestName {
		cr.ObjectMeta.UID = types.UID(getSHA256())
	}
}

// finalizeTangServerApp runs required tasks before deleting the objects owned by the CR
func (r *TangServerReconciler) finalizeTangServer(cr *daemonsv1alpha1.TangServer) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	GetLogInstance().Info("Successfully finalized TangServer")
	return nil
}

// checkCRReadyForDeletion will check if CR can be deleted appropriately
func (r *TangServerReconciler) checkCRReadyForDeletion(ctx context.Context, tangserver *daemonsv1alpha1.TangServer) (ctrl.Result, error) {
	if contains(tangserver.GetFinalizers(), DEFAULT_TANG_FINALIZER) {
		// Run the finalizer logic
		err := r.finalizeTangServer(tangserver)
		if err != nil {
			// Don't remove the finalizer if we failed to finalize the object
			return ctrl.Result{}, err
		}
		GetLogInstance().Info("TangServer finalizers completed")
		// Remove finalizer once the finalizer logic has run
		controllerutil.RemoveFinalizer(tangserver, DEFAULT_TANG_FINALIZER)
		err = r.Update(ctx, tangserver)
		if err != nil {
			// If the object update fails, requeue
			return ctrl.Result{}, err
		}
	}
	GetLogInstance().Info("TangServer can be deleted now")
	return ctrl.Result{}, nil
}

//+kubebuilder:rbac:groups=daemons.redhat.com,resources=tangservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=daemons.redhat.com,resources=tangservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=daemons.redhat.com,resources=tangservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=core,resources=pods/log,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=core,resources=pods/exec,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;list;watch;create;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TangServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
// +kubebuilder:rbac:groups=apps.redhat,resources=tangservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.redhat,resources=tangservers/status,verbs=get;update;patch
func (r *TangServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	SetLogInstance(l)

	tangserver := &daemonsv1alpha1.TangServer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.NamespacedName.Namespace,
			Name:      req.NamespacedName.Name,
		},
	}
	updateUID(tangserver, req)

	err := r.Get(ctx, req.NamespacedName, tangserver)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("TangServer resource not found")
		}
	}

	// Check if the CR is marked to be deleted
	if isInstanceMarkedToBeDeleted(tangserver) {
		l.Info("Instance marked for deletion, running finalizers")
		return r.checkCRReadyForDeletion(ctx, tangserver)
	}

	// Reconcile Deployment object
	result, err := r.reconcileDeployment(tangserver)
	if err != nil {
		l.Error(err, "Error on deployment reconciliation", "Error:", err.Error())
		dumpToErrFile("Error on deployment reconciliation, Error:" + err.Error() + "\n")
		return result, err
	}
	// Reconcile Service object
	result, err = r.reconcileService(tangserver)
	if err != nil {
		l.Error(err, "Error on service reconciliation")
		return result, err
	}

	// Reconcile finished, requeue for key refresh if necessary
	var reconcile bool
	if result, reconcile = r.reconcilePeriodic(tangserver); reconcile {
		return result, nil
	}
	return ctrl.Result{}, nil
}

// checkDeploymentImage returns wether the deployment image is different or not
func checkDeploymentImage(current *appsv1.Deployment, desired *appsv1.Deployment) bool {
	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}
	return false
}

// mustRedeploy checks for cases where redeploy must be performed
func mustRedeploy(new *appsv1.Deployment, prev *appsv1.Deployment) bool {
	if new.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] !=
		prev.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] ||
		new.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory] ||
		new.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] ||
		new.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] !=
			prev.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] {
		return true
	}
	return false
}

// handleHiddenKeys rotate keys if user specifies so in the spec
func (r *TangServerReconciler) handleHiddenKeys(keyinfo KeyObtainInfo) bool {
	rotated := false
	// check hidden keys to be maintained (and delete those in the status non specified)
	keepKeyMap := make(KeySelectiveMap, 0)
	for _, hk := range keyinfo.TangServer.Spec.HiddenKeys {
		for _, hsk := range keyinfo.TangServer.Status.HiddenKeys {
			if hk.Sha1 == hsk.Sha1 {
				keepKeyMap[hk.Sha1] = hk.Sha1
				keepKeyMap[hsk.Sha256] = hsk.Sha256
			} else if hk.Sha256 == hsk.Sha256 {
				keepKeyMap[hk.Sha256] = hk.Sha256
				keepKeyMap[hsk.Sha1] = hsk.Sha1
			}
		}
	}
	// only delete selectively if have something to keep
	if len(keepKeyMap) > 0 {
		if err := deleteHiddenKeysSelectively(keepKeyMap, keyinfo); err != nil {
			GetLogInstance().Error(err, "Unable to delete keys selectively", "keymap", keepKeyMap, "keyInfo", keyinfo)
		}
	}

	// check if key is in active keys and rotate it
	for _, hk := range keyinfo.TangServer.Spec.HiddenKeys {
		for _, ak := range keyinfo.TangServer.Status.ActiveKeys {
			if ak.Sha1 == hk.Sha1 || ak.Sha256 == hk.Sha256 {
				GetLogInstance().Info("Key must be rotated", "sha1", hk.Sha1,
					"sha256", hk.Sha256)
				kr := KeyRotateInfo{
					KeyInfo:     &keyinfo,
					KeyFileName: ak.FileName,
				}
				if err := rotateKey(kr); err == nil {
					rotated = true
					GetLogInstance().Info("Key rotated correctly", "sha1", hk.Sha1, "sha256", hk.Sha256)
					keyinfo.TangServer.Status.TangServerError = daemonsv1alpha1.NoError
					r.Recorder.Event(keyinfo.TangServer, "Normal", "KeyRotation", fmt.Sprintf("Key Rotated Correctly, Key File: %s", ak.FileName))
					if err := rotateUnadvertisedKeys(kr); err != nil {
						GetLogInstance().Error(err, "Unable to rotate unadvertised keys", "Rotating Key", kr)
					}
				} else {
					GetLogInstance().Error(err, "Key not rotated correctly", "sha1", hk.Sha1, "sha256", hk.Sha256)
					r.Recorder.Event(keyinfo.TangServer, "Error", "KeyRotation", fmt.Sprintf("Key NOT Rotated Correctly, Key File: %s", ak.FileName))
					keyinfo.TangServer.Status.TangServerError = daemonsv1alpha1.ActiveKeyNotFoundError
				}
			}
		}
	}
	return rotated
}

// UpdateKeys updates keys in the CR status
func (r *TangServerReconciler) UpdateKeys(k KeyObtainInfo) {
	newKeysCreated := r.CreateNewKeysIfNecessary(k)
	// Read first hidden keys, as created will be retrieved from active keys (if exists)
	hiddenKeys, _ := readHiddenKeys(k, ONLY_ADVERTISED)
	activeKeys, _ := readActiveKeys(k, ONLY_ADVERTISED)
	k.TangServer.Status.ActiveKeys = activeKeys
	k.TangServer.Status.HiddenKeys = hiddenKeys
	if newKeysCreated {
		GetLogInstance().Info("New active keys created", "Active Keys",
			activeKeys, "Hidden Keys", hiddenKeys)
	} else {
		GetLogInstance().Info("No new active keys created",
			"Active Keys", activeKeys, "Hidden Keys", hiddenKeys)
	}
}

// CreateNewKeysIfNecessary creates new keys if spec mandates so
func (r *TangServerReconciler) CreateNewKeysIfNecessary(k KeyObtainInfo) bool {
	requiredActiveKeyPairs := daemonsv1alpha1.DefaultActiveKeyPairs
	if k.TangServer.Spec.RequiredActiveKeyPairs > 0 {
		requiredActiveKeyPairs = k.TangServer.Spec.RequiredActiveKeyPairs
		GetLogInstance().Info("Using specified required active keys", "Key Amount", requiredActiveKeyPairs)
	} else {
		GetLogInstance().Info("Using default active keys", "Key Amount", requiredActiveKeyPairs)
	}
	GetLogInstance().Info("createNewKeysIfNecessary", "Active Keys", uint32(len(k.TangServer.Status.ActiveKeys)), "Required Active Keys", requiredActiveKeyPairs)
	// Only create if more than one required active key pairs. Otherwise, they are automatically created
	if uint32(len(k.TangServer.Status.ActiveKeys)) < requiredActiveKeyPairs && requiredActiveKeyPairs > 1 {
		if err := createNewPairOfKeys(k); err != nil {
			GetLogInstance().Error(err, "Unable to create new keys", "KeyObtainInfo", k)
			r.Recorder.Event(k.TangServer, "Error", "NewKeys", "Unable to create new pair of keys")
		} else {
			GetLogInstance().Info("New Active Keys Created", "KeyObtainInfo", k, "Active Keys", uint32(len(k.TangServer.Status.ActiveKeys)), "Required Active Keys", requiredActiveKeyPairs)
			r.Recorder.Event(k.TangServer, "Normal", "NewKeys", fmt.Sprintf("Created %d active pair of keys", uint32(len(k.TangServer.Status.ActiveKeys))))
			return true
		}
	}
	return false
}

// reconcileDeployment creates deployment appropriate for this CR
func (r *TangServerReconciler) reconcileDeployment(cr *daemonsv1alpha1.TangServer) (ctrl.Result, error) {
	// Define a new Deployment object
	GetLogInstance().Info("reconcileDeployment")
	deployment := getDeployment(cr)

	// Set tangserver instance as the owner and controller of the Deployment
	if err := ctrl.SetControllerReference(cr, deployment, r.Scheme); err != nil {
		cr.Status.TangServerError = daemonsv1alpha1.CreateError
		return ctrl.Result{}, err
	}

	// Check if this Deployment already exists
	deploymentFound := &appsv1.Deployment{}
	err := r.Get(context.Background(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deploymentFound)
	if err != nil && errors.IsNotFound(err) {
		GetLogInstance().Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(context.Background(), deployment)
		if err != nil {
			cr.Status.TangServerError = daemonsv1alpha1.CreateError
			return ctrl.Result{}, err
		}
		// Requeue the object to update its status
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		cr.Status.TangServerError = daemonsv1alpha1.CreateError
		return ctrl.Result{}, err
	} else {
		// Deployment already exists
		GetLogInstance().Info("Deployment already exists", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
		GetLogInstance().Info("Checking redeployment")
		// Check if it is needed to be redeployed
		if mustRedeploy(deployment, deploymentFound) {
			GetLogInstance().Info("Updating deployment, must redeploy")
			err = r.Update(context.Background(), deployment)
			if err != nil {
				GetLogInstance().Error(err, "Failed to redeploy", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
				r.Recorder.Event(cr, "Error", "Redeploy", "Failed to redeploy")
				return ctrl.Result{}, err
			}
		}
	}

	// Ensure deployment replicas match the desired state
	if !reflect.DeepEqual(deploymentFound.Spec.Replicas, deployment.Spec.Replicas) {
		GetLogInstance().Info("Current deployment do not match Tang Server configured Replicas")
		// Update the replicas
		err = r.Update(context.Background(), deployment)
		if err != nil {
			GetLogInstance().Error(err, "Failed to update Deployment.", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			r.Recorder.Event(cr, "Error", "Update", fmt.Sprintf("Failed to update deployment, name:%s, namespace:%s", deploymentFound.Name, deploymentFound.Namespace))
			return ctrl.Result{}, err
		}
	}

	// Ensure deployment container image match the desired state, returns true if deployment needs to be updated
	if checkDeploymentImage(deploymentFound, deployment) {
		GetLogInstance().Info("Current deployment image version do not match TangServers configured version")
		// Update the image
		err = r.Update(context.Background(), deployment)
		if err != nil {
			GetLogInstance().Error(err, "Failed to update Deployment", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			r.Recorder.Event(cr, "Error", "Update", fmt.Sprintf("Failed to update deployment, name:%s, namespace:%s", deploymentFound.Name, deploymentFound.Namespace))
			return ctrl.Result{}, err
		}
	}

	// Check if the deployment is ready and update replicas as they get ready
	deploymentReady := isDeploymentReady(deploymentFound)
	ready := getDeploymentReadyReplicas(deploymentFound)
	GetLogInstance().Info("Deployment Found Info", "Replicas", deploymentFound.Status.Replicas, "Ready", deploymentFound.Status.ReadyReplicas)
	GetLogInstance().Info("Updating status with ready/running replicas", "Ready", ready, "Running", cr.Spec.Replicas, "DeploymentReady", deploymentReady)
	cr.Status.Running = cr.Spec.Replicas
	cr.Status.Ready = ready
	if !deploymentReady {
		GetLogInstance().Info("Deployment not ready", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
	} else {
		// Create list options to get deployment pods and extract podname
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(deploymentFound.Namespace),
			client.MatchingLabels(deploymentFound.Labels),
		}
		// List the pods for this deployment
		err = r.List(context.Background(), podList, listOpts...)
		if err != nil || len(podList.Items) == 0 {
			GetLogInstance().Error(err, "Failed to list Pods, required for keys", "Deployment.Namespace",
				deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
			r.Recorder.Event(cr, "Error", "PodList", fmt.Sprintf("Failed to list pods in deployment, name:%s, namespace:%s", deploymentFound.Name, deploymentFound.Namespace))
			return ctrl.Result{}, err
		}
		GetLogInstance().Info("Deployment ready", "Deployment.Namespace", deploymentFound.Namespace, "Deployment.Name", deploymentFound.Name)
		k := KeyObtainInfo{
			PodName:    podList.Items[0].Name,
			Namespace:  deploymentFound.Namespace,
			DbPath:     getDefaultKeyPath(cr),
			TangServer: cr,
		}
		if cr.Spec.HiddenKeys == nil {
			GetLogInstance().Info("No hidden keys specified")
		} else if len(cr.Spec.HiddenKeys) == 0 {
			GetLogInstance().Info("Hidden keys specified with len 0, deleting all hidden keys")
			if deleteAllHiddenKeys(k) {
				r.Recorder.Event(cr, "Normal", "HiddenKeysDeletion", "Hidden keys deleted correctly")
			} else {
				r.Recorder.Event(cr, "Error", "HiddenKeysDeletion", "Hidden keys not deleted correctly")
			}
		} else if len(cr.Spec.HiddenKeys) > 0 {
			rotated := r.handleHiddenKeys(k)
			if rotated {
				GetLogInstance().Info("Key(s) rotated", "Keys", cr.Spec.HiddenKeys)
				// if keys are rotated, set the counter of active keys retries to zero
				// just in case no active keys exist
				activeKeyRetries = 0
			} else {
				GetLogInstance().Info("Key(s) not rotated", "Keys", cr.Spec.HiddenKeys)
			}
		}
		r.UpdateKeys(k)
	}
	err = r.Client.Status().Update(context.Background(), cr)
	if err != nil {
		GetLogInstance().Error(err, "Unable to update TangServer status")
		r.Recorder.Event(cr, "Error", "Update", "Unable to update TangServer status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *TangServerReconciler) reconcileService(cr *daemonsv1alpha1.TangServer) (ctrl.Result, error) {
	GetLogInstance().Info("reconcileService")
	service := getService(cr)

	// Set TangServer instance as the owner and controller of the Service
	if err := controllerutil.SetControllerReference(cr, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Service already exists
	serviceFound := &corev1.Service{}
	err := r.Get(context.Background(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, serviceFound)
	if err != nil && errors.IsNotFound(err) {
		GetLogInstance().Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(context.Background(), service)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		GetLogInstance().Error(err, "Error on service Get")
		r.Recorder.Event(cr, "Error", "Service", fmt.Sprintf("Error getting service: name:%s, namespace:%s", service.Name, service.Namespace))
		return ctrl.Result{}, err
	} else {
		// Service already exists
		GetLogInstance().Info("Service already exists", "Service.Namespace", serviceFound.Namespace, "Service.Name", serviceFound.Name)
		if len(serviceFound.Status.LoadBalancer.Ingress) > 0 {
			GetLogInstance().Info("Service Information", "Load Balancer IP", serviceFound.Status.LoadBalancer.Ingress[0].IP, "Load Balancer Hostname", serviceFound.Status.LoadBalancer.Ingress[0].Hostname)
			cr.Status.ServiceExternalURL = getExternalServiceURL(cr, serviceFound.Status.LoadBalancer.Ingress[0])
			err := r.Client.Status().Update(context.Background(), cr)
			if err != nil {
				GetLogInstance().Error(err, "Unable to update TangServer status with Service IP URL")
				r.Recorder.Event(cr, "Error", "Update", "Unable to update TangServer status with Service IP URL")
				return ctrl.Result{}, err
			}
		} else {
			GetLogInstance().Info("Service Information, NO Ingress")
		}
		GetLogInstance().Info("Service Spec", "Spec", serviceFound.Spec)
		GetLogInstance().Info("Service Status", "Status", serviceFound.Status)
	}
	// Service reconcile finished
	return ctrl.Result{}, nil
}

func (r *TangServerReconciler) reconcilePeriodic(cr *daemonsv1alpha1.TangServer) (ctrl.Result, bool) {
	if cr.Spec.KeyRefreshInterval != 0 {
		GetLogInstance().Info("Key reconciliation non zero", "Refresh Interval", cr.Spec.KeyRefreshInterval)
		return ctrl.Result{RequeueAfter: time.Duration(cr.Spec.KeyRefreshInterval) * time.Second}, true
	} else if len(cr.Status.ActiveKeys) == 0 {
		activeKeyRetries = activeKeyRetries + 1
		cr.Status.TangServerError = daemonsv1alpha1.ActiveKeysError
		GetLogInstance().Info("Retrying key retrieval", "Retries:", fmt.Sprint(activeKeyRetries))
		r.Recorder.Event(cr, "Normal", "ActiveKeyRetrieval", fmt.Sprintf("Empty Active Key List Retries: %d", activeKeyRetries))
		err := r.Client.Status().Update(context.Background(), cr)
		if err != nil {
			GetLogInstance().Error(err, "Unable to update TangServer status with active key retries and error")
			r.Recorder.Event(cr, "Error", "Update", "Unable to update TangServer status clearing active key retries and error")
		}
		return ctrl.Result{RequeueAfter: time.Duration(DEFAULT_RECONCILE_TIMER_NO_ACTIVE_KEYS) * time.Second}, true
	} else {
		activeKeyRetries = 0
		cr.Status.TangServerError = daemonsv1alpha1.NoError
		err := r.Client.Status().Update(context.Background(), cr)
		if err != nil {
			GetLogInstance().Error(err, "Unable to update TangServer status clearing active key retries and error")
			r.Recorder.Event(cr, "Error", "Update", "Unable to update TangServer status clearing active key retries and error")
		}
	}
	return ctrl.Result{}, false
}

// SetupWithManager sets up the controller with the Manager.
func (r *TangServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&daemonsv1alpha1.TangServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
