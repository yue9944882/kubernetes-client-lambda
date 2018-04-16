package lambda

import (
	"time"
)

var (
	// core
	Pod                   = Resource{"Pods", ""}
	Namespace             = Resource{"Namespaces", ""}
	Node                  = Resource{"Nodes", ""}
	Event                 = Resource{"Events", ""}
	Service               = Resource{"Services", ""}
	Endpoints             = Resource{"Endpoints", ""}
	LimitRange            = Resource{"LimitRanges", ""}
	Secret                = Resource{"Secrets", ""}
	ConfigMap             = Resource{"ConfigMaps", ""}
	ServiceAccout         = Resource{"ServiceAccounts", ""}
	PodTemplate           = Resource{"PodTemplates", ""}
	ResourceQuota         = Resource{"ResourceQuotas", ""}
	PersistentVolume      = Resource{"PersistentVolumes", ""}
	PersistentVolumeClaim = Resource{"PersistentVolumeClaims", ""}
	ReplicationController = Resource{"ReplicationControllers", ""}

	// extensions
	Ingress           = Resource{"Ingresses", ""}
	ReplicaSet        = Resource{"ReplicaSets", "v1beta1"}
	Deployment        = Resource{"Deployments", ""}
	DaemonSet         = Resource{"DaemonSets", ""}
	PodSecurityPolicy = Resource{"PodSecurityPolicies", ""}

	// apps
	StatefulSet        = Resource{"StatefulSets", ""}
	ControllerRevision = Resource{"ControllerRevisions", ""}

	// rbac
	ClusterRole        = Resource{"ClusterRoles", ""}
	ClusterRoleBinding = Resource{"ClusterRoleBindings", ""}
	Role               = Resource{"Roles", ""}
	RoleBinding        = Resource{"RoleBindings", ""}

	// batch
	Job     = Resource{"Jobs", ""}
	CronJob = Resource{"CronJobs", ""}

	// storage
	StorageClass = Resource{"StorageClasses", ""}

	// settings

	// network

	// autoscaling
	HorizontalPodAutoscalerV1 = Resource{"HorizontalPodAutoscalers", "v1"}
	HorizontalPodAutoscalerV2 = Resource{"HorizontalPodAutoscalers", "v2beta1"}

	// authentication

	// admissionregistration

	// certificates

	// policy

	// scheduling
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

		// comment alpha resource temporary
		// CronJob,

		// storage
		StorageClass,

		// network
		// NetworkPolicy,

		// autoscaling
		HorizontalPodAutoscalerV1,
		HorizontalPodAutoscalerV2,
	}
}

func (r Resource) GetKind() string {
	return indexerInstance.GetGroupVersionKind(r).Kind
}

func (r Resource) GetResource() string {
	return indexerInstance.GetGroupVersionResource(r).Resource
}

func (r Resource) GetAPIVersion() string {
	return indexerInstance.GetGroupVersionKind(r).GroupVersion().String()
}
