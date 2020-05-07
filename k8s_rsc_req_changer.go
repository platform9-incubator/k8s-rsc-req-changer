package main

import (
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	var kubeconfig *string
	kcEnvVar := os.Getenv("KUBECONFIG")
	if kcEnvVar != "" {
		kubeconfig = flag.String("kubeconfig", kcEnvVar, "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	containerNamePrefix := flag.Arg(0)
	if containerNamePrefix == "" {
		panic("container name prefix not provided")
	}

	t := flag.Arg(1)
	if t == "" {
		panic("resource name not provided")
	}
	rscName := v1.ResourceName(t)
	if rscName != v1.ResourceCPU && rscName != v1.ResourceMemory {
		panic("invalid resource name")
	}
	
	reqValue := flag.Arg(2)
	if reqValue == "" {
		panic("new request value not provided")
	}
	newVal, err := resource.ParseQuantity(reqValue)
	if err != nil {
		panic(err.Error())
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	corev1 := clientset.CoreV1()
	nses, err := corev1.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	extv1beta1 := clientset.ExtensionsV1beta1()
	for _, ns := range nses.Items {
		nsName := ns.Name
		depApi := extv1beta1.Deployments(nsName)
		deps, err := depApi.List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		//fmt.Printf("There are %d deployments in namespace %s\n", len(deps.Items), nsName)
		for _, dep := range deps.Items {
			changed := false
			for _, cont := range dep.Spec.Template.Spec.Containers {
				if !strings.HasPrefix(cont.Name, containerNamePrefix) {
					continue
				}
				path := fmt.Sprintf("%s/%s/%s", nsName, dep.Name, cont.Name)
				if val, ok := cont.Resources.Requests[rscName]; ok {
					if newVal.String() == "0" {
						fmt.Printf("%s: removing %s request of %s\n",
							path, rscName, val.String())
						delete(cont.Resources.Requests, rscName)
						changed = true
					} else if val != newVal {
						fmt.Printf("%s: changing %s request of %s to %s\n",
							path, rscName, val.String(), newVal.String())
						cont.Resources.Requests[rscName] = newVal
						changed = true
					}
				} else if newVal.String() != "0" {
					fmt.Printf("%s: adding %s request of %s\n",
						path, rscName, newVal.String())
					cont.Resources.Requests[rscName] = newVal
					changed = true
				}
			}
			if changed {
				_, err := depApi.Update(&dep)
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}
}


