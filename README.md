# Kubernetes Client Lambda #

[![Build Status](https://travis-ci.org/yue9944882/kubernetes-client-lambda.svg?branch=master)](https://travis-ci.org/yue9944882/kubernetes-client-lambda)
[![codecov](https://codecov.io/gh/yue9944882/kubernetes-client-lambda/branch/master/graph/badge.svg)](https://codecov.io/gh/yue9944882/kubernetes-client-lambda)

### What is Kubernetes Client Lambda? ###



Kubernetes Client Lambda (KCL) is a wrapper library over [kubernetes/client-go](https://github.com/kubernetes/client-go) which provides light-weight lambda-styled streamized kubernetes resource manipulation interface. This project is basically inspired by Groovy style lambda, and is aiming at reducing the coding-overhead when referencing too many struct / interface provided by  [kubernetes/client-go](https://github.com/kubernetes/client-go). The only dependency of this project is [kubernetes/client-go](https://github.com/kubernetes/client-go), so this project can be pure and light-weight for you. Currently KCL only provides support for those common-used resources like Pod / Service.. [Click](https://github.com/yue9944882/kubernetes-client-lambda/blob/cfaa5564df0a4212ef9230be9ddd05a5c7034916/resource.go#L9) to see all the supported resources in KCL. 

Also, KCL defines more useful primitive operation type beyond those provided, e.g. Create-If-Not-Exist, Update-Or-Create, Delete-If-Exist.., which can help you make your code more clean. These primitive operation is always at the end of streaming and consumes all the remaining element and apply operation on every element. 

Basically, KCL is implemented by golang `reflect` and golang native facilities. KCL defines various resources provided by kubernetes as enumeratoin which hides some unneccessary details like API-version. Sometimes, managing these details can be painful for kubernetes developers. Meanwhile, I find it costs time when mocking kubernetes resources in static unit test, so KCL provides a very simple mocking for kubernetes resources which is implemented via `map`. 

With KCL, you can operate kubernetes resources like this example:

```go
// See doc for more info about lambda functions Grep / Map..

// In-Cluster example
ReplicaSet.InCluster().InNamespace("test").Grep(func(rs *api_ext_v1.ReplicaSet) bool {
    // Assuming we already have foo-v001, foo-v002, bar-v001 
    return strings.HasPrefix(rs.Name, "foo-")
}).Map(func(rs *api_ext_v1.ReplicaSet) rs*api_ext_v1.ReplicaSet {
    // Edit in-place or clone a new one
    rs.Meta.Labels["foo-label1"] = "test" 
    return rs
}).Update()


// Out-Of-Cluster example
ReplicaSet.OutOfCluster(rest_config).InNamespace("test").Grep(func(rs *api_ext_v1.ReplicaSet) bool {
    // Assuming we already have foo-v001, foo-v002, bar-v001 
    return strings.HasPrefix(rs.Name, "foo-")
}).Map(func(rs *api_ext_v1.ReplicaSet) rs*api_ext_v1.ReplicaSet {
    // Edit in-place or clone a new one
    rs.Meta.Labels["foo-label1"] = "test" 
    return rs
}).Update()
```

### How to Use it? ###

```
go get yue9944882/kubernetes-client-lambda
```

### Why Kubernetes Client Lambda is better? ###

- Manipulating kubernetes resources in one line
- Lambda-styled kubernetes resource processing.
- Pipelined and streamlized.
- Light-weight and only depends on [kubernetes/client-go](https://github.com/kubernetes/client-go)
- User-friendly mocking kubernetes static interface

### How to Mock Kubernetes Resources? ###

As the following example shown, Calling `Mock(autoCreateNamespace bool)` on Kubernetes Type Enumeration will create the expected mocking resources for you:

```go
var rs api_ext_v1.ReplicaSet
autoCreateNamespace := false
ReplicaSet.Mock(autoCreateNamespace).InNamespace("test").Add(
    // An anonymous function simply returns a pointer to kubernetes resource 
    // Returned objects will be added to stream
    func(){
        rs.Name = "foo"
        rs.Namespace = "test"
        return &rs
    },
).Create()
```


Checkout more examples under `example` folder.


### Supported Lambda Function Type  ###

First we have following types of lambda function: 

(KR denotes Kubernetes Resources, a pointer to resouce, e.g. *api_v1.Pod, *api_ext_v1.ReplicaSet..)

| Index | Parameter1 | Parameter2 | Parameter3 | Return1 |
|---|---|---|---|---|
| 1 | - | - | - | KR |
| 2 | KR | - | - | KR |
| 3 | KR | - | - | bool |
| 4 | KR | - | - | - |


And these lambda can be consumed by: 


| Name | Pipelinable | Lambda Type | Description |
|---|---|----|---|
| Add | yes | 1 | Add the element returned by lambda into collection |
| Map | yes | 2 | Add all the elements returned by lambda to a new collection |
| Grep | yes | 3 | Remove the element from collection if applied lambda returned a `false` |
| First | yes | 3 | Take only the first element when applied lambda returned a `true` |
| Each | no | 4 | Apply the lambda to every elements in the collection |
| Any | no | 3 | Return true if at least one element when applied lambda returned a `true` | 
| Every | no | 3 | Return true if every element when applied lambda returned a `true` | 


Primitive methods like `CreateIfNotExist`, `DeleteIfExist` have no parameter and just consumes all elements at the end of the pipelining. 
Here are supported primitive kubernetes operation functions below:


| Operation | Param | Return1 | Return2 | 
|---|---|---|---|
| Create | - | bool(sucess) | lambda error |
| CreateIfNotExists | - | bool(sucess) | lambda error |
| Delete | - | bool(sucess) | lambda error |
| DeleteIfExists | - |  bool(sucess) | lambda error |
| Update | - |  bool(sucess) | lambda error |
| UpdateIfExists | - |  bool(sucess) | lambda error |




### Help & Contact ###


 E-mail: yue9944882@gmail.com


