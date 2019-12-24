package main

import (
	"log"
	"os"
	"net/http"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"

)
// Proxy for communication with k8s
type Proxy struct {
	clientset  *kubernetes.Clientset
	
}

// Operator interface init
type Operator interface {
	ScaleUp() error
}

// NewProxy for default init function
func NewProxy() *Proxy {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return &Proxy {
		clientset: clientset,
	}
}

// ScaleUp for interface
func (p *Proxy) ScaleUp() error {
	namespace := os.Getenv("NAME_SPACE")
	deploySpec := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{

		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "tx-test",
		},
		Spec: v1.DeploymentSpec{
			Replicas: new(int32),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta:metav1.ObjectMeta{
					Name: "nginx",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "nginx",
							Image: "nginx",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name: "http",
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	deploySpec, err := p.clientset.AppsV1().Deployments(namespace).Create(deploySpec)
	if err != nil {
		return err
	}
	return nil	
}

// ScaleUp for http server
func ScaleUp(response http.ResponseWriter, Request *http.Request) {
	var operator Operator;
	proxy := NewProxy()
	operator = proxy
	err := operator.ScaleUp()
	if (err != nil) {
		http.Error(response, err.Error(), http.StatusInternalServerError)	
	} else {
		response.WriteHeader(http.StatusOK)
	}
}