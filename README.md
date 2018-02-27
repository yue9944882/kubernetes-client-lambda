# Kubernetes Client Lambda #

[![Build Status](https://travis-ci.org/yue9944882/kubernetes-client-lambda.svg?branch=master)](https://travis-ci.org/yue9944882/kubernetes-client-lambda)
[![codecov](https://codecov.io/gh/yue9944882/kubernetes-client-lambda/branch/master/graph/badge.svg)](https://codecov.io/gh/yue9944882/kubernetes-client-lambda)
[![Go Doc](https://godoc.org/github.com/yue9944882/kubernetes-client-lambda?status.svg)](https://godoc.org/github.com/yue9944882/kubernetes-client-lambda)


![logo](image/logo.png)

### What is Kubernetes Client Lambda? ###


Kubernetes Client Lambda (aka KCL) is a wrapper library for [kubernetes/client-go](https://github.com/kubernetes/client-go). Basically it contains these following feature:

- Dynamic client & client pool
- Hide details about client-go's informer and lister.
- Hide annoying group & versions and use resources as enum.
- Lambda styled resource filtering & manipulating. (inspired by Groovy)
- It's really easy to use.

With KCL, you can operate kubernetes resources like this example:

```go
// See doc for more info about lambda functions Grep / Map..
import kubernetes "github.com/yue9944882/kubernetes-client-lambda"

// In-Cluster example
// var kcl kubernetes.KubernetesClientLambda = kubernetes.InCluster()
kubernetes.InCluster().Type(kubernetes.ReplicaSet).InNamespace("test").NamePrefix("foo-").Map(func(rs *api_ext_v1.ReplicaSet) rs*api_ext_v1.ReplicaSet {
    // Edit in-place or clone a new one
    rs.Meta.Labels["foo-label1"] = "test" 
    return rs
}).Update()


// Out-Of-Cluster example
kubernetes.OutOfClusterDefault().Type(kubernetes.Pod).InNamespace("devops").NameEqual("test-pod").Each(
    func(pod *api_v1.Pod) {
        count++
})
```

As the following example is shown, Calling `Mock()` on Kubernetes Type Enumeration will create the expected mocking resources for you:

```go
import kubernetes "github.com/yue9944882/kubernetes-client-lambda"

var kcl KubernetesClientLambda = kubernetes.Mock()
```

### How to Get it? ###

```
go get github.com/yue9944882/kubernetes-client-lambda
```

### Supported Lambda Function Type ###

We support following types of lambda function: 

##### Primitive Lambda Type #####

<a name="lambda-type"></a>

| Name | Parameter Type | Return Type |
|---|---|---|
| Function | Resource | - |
| Consumer | Resource |  |
| Predicate | Resource | bool |
| Producer | - | Resource |

##### Kubernetes Resource Lambda Snippet #####

| Name | Pipelinable | Description |
|---|---|----|
| NameEqual | yes | Filter out resources if its name mismatches |
| NamePrefix | yes | Filter out resources if its name doesn't have the prefix |
| NameRegex | yes | Filter out resources if its name doesn't match the regular expression |
| HasAnnotation | yes | Filter out resources if it doesn't have the annotation |
| HasAnnotationKey | yes | Filter out resources if it doesn't have the annotation key |
| HasLabel | yes | Filter out resources if it doesn't have the label |
| HasLabelKey | yes | Filter out resources if it doesn't have the label key |


And these lambda can be consumed by following function: 


<a name="pipeline-type"></a>
##### Primitive Pipeline Type #####

| Name | Pipelinable | Lambda Type | Description |
|---|---|----|---|
| Collect | yes | - | Deep copies the elements and put them into collection | 
| Add | yes | Producer | Add the element returned by lambda into collection |
| Map | yes | Consumer | Add all the elements returned by lambda to a new collection |
| Grep | yes | Predicate | Remove the element from collection if applied lambda returned a `false` |
| First | yes | Predicate | Take only the first element when applied lambda returned a `true` |
| Iter | no | Function | Apply the lambda to every elements in the collection |


Primitive methods like `CreateIfNotExist`, `DeleteIfExist` have no parameter and just consumes all elements at the end of the pipelining. 
Here are supported primitive kubernetes operation functions below:

##### Basic Operation #####

| Operation | Param | Return1 | Return2 | 
|---|---|---|---|
| Each | Function | lambda error | - |
| Any | Predicate | bool | lambda error |
| Every | Predicate | bool | lambda error |
| NotEmpty | - | bool | lambda error |

##### Kubernetes Operation #####

| Operation | Param | Return1 | Return2 | 
|---|---|---|---|
| Create | - | bool(sucess) | lambda error |
| CreateIfNotExists | - | bool(success) | lambda error |
| Delete | - | bool(sucess) | lambda error |
| DeleteIfExists | - |  bool(success) | lambda error |
| Update | - |  bool(sucess) | lambda error |
| UpdateIfExists | - |  bool(success) | lambda error |
| UpdateOrCreate | - | bool(success) | lambda error |


