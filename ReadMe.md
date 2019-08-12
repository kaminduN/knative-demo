Follow the following steps to install knative
https://knative.dev/docs/install/knative-with-gke/


gcloud beta container clusters create $CLUSTER_NAME \
  --addons=HorizontalPodAutoscaling,HttpLoadBalancing,Istio \
  --machine-type=n1-standard-4 \
  --cluster-version=latest --zone=$CLUSTER_ZONE \
  --enable-stackdriver-kubernetes --enable-ip-alias \
  --enable-autoscaling --min-nodes=1 --max-nodes=2 \
  --scopes cloud-platform

kubectl create clusterrolebinding cluster-admin-binding \
     --clusterrole=cluster-admin \
     --user=$(gcloud config get-value core/account)

kubectl label namespace default istio-injection=enabled

-----------------------------------------
To install the controller 
the build is done using "ko" (https://github.com/google/ko) tool

cd scale-test1
export GOPATH=`pwd`
export KO_DOCKER_REPO=gcr.io/xxxx

ko apply -f artifacts/controller.yaml

--------------
to install the applicaiton
https://knative.dev/docs/install/getting-started-knative-app/

kubectl apply -f autoscale-go/service-deploy.yaml


-----------------------------------------

To test the application

export IP_ADDRESS=`kubectl get svc istio-ingressgateway --namespace istio-system --output jsonpath=' .  {.status.loadBalancer.ingress[*].ip}'`


curl --header 'Host: app.default.example.com' "http://${IP_ADDRESS?}?sleep=100&prime=10000&bloat=5"


