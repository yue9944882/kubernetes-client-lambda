# Kubernetes Client Lambda #

[![Build Status](https://travis-ci.com/yue9944882/kubernetes-client-lambda.svg?token=pzmnymNtKXSyQJpSi3Kq&branch=master)](https://travis-ci.com/yue9944882/kubernetes-client-lambda)

### What is Kubernetes Client Lambda? ###



Kubernetes Client Lambda (KCL) is a wrapper library over [kubernetes/client-go](https://github.com/kubernetes/client-go) which provides light-weight lambda-styled streamized kubernetes resource manipulation interface. This project is basically inspired by Groovy style lambda, and is aiming at reducing the coding-overhead when referencing too many struct / interface provided by  [kubernetes/client-go](https://github.com/kubernetes/client-go). The only dependency of this project is [kubernetes/client-go](https://github.com/kubernetes/client-go), so this project can be pure and light-weight for you. Currently KCL only provides those common-used resources like Pod / Service.. [Click](https://github.com/yue9944882/kubernetes-client-lambda/blob/cfaa5564df0a4212ef9230be9ddd05a5c7034916/resource.go#L9) to see all the supported resources in KCL.

Basically, KCL is implemented by golang `reflect` and golang native facilities. KCL defines various resources provided by kubernetes as enumeratoin which hides some unneccessary details like API-version. Sometimes, managing these details can be painful for kubernetes developers. Meanwhile, I find it costs time when mocking kubernetes resources in static unit test, so KCL provides a very simple mocking for kubernetes resources which is implemented via `map`. 

With KCL, you can operate kubernetes resources like this example:

```
// See doc for more info about lambda functions Grep / Map..

ReplicaSet.InNamespace("test").Grep(func(rs *api_ext_v1.ReplicaSet) bool {
    // Assuming we already have foo-v001, foo-v002, bar-v001 
    return strings.HasPrefix(rs.Name, "foo-")
}).Map(func(rs *api_ext_v1.ReplicaSet) rs*api_ext_v1.ReplicaSet {
    // Edit in-place or clone a new one
    rs.Meta.Labels["foo-label1"] = "test" 
    return rs
}).Update()
```

### Why Kubernetes Client Lambda is better? ###

- Lambda-styled Kubernetes resource manipulating.
- Pipelined and streamlized.
- Light-weight and only depends on [kubernetes/client-go](https://github.com/kubernetes/client-go)
- User-friendly mocking kubernetes static interface

### How to Mock Kubernetes Resources? ###

```
ReplicaSet.Mock().InNamespace("")
```