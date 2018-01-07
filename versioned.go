package lambda

import (
	"errors"
	"fmt"

	api_admission_v1alpha1 "k8s.io/api/admissionregistration/v1alpha1"
	api_admission_v1beta1 "k8s.io/api/admissionregistration/v1beta1"
	api_app_v1 "k8s.io/api/apps/v1beta1"
	api_autoscale_v1 "k8s.io/api/autoscaling/v1"
	api_batch_v1 "k8s.io/api/batch/v1"
	api_batch_v2alpha1 "k8s.io/api/batch/v2alpha1"
	api_v1 "k8s.io/api/core/v1"
	api_ext_v1beta1 "k8s.io/api/extensions/v1beta1"
	api_network_v1 "k8s.io/api/networking/v1"
	api_policy_v1beta1 "k8s.io/api/policy/v1beta1"
	api_rbac_v1 "k8s.io/api/rbac/v1"
	api_scheduling_v1alpha1 "k8s.io/api/scheduling/v1alpha1"
	api_settings_v1alpha1 "k8s.io/api/settings/v1alpha1"
	api_store_v1 "k8s.io/api/storage/v1"
	api_store_v1alpha1 "k8s.io/api/storage/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
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
	ReplicationController Resource = "ReplicationController"

	// extensions
	Ingress           Resource = "ingresses"
	ReplicaSet        Resource = "replicasets"
	Deployment        Resource = "deployments"
	DaemonSet         Resource = "daemonsets"
	PodSecurityPolicy Resource = "podsecuritypolicies"

	// apps
	StatefulSet        Resource = "StatefulSet"
	ControllerRevision Resource = "controllerrevisions"

	// rbac
	ClusterRole        Resource = "clusterroles"
	ClusterRoleBinding Resource = "clusterrolebindings"
	Role               Resource = "roles"
	RoleBinding        Resource = "rolebindings"

	// batch
	Job     Resource = "Job"
	CronJob Resource = "CronJob"

	// storage
	StorageClass     Resource = "StorageClass"
	VolumeAttachment Resource = "volumeattachments"

	// settings
	PodPreset Resource = "podpresets"

	// network
	NetworkPolicy Resource = "networkpolicies"

	// autoscaling
	HorizontalPodAutoscaler Resource = "horizontalpodautoscalers"

	// authentication

	// admissionregistration
	InitializerConfiguration       Resource = "initializerconfigurations"
	MutatingWebhookConfiguration   Resource = "mutatingwebhookconfigurations"
	ValidatingWebhookConfiguration Resource = "validatingwebhookconfigurations"

	// certificates

	// policy
	PodDisruptionBudget Resource = "poddisruptionbudgets"

	// scheduling
	PriorityClass Resource = "priorityclasses"
)

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
		VolumeAttachment.GetObject(),

		// settings
		PodPreset.GetObject(),

		// network
		NetworkPolicy.GetObject(),

		// autoscaling
		HorizontalPodAutoscaler.GetObject(),

		// authentication

		// admissionregistration
		InitializerConfiguration.GetObject(),
		MutatingWebhookConfiguration.GetObject(),
		ValidatingWebhookConfiguration.GetObject(),

		// policy
		PodDisruptionBudget.GetObject(),

		// scheduling
		PriorityClass.GetObject(),
	}
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
	case VolumeAttachment:
		return &api_store_v1alpha1.VolumeAttachment{}

	// settings
	case PodPreset:
		return &api_settings_v1alpha1.PodPreset{}

	// network
	case NetworkPolicy:
		return &api_network_v1.NetworkPolicy{}

	// autoscaling
	case HorizontalPodAutoscaler:
		return &api_autoscale_v1.HorizontalPodAutoscaler{}

	// authentication

	// admissionregistration
	case InitializerConfiguration:
		return &api_admission_v1alpha1.InitializerConfiguration{}
	case MutatingWebhookConfiguration:
		return &api_admission_v1beta1.MutatingWebhookConfiguration{}
	case ValidatingWebhookConfiguration:
		return &api_admission_v1beta1.ValidatingWebhookConfiguration{}

	// certificates

	// policy
	case PodDisruptionBudget:
		return &api_policy_v1beta1.PodDisruptionBudget{}

	// scheduling
	case PriorityClass:
		return &api_scheduling_v1alpha1.PriorityClass{}

	default:
		return nil
	}
}

func opInterface(rs Resource, namespace string, clientset kubernetes.Interface) (kubernetesOpInterface, error) {
	if clientset == nil {
		return nil, errors.New("nil clientset proceed")
	}
	switch rs {
	// core
	case Pod:
		return clientset.CoreV1().Pods(namespace), nil
	case Namespace:
		return clientset.CoreV1().Namespaces(), nil
	case Node:
		return clientset.CoreV1().Nodes(), nil
	case Event:
		return clientset.CoreV1().Events(namespace), nil
	case Service:
		return clientset.CoreV1().Services(namespace), nil
	case Endpoints:
		return clientset.CoreV1().Endpoints(namespace), nil
	case LimitRange:
		return clientset.CoreV1().LimitRanges(namespace), nil
	case Secret:
		return clientset.CoreV1().Secrets(namespace), nil
	case ConfigMap:
		return clientset.CoreV1().ConfigMaps(namespace), nil
	case ServiceAccout:
		return clientset.CoreV1().ServiceAccounts(namespace), nil
	case PodTemplate:
		return clientset.CoreV1().PodTemplates(namespace), nil
	case ResourceQuota:
		return clientset.CoreV1().ResourceQuotas(namespace), nil
	case PersistentVolume:
		return clientset.CoreV1().PersistentVolumes(), nil
	case PersistentVolumeClaim:
		return clientset.CoreV1().PersistentVolumeClaims(namespace), nil
	case ReplicationController:
		return clientset.CoreV1().ReplicationControllers(namespace), nil

	// extensions
	case Ingress:
		return clientset.ExtensionsV1beta1().Ingresses(namespace), nil
	case ReplicaSet:
		return clientset.ExtensionsV1beta1().ReplicaSets(namespace), nil
	case Deployment:
		return clientset.ExtensionsV1beta1().Deployments(namespace), nil
	case DaemonSet:
		return clientset.ExtensionsV1beta1().DaemonSets(namespace), nil
	case PodSecurityPolicy:
		return clientset.ExtensionsV1beta1().PodSecurityPolicies(), nil

	// apps
	case StatefulSet:
		return clientset.AppsV1().StatefulSets(namespace), nil
	case ControllerRevision:
		return clientset.AppsV1().ControllerRevisions(namespace), nil

	// rbac
	case ClusterRole:
		return clientset.RbacV1().ClusterRoles(), nil
	case ClusterRoleBinding:
		return clientset.RbacV1().ClusterRoleBindings(), nil
	case Role:
		return clientset.RbacV1().Roles(namespace), nil
	case RoleBinding:
		return clientset.RbacV1().RoleBindings(namespace), nil

	// batch
	case Job:
		return clientset.BatchV1().Jobs(namespace), nil
	case CronJob:
		return clientset.BatchV2alpha1().CronJobs(namespace), nil

	// storage
	case StorageClass:
		return clientset.StorageV1().StorageClasses(), nil
	case VolumeAttachment:
		return clientset.StorageV1alpha1().VolumeAttachments(), nil

	// settings
	case PodPreset:
		return clientset.SettingsV1alpha1().PodPresets(namespace), nil

	// network
	case NetworkPolicy:
		return clientset.NetworkingV1().NetworkPolicies(namespace), nil

	// autoscaling
	case HorizontalPodAutoscaler:
		return clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace), nil

	// authentication

	// admissionregistration
	case InitializerConfiguration:
		return clientset.AdmissionregistrationV1alpha1().InitializerConfigurations(), nil
	case MutatingWebhookConfiguration:
		return clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations(), nil
	case ValidatingWebhookConfiguration:
		return clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations(), nil

	// certificates

	// policy
	case PodDisruptionBudget:
		return clientset.PolicyV1beta1().PodDisruptionBudgets(namespace), nil

	// scheduling
	case PriorityClass:
		return clientset.SchedulingV1alpha1().PriorityClasses(), nil

	default:
		return nil, fmt.Errorf("unknown resource type %s", rs.String())
	}
}
