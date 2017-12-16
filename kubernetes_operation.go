package lambda

// KubernetesOperation is an wrapper interface for k8s client
type KubernetesOperation interface {
	opCreateInterface(kubernetesResource) (kubernetesResource, error)
	opDeleteInterface(string) error
	opUpdateInterface(kubernetesResource) (kubernetesResource, error)

	opGetInterface(string) (kubernetesResource, error)
	opListInterface() ([]kubernetesResource, error)
}
