# Operatify

***Operators made simple for resources with CRUD management APIs.***

Operatify provides a generic custom controller implementation for [Kubebuilder](https://book.kubebuilder.io/) operators for resources with management APIs 
consisting of CRUD (create, read, update and delete) operations, typically REST APIs. These may be for example:
* Cloud resources and services
* PaaS (Platform as a Service) services

For a full motivation and introduction, see [this blog post](https://www.stephenzoio.com/kubernetes-operators-for-resource-management/).

## Roadmap

* Example operators forthcoming... watch this space for details.

## Getting Started

### Building Operatify

A dev container is provided with all the dependencies installed for your convenience. These include:

* [Kubebuilder](https://book.kubebuilder.io/)
* [Kustomise](https://github.com/kubernetes-sigs/kustomize)
* [Kind (Kubernetes in Docker)](https://github.com/kubernetes-sigs/kind)

To make and run tests:
* Changed to the `./.devcontainer` folder.
* Run `./build.sh`
* Run `docker-conmpose up`
* In another terminal window run a bash shell in this container `docker exec -it devcontainer_docker-in-docker_1 bash`
* In this shell run `make test`

If all goes well, you should see the following output or something similar:

```text
Ran 19 of 19 Specs in 24.559 seconds
SUCCESS! -- 19 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestAPIs (24.56s)
```

### Using Operatify

Start by following the Kubebuilder Instructions [as in their tutorial](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html)

Initialise the Kubebuilder project:

```bash
kubebuilder init --domain my.domain
```

Then create a new API:

```bash
kubebuilder create api --group mygroup --version v1 --kind MyResource
```

This will ask you if you want to:
* create a resource - type `y`.
* create a controller - you can do this, but you will end up deleting the code it generates. We only need the following markers, which Kubebuilder uses:
    ```go
    // +kubebuilder:rbac:groups=mygroup.my.domain,resources=myresources,verbs=get;list;watch;create;update;patch;delete
    // +kubebuilder:rbac:groups=mygroup.my.domain,resources=myresources/status,verbs=get;update;patch
    ```

Having created a resource, we now create an operator controller for this resource.

To do so we need to:
* Implement the `ResourceManager` interface.
* Implement the `DefinitionManager` interface.
* Call the `CreateGenericController` method to create a `GenericController`.

This `GenericController` implements the `Reconciler` interface:
```go
type Reconciler interface {
	Reconcile(Request) (Result, error)
}
```

We then need to wire this into our main method. Take a look at the example `main.go` to see how this is done.
