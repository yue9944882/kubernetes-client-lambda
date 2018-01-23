package lambda

import (
	"errors"
	"fmt"
	"reflect"

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
	StorageClass Resource = "StorageClass"
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

func getResourceFromObject(t reflect.Type) (Resource, error) {
	switch t {
	// core
	case reflect.TypeOf(Pod.GetObject()):
		return Pod, nil
	case reflect.TypeOf(Namespace.GetObject()):
		return Namespace, nil
	case reflect.TypeOf(Node.GetObject()):
		return Node, nil
	case reflect.TypeOf(Event.GetObject()):
		return Event, nil
	case reflect.TypeOf(Service.GetObject()):
		return Service, nil
	case reflect.TypeOf(Endpoints.GetObject()):
		return Endpoints, nil
	case reflect.TypeOf(LimitRange.GetObject()):
		return LimitRange, nil
	case reflect.TypeOf(Secret.GetObject()):
		return Secret, nil
	case reflect.TypeOf(ConfigMap.GetObject()):
		return ConfigMap, nil
	case reflect.TypeOf(ServiceAccout.GetObject()):
		return ServiceAccout, nil
	case reflect.TypeOf(PodTemplate.GetObject()):
		return PodTemplate, nil
	case reflect.TypeOf(ResourceQuota.GetObject()):
		return ResourceQuota, nil
	case reflect.TypeOf(PersistentVolume.GetObject()):
		return PersistentVolume, nil
	case reflect.TypeOf(PersistentVolumeClaim.GetObject()):
		return PersistentVolumeClaim, nil
	case reflect.TypeOf(ReplicationController.GetObject()):
		return ReplicationController, nil

	// extensions
	case reflect.TypeOf(Ingress.GetObject()):
		return Ingress, nil
	case reflect.TypeOf(ReplicaSet.GetObject()):
		return ReplicaSet, nil
	case reflect.TypeOf(Deployment.GetObject()):
		return Deployment, nil
	case reflect.TypeOf(DaemonSet.GetObject()):
		return DaemonSet, nil
	case reflect.TypeOf(PodSecurityPolicy.GetObject()):
		return PodSecurityPolicy, nil

	// apps
	case reflect.TypeOf(StatefulSet.GetObject()):
		return StatefulSet, nil
	case reflect.TypeOf(ControllerRevision.GetObject()):
		return ControllerRevision, nil

	// rbac
	case reflect.TypeOf(ClusterRole.GetObject()):
		return ClusterRole, nil
	case reflect.TypeOf(ClusterRoleBinding.GetObject()):
		return ClusterRoleBinding, nil
	case reflect.TypeOf(Role.GetObject()):
		return Role, nil
	case reflect.TypeOf(RoleBinding.GetObject()):
		return RoleBinding, nil

	// batch
	case reflect.TypeOf(Job.GetObject()):
		return Job, nil
	case reflect.TypeOf(CronJob.GetObject()):
		return CronJob, nil

	// storage
	case reflect.TypeOf(StorageClass.GetObject()):
		return StorageClass, nil

	// network
	case reflect.TypeOf(NetworkPolicy.GetObject()):
		return NetworkPolicy, nil

	// autoscaling
	case reflect.TypeOf(HorizontalPodAutoscaler.GetObject()):
		return HorizontalPodAutoscaler, nil
	}
	return "", fmt.Errorf("unrecognized obj type: %#v", t.String())
}

func opInterface(rs Resource, namespace string, clientset kubernetes.Interface) (kubernetesOpInterface, error) {
	if clientset == nil {
		return nil, errors.New("nil clientset proceed")
	}
	apiInterface := rs.GetApiGroupInterface(clientset)
	if apiInterface == nil {
		return nil, fmt.Errorf("resource not implemented: %s", string(rs))
	}
	namespaced, err := rs.IsNamespaced(apiInterface)
	if err != nil {
		return nil, err
	}
	args := []reflect.Value{}
	if namespaced {
		args = append(args, reflect.ValueOf(namespace))
	}
	ret := reflect.ValueOf(apiInterface).Call(args)
	if len(ret) != 1 || ret[0].IsNil() {
		return nil, fmt.Errorf("unexpected return type: %s", string(rs))
	}
	return ret[0].Interface(), nil
}

func (rs Resource) IsNamespaced(i kubernetesApiGroupInterface) (bool, error) {
	if reflect.TypeOf(i).NumIn() == 1 && reflect.TypeOf(i).In(0).String() == "string" {
		return true, nil
	} else if reflect.TypeOf(i).NumIn() == 0 {
		return false, nil
	}
	return false, fmt.Errorf("invalid method signature %s", string(rs))
}

func (rs Resource) GetApiGroupInterface(clientset kubernetes.Interface) kubernetesApiGroupInterface {
	switch rs {
	// core
	case Pod:
		return clientset.CoreV1().Pods
	case Namespace:
		return clientset.CoreV1().Namespaces
	case Node:
		return clientset.CoreV1().Nodes
	case Event:
		return clientset.CoreV1().Events
	case Service:
		return clientset.CoreV1().Services
	case Endpoints:
		return clientset.CoreV1().Endpoints
	case LimitRange:
		return clientset.CoreV1().LimitRanges
	case Secret:
		return clientset.CoreV1().Secrets
	case ConfigMap:
		return clientset.CoreV1().ConfigMaps
	case ServiceAccout:
		return clientset.CoreV1().ServiceAccounts
	case PodTemplate:
		return clientset.CoreV1().PodTemplates
	case ResourceQuota:
		return clientset.CoreV1().ResourceQuotas
	case PersistentVolume:
		return clientset.CoreV1().PersistentVolumes
	case PersistentVolumeClaim:
		return clientset.CoreV1().PersistentVolumeClaims
	case ReplicationController:
		return clientset.CoreV1().ReplicationControllers

	// extensions
	case Ingress:
		return clientset.ExtensionsV1beta1().Ingresses
	case ReplicaSet:
		return clientset.ExtensionsV1beta1().ReplicaSets
	case Deployment:
		return clientset.ExtensionsV1beta1().Deployments
	case DaemonSet:
		return clientset.ExtensionsV1beta1().DaemonSets
	case PodSecurityPolicy:
		return clientset.ExtensionsV1beta1().PodSecurityPolicies

	// apps
	case StatefulSet:
		return clientset.AppsV1beta1().StatefulSets
	case ControllerRevision:
		return clientset.AppsV1beta2().ControllerRevisions

	// rbac
	case ClusterRole:
		return clientset.RbacV1().ClusterRoles
	case ClusterRoleBinding:
		return clientset.RbacV1().ClusterRoleBindings
	case Role:
		return clientset.RbacV1().Roles
	case RoleBinding:
		return clientset.RbacV1().RoleBindings

	// batch
	case Job:
		return clientset.BatchV1().Jobs
	case CronJob:
		return clientset.BatchV2alpha1().CronJobs

	// storage
	case StorageClass:
		return clientset.StorageV1().StorageClasses
	// case VolumeAttachment:
	// 	return clientset.Storage().VolumeAttachments

	// settings
	// case PodPreset:
	//	return clientset.SettingsV1alpha1().PodPresets

	// network
	case NetworkPolicy:
		return clientset.NetworkingV1().NetworkPolicies

	// autoscaling
	case HorizontalPodAutoscaler:
		return clientset.AutoscalingV1().HorizontalPodAutoscalers

	// authentication

	// admissionregistration
	// case InitializerConfiguration:
	// 	return clientset.AdmissionregistrationV1alpha1().InitializerConfigurations
	// case MutatingWebhookConfiguration:
	// 	return clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations
	// case ValidatingWebhookConfiguration:
	// 	return clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations

	// certificates

	// policy
	// case PodDisruptionBudget:
	//	return clientset.PolicyV1beta1().PodDisruptionBudgets

	// scheduling
	// case PriorityClass:
	//	return clientset.SchedulingV1alpha1().PriorityClasses

	default:
		return nil
	}
}
