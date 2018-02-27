package lambda

import (
	"strings"
	"time"
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

func (rs Resource) GetCanonicalName() string {
	return strings.ToLower(string(rs))
}
