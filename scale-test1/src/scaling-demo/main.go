package main

import (
	"flag"
	"log"
	"reflect"

	"scaling-demo/pkg/apis/autoscaling/v1alpha1"
	clientset "scaling-demo/pkg/client/clientset/versioned"

	duck "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

// Pro Tip: if you want to do this in production, copy
// https://github.com/kubernetes/sample-controller which makes use of
// client-go informers and implements controller best practices.

func establishWatch() watch.Interface {
	// Watch for changes
	w, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers("default").Watch(v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return w
}

func main() {

	w := establishWatch()
	//resync := time.NewTicker(30 * time.Second).C

	for {
		//select {
		//case event, ok := <-w.ResultChan():
		event, ok := <-w.ResultChan()
		if !ok {
			log.Printf("My channel is closed. I'm going home now.")
			w.Stop()
			w = establishWatch()
			continue
		}
		pa, ok := event.Object.(*v1alpha1.PodAutoscaler)
		if !ok {
			log.Printf("Ignoring non-PodAutoscaler object %v", event.Object)
			continue
		}
		switch event.Type {
		case watch.Added:

			// Take control of yolo-class PodAutoscalers only
			if pa.Annotations["autoscaling.knative.dev/class"] == "yolo" {
				log.Printf("Selecting Deployment %q.", pa.Name)
				// Calculate a recommended scale
				replicas := recommendedScale(pa)

				// Update the Deployment
				scaleTo(pa, replicas)

				// // Update status
				updateStatus(pa, replicas)

			} else if pa.Annotations["autoscaling.knative.dev/class"] != "yolo" {

				log.Printf("Skipping Deployment %q.", pa.Name)
				log.Printf("Got PodAutoscaler %q, state %v", pa.Name, pa.Status.IsReady())
				log.Printf("Got PodAutoscaler %q, state %v", pa.Name, pa.Status.IsActivating())
				log.Printf("got auoscale %v", pa)

			}
			// }
			// case <-resync:
			// 	w.Stop()
			// 	w = establishWatch()
		default:
			log.Printf("Ignoring event %q for PodAutoscaler %q.", event.Type, pa.Name)
		}
	}
}

func recommendedScale(pa *v1alpha1.PodAutoscaler) int32 {

	// Do something really smart here ...
	return 1
}

func scaleTo(pa *v1alpha1.PodAutoscaler, replicas int32) {
	deployment, err := kubeClient.AppsV1().Deployments(pa.Namespace).Get(pa.Name+"-deployment", v1.GetOptions{})
	if err != nil {
		log.Printf("Error getting Deployment %q: %v", pa.Name, err)
		return
	}
	log.Printf("scaleTo: Got deployment %q, state %v", deployment.Name, deployment.Status)

	deployment.Spec.Replicas = &replicas
	if _, err := kubeClient.AppsV1().Deployments(pa.Namespace).Update(deployment); err != nil {
		log.Printf("Error updating Deployment %q: %v", pa.Name, err)
	}
}

func updateStatus(oldPa *v1alpha1.PodAutoscaler, replicas int32) {
	pa, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers(oldPa.Namespace).Get(oldPa.Name, v1.GetOptions{})
	if err != nil {
		log.Printf("Error getting PodAutoscaler %q: %v.", pa.Name, err)
		return
	}
	log.Printf("Got PodAutoscaler %q, oldPa %q", pa.Name, oldPa.Name)
	log.Printf("Got PodAutoscaler %q, ready state %v, old -> %v", pa.Name, pa.Status.IsReady(), oldPa.Status.IsReady())
	log.Printf("Got PodAutoscaler %q, activate state %v, old -> %v", pa.Name, pa.Status.IsActivating(), oldPa.Status.IsActivating())

	// pa.SetDefaults()
	// pa.Status.InitializeConditions()

	// pa.Status.RecommendedScale = &replicas
	// pa.Status.MarkActivating(
	// "Queued", "Requests to the target are being buffered as resources are provisioned.")

	pa.Status.MarkActive()
	pa.Status.SetConditions(duck.Conditions{{
		Type:   duck.ConditionReady,
		Status: corev1.ConditionTrue,
		Reason: "I was born ready",
	}})

	log.Printf("---pre update PodAutoscaler %q, ready state %v, old -> %v", pa.Name, pa.Status.IsReady(), oldPa.Status.IsReady())
	log.Printf("---pre PodAutoscaler %q, activate state %v, old -> %v", pa.Name, pa.Status.IsActivating(), oldPa.Status.IsActivating())
	log.Printf("---pre PodAutoscaler %q, activate state %v, old -> %v", pa.Name, pa.Status.IsInactive(), oldPa.Status.IsInactive())
	// Check if there is anything to update.
	if !reflect.DeepEqual(oldPa.Status, pa.Status) {
		existing := oldPa.DeepCopy()
		existing.Status = pa.Status
		log.Printf("xxxx auoscale old %v", oldPa)
		log.Printf("xxxx auoscale update %v", pa)
		if out, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers(pa.Namespace).Update(pa); err != nil {
			log.Printf("Error updating PodAutoscaler %q: %v.", pa.Name, err)
		} else {
			log.Printf("updating PodAutoscaler %q: %v.", pa.Name, out)
		}
	} else {
		log.Printf("skipping updating PodAutoscaler no change %q: ", pa.Name)
	}

}

var (
	kubeconfig        string
	masterURL         string
	kubeClient        kubernetes.Interface
	autoscalingClient clientset.Interface
)

func init() {

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	// Make Kubernetes and Knative Autoscaling clients.
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}
	kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
	autoscalingClient, err = clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building autoscaling clientset: %v", err)
	}
}
