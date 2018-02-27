package lambda

import (
	"time"
)

const (
	// core
	Pod                   Resource = "Pod"
	Namespace             Resource = "Namespace"
	Node                  Resource = "Node"
	Event                 Resource = "Event"
	Service               Resource = "Service"
	Endpoints             Resource = "Endpoints"
	LimitRange            Resource = "LimitRange"
	Secret                Resource = "Secret"
	ConfigMap             Resource = "ConfigMap"
	ServiceAccout         Resource = "ServiceAccount"
	PodTemplate           Resource = "PodTemplate"
	ResourceQuota         Resource = "ResourceQuota"
	PersistentVolume      Resource = "PersistentVolume"
	PersistentVolumeClaim Resource = "PersistentVolumeClaim"
	ReplicationController Resource = "ReplicationController"

	// extensions
	Ingress           Resource = "Ingress"
	ReplicaSet        Resource = "ReplicaSet"
	Deployment        Resource = "Deployment"
	DaemonSet         Resource = "DaemonSet"
	PodSecurityPolicy Resource = "PodSecurityPolicy"

	// apps
	StatefulSet        Resource = "StatefulSet"
	ControllerRevision Resource = "ControllerRevision"

	// rbac
	ClusterRole        Resource = "ClusterRole"
	ClusterRoleBinding Resource = "ClusterRoleBinding"
	Role               Resource = "Role"
	RoleBinding        Resource = "RoleBinding"

	// batch
	Job     Resource = "Job"
	CronJob Resource = "CronJob"

	// storage
	StorageClass Resource = "StorageClass"
	// VolumeAttachment Resource = "volumeattachments"

	// settings
	// PodPreset Resource = "podpresets"

	// network
	NetworkPolicy Resource = "NetworkPolicy"

	// autoscaling
	HorizontalPodAutoscaler Resource = "HorizontalPodAutoscaler"

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

		// comment alpha resource temporary
		// CronJob,

		// storage
		StorageClass,

		// network
		NetworkPolicy,

		// autoscaling
		HorizontalPodAutoscaler,
	}
}
