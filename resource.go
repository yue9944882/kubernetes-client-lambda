package lambda

import (
	"strings"
	"time"

	api_app_v1 "k8s.io/api/apps/v1beta1"
	api_autoscale_v1 "k8s.io/api/autoscaling/v1"
	api_batch_v1 "k8s.io/api/batch/v1"
	api_batch_v2alpha1 "k8s.io/api/batch/v2alpha1"
	api_v1 "k8s.io/api/core/v1"
	api_ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_network_v1 "k8s.io/api/networking/v1"
	api_rbac_v1 "k8s.io/api/rbac/v1"
	api_store_v1 "k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// core
	Pod                   Resource = "pods"
	Namespace             Resource = "namespaces"
	Node                  Resource = "nodes"
	Event                 Resource = "events"
	Service               Resource = "services"
	Endpoints             Resource = "endpoints"
	LimitRange            Resource = "limitranges"
	Secret                Resource = "secrets"
	ConfigMap             Resource = "configmaps"
	ServiceAccout         Resource = "serviceaccounts"
	PodTemplate           Resource = "podtemplates"
	ResourceQuota         Resource = "resourcequotas"
	PersistentVolume      Resource = "persistentvolumes"
	PersistentVolumeClaim Resource = "persistentvolumeclaims"
	ReplicationController Resource = "replicationcontrollers"

	// extensions
	Ingress           Resource = "ingresses"
	ReplicaSet        Resource = "replicasets"
	Deployment        Resource = "deployments"
	DaemonSet         Resource = "daemonsets"
	PodSecurityPolicy Resource = "podsecuritypolicies"

	// apps
	StatefulSet        Resource = "statefulsets"
	ControllerRevision Resource = "controllerrevisions"

	// rbac
	ClusterRole        Resource = "clusterroles"
	ClusterRoleBinding Resource = "clusterrolebindings"
	Role               Resource = "roles"
	RoleBinding        Resource = "rolebindings"

	// batch
	Job     Resource = "jobs"
	CronJob Resource = "cronjobs"

	// storage
	StorageClass Resource = "storageclasses"
	// VolumeAttachment Resource = "volumeattachments"

	// settings
	// PodPreset Resource = "podpresets"

	// network
	NetworkPolicy Resource = "networkpolicies"

	// autoscaling
	HorizontalPodAutoscaler Resource = "horizontalpodautoscalers"

	// authentication

	// admissionregistration
	// InitializerConfiguration       Resource = "initializerconfigurations"
	// MutatingWebhookConfiguration   Resource = "mutatingwebhookconfigurations"
	// ValidatingWebhookConfiguration Resource = "validatingwebhookconfigurations"

	// certificates

	// policy
	// PodDisruptionBudget Resource = "poddisruptionbudgets"

	// scheduling
	// PriorityClass Resource = "priorityclasses"
)

var (
	defaultListTimeout = time.Minute
)

func GetResources() []Resource {
	return []Resource{
		// core
		Pod,
		Namespace,
		Node,
		Event,
		Service,
		Endpoints,
		LimitRange,
		Secret,
		ConfigMap,
		ServiceAccout,
		PodTemplate,
		ResourceQuota,
		PersistentVolume,
		PersistentVolumeClaim,
		ReplicationController,

		// extensions
		Ingress,
		ReplicaSet,
		Deployment,
		DaemonSet,
		PodSecurityPolicy,

		// apps
		StatefulSet,
		ControllerRevision,

		// rbac
		ClusterRole,
		ClusterRoleBinding,
		Role,
		RoleBinding,

		// batch
		Job,
		CronJob,

		// storage
		StorageClass,

		// network
		NetworkPolicy,

		// autoscaling
		HorizontalPodAutoscaler,
	}
}

func getAllRuntimeObject() []runtime.Object {
	return []runtime.Object{
		// core
		Pod.GetObject(),
		Namespace.GetObject(),
		Node.GetObject(),
		Event.GetObject(),
		Service.GetObject(),
		Endpoints.GetObject(),
		LimitRange.GetObject(),
		Secret.GetObject(),
		ConfigMap.GetObject(),
		ServiceAccout.GetObject(),
		PodTemplate.GetObject(),
		ResourceQuota.GetObject(),
		PersistentVolume.GetObject(),
		PersistentVolumeClaim.GetObject(),
		ReplicationController.GetObject(),

		// extensions
		Ingress.GetObject(),
		ReplicaSet.GetObject(),
		Deployment.GetObject(),
		DaemonSet.GetObject(),
		PodSecurityPolicy.GetObject(),

		// apps
		StatefulSet.GetObject(),
		ControllerRevision.GetObject(),

		// rbac
		ClusterRole.GetObject(),
		ClusterRoleBinding.GetObject(),
		Role.GetObject(),
		RoleBinding.GetObject(),

		// batch
		Job.GetObject(),
		CronJob.GetObject(),

		// storage
		StorageClass.GetObject(),

		// network
		NetworkPolicy.GetObject(),

		// autoscaling
		HorizontalPodAutoscaler.GetObject(),
	}
}

func (rs Resource) GetCanonicalName() string {
	return strings.ToLower(string(rs))
}

// GetObject gets an empty object of the resource
func (rs Resource) GetObject() runtime.Object {
	switch rs {
	// core
	case Pod:
		return &api_v1.Pod{}
	case Namespace:
		return &api_v1.Namespace{}
	case Node:
		return &api_v1.Node{}
	case Event:
		return &api_v1.Event{}
	case Service:
		return &api_v1.Service{}
	case Endpoints:
		return &api_v1.Endpoints{}
	case LimitRange:
		return &api_v1.LimitRange{}
	case Secret:
		return &api_v1.Secret{}
	case ConfigMap:
		return &api_v1.ConfigMap{}
	case ServiceAccout:
		return &api_v1.ServiceAccount{}
	case PodTemplate:
		return &api_v1.PodTemplate{}
	case ResourceQuota:
		return &api_v1.ResourceQuota{}
	case PersistentVolume:
		return &api_v1.PersistentVolume{}
	case PersistentVolumeClaim:
		return &api_v1.PersistentVolumeClaim{}
	case ReplicationController:
		return &api_v1.ReplicationController{}

	// extensions
	case Ingress:
		return &api_ext_v1beta1.Ingress{}
	case ReplicaSet:
		return &api_ext_v1beta1.ReplicaSet{}
	case Deployment:
		return &api_ext_v1beta1.Deployment{}
	case DaemonSet:
		return &api_ext_v1beta1.DaemonSet{}
	case PodSecurityPolicy:
		return &api_ext_v1beta1.PodSecurityPolicy{}

	// apps
	case StatefulSet:
		return &api_app_v1.StatefulSet{}
	case ControllerRevision:
		return &api_app_v1.ControllerRevision{}

	// rbac
	case ClusterRole:
		return &api_rbac_v1.ClusterRole{}
	case ClusterRoleBinding:
		return &api_rbac_v1.ClusterRoleBinding{}
	case Role:
		return &api_rbac_v1.Role{}
	case RoleBinding:
		return &api_rbac_v1.RoleBinding{}

	// batch
	case Job:
		return &api_batch_v1.Job{}
	case CronJob:
		return &api_batch_v2alpha1.CronJob{}

	// storage
	case StorageClass:
		return &api_store_v1.StorageClass{}

	// network
	case NetworkPolicy:
		return &api_network_v1.NetworkPolicy{}

	// autoscaling
	case HorizontalPodAutoscaler:
		return &api_autoscale_v1.HorizontalPodAutoscaler{}

	default:
		return nil
	}
}
