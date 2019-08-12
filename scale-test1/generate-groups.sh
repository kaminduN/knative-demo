# ROOT_PACKAGE :: the package (relative to $GOPATH/src) that is the target for code generation
ROOT_PACKAGE="scaling-demo"
# CUSTOM_RESOURCE_NAME :: the name of the custom resource that we're generating client code for
CUSTOM_RESOURCE_NAME="autoscaling"
# CUSTOM_RESOURCE_VERSION :: the version of the resource
CUSTOM_RESOURCE_VERSION="v1alpha1"

# retrieve the code-generator scripts and bins
go get -u k8s.io/code-generator/...
cd $GOPATH/src/k8s.io/code-generator

# run the code-generator entrypoint script
./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$CUSTOM_RESOURCE_NAME:$CUSTOM_RESOURCE_VERSION"

# view the newly generated files
tree $GOPATH/src/$ROOT_PACKAGE/pkg/client
# pkg/client/
# ├── clientset
# │   └── versioned
# │       ├── clientset.go
# │       ├── doc.go
# │       ├── fake
# │       │   ├── clientset_generated.go
# │       │   ├── doc.go
# │       │   └── register.go
# │       ├── scheme
# │       │   ├── doc.go
# │       │   └── register.go
# │       └── typed
# │           └── myautoscaling
# │               └── v1alpha1
# │                   ├── doc.go
# │                   ├── fake
# │                   │   ├── doc.go
# │                   │   ├── fake_myautoscaling_client.go
# │                   │   └── fake_myautoscaling.go
# │                   ├── generated_expansion.go
# │                   ├── myautoscaling_client.go
# │                   └── myautoscaling.go
# ├── informers
# │   └── externalversions
# │       ├── factory.go
# │       ├── generic.go
# │       ├── internalinterfaces
# │       │   └── factory_interfaces.go
# │       └── myautoscaling
# │           ├── interface.go
# │           └── v1alpha1
# │               ├── interface.go
# │               └── myautoscaling.go
# └── listers
#     └── myautoscaling
#         └── v1alpha1
#             ├── expansion_generated.go
#             └── myautoscaling.go
# 
# 16 directories, 22 files