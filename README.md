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

If all goes according to plan, you should see the following output or something similar:

```text
Ran 19 of 19 Specs in 24.559 seconds
SUCCESS! -- 19 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestAPIs (24.56s)
```

You can also spin up a Kind cluster and run tests against that:
    
```
make set-kindcluster
USE_EXISTING_CLUSTER="true" make test
```

### Using Operatify

Start by following the Kubebuilder Instructions [as in their tutorial](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html)

1. Initialise the Kubebuilder project:

    ```bash
    kubebuilder init --domain my.domain
    ```

2. Add the following require to your `go.mod` file (use the latest release version rather than v0.1.1 below).

    ```go
    github.com/operatify/operatify v0.1.1
    ```
    
3. Add the following import where required:

    ```go
    import "github.com/operatify/operatify/reconciler"
    ```
    
4. Create a new API:

    ```bash
    kubebuilder create api --group mygroup --version v1 --kind MyResource
    ```
    
    This will ask you if you want to:
    * create a resource (y/n) - type `y`.
    * create a controller (y/n) - you can do this, but you will end up deleting the code it generates. We only need the following markers, which Kubebuilder uses:
    ```go
    // +kubebuilder:rbac:groups=mygroup.my.domain,resources=myresources,verbs=get;list;watch;create;update;patch;delete
    // +kubebuilder:rbac:groups=mygroup.my.domain,resources=myresources/status,verbs=get;update;patch
    ```
    
5. Create an operator controller for this resource.

    To do so we need to:
    * Implement the `ResourceManager` interface.
    * Implement the `DefinitionManager` interface.
    * Call the `CreateGenericController` method to create a `GenericController`.
    
    This `GenericController` implements the `Reconciler` interface of the Kubernetes controller runtime:
    ```go
    type Reconciler interface {
    	Reconcile(Request) (Result, error)
    }
    ```

6. Wire this into our main method. 

    Take a look at the example `main.go` to see how this is done.

## Implementation details

### Resource diffing

Every time the `Create` or `Update` method returns successfully, 
an `[annotation-base-name]/last-applied-spec` 
annotation is saved with the Json representation of the `spec` that was used to create or update the resource. 

### Passing back status data

The `Create`, `Update` and `Verify` can also return an extra status payload return parameter. 
This is an `interface{}`, so can be anything. The idea is this should be saved as a status field in the manifest. 
The reconciler will not care about it, but can be passed back into the subsequent calls to `Verify`.

#### Locking down access control

It is possible to restrict acess control to certain external resources to prevent unintended modifications and deletes.
The reconciler recognises an annotation `[annotation-base-name]/access-permissions`, 
in which it recognises each one of the permissions, create, update and delete via the initial. 
Read permission is implicit. For example `"CD"` is permission to create and delete, `"CUD"` is everything (the default if this annotation is not defined), and anything that doesn't have these initials (e.g. `"none"`)
 is read-only permissions.
 
Read-only permission allows one to assert the state of a dependent resource without being able to modify or deleting it.

If the delete permission is not set, it will simply not delete the external resource when the Kubernetes resource is delete.
However if the `Verify` method returns `VerifyResultRecreateRequired` and delete permission is not present, it will return an error.

#### Implementing a handler upon success

Sometimes after creating or updating a resource, further interaction with Kubenetes is necessary.

For example an operator that provisions a database may need to create a Kubernetes secret with database credentials.  

To facilitate this, a hook to invoke these interactions which gets called, if it is defined, after a successful create or update operation.

To define this hook, we need to call the `CreateGenericController` with a non-nil `completetionRunner` parameter.

This is more accurately a factory method for a `CompletionRunner`:

```go
type CompletionRunner interface {
	Run(ctx context.Context, r runtime.Object) error
}
```

The implementation of this operation must be idempotent.

Because the factory method gets the `GenericController` instance, it has full access to both the Kubernetes client and the `ResourceManager` instance.
This the only instance where such power is given to the user. 
