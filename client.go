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
	ScaleUp(index string) error
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
func (p *Proxy) ScaleUp(index string) error {
	namespace := os.Getenv("NAME_SPACE")
	version := os.Getenv("VERSION")
	// create statefulset
	configName := "backend"
	labels := map[string]string {
		"name": "test",
		"shard": index,
	}
	statefulSpec := &v1.StatefulSet{
		TypeMeta: metav1.TypeMeta{

		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "url-" + index,
		},
		Spec: v1.StatefulSetSpec{
			Replicas: new(int32),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: "url-" + index,
			Template: corev1.PodTemplateSpec{
				ObjectMeta:metav1.ObjectMeta{
					Name: "url",
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name: "url-operator",
							Image: "ty0207/link-server:v"+version,
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name: "http",
									ContainerPort: 9090,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name: configName,
									MountPath: "/root/config",
								},
							},
						},
					}, 
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: configName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	*statefulSpec.Spec.Replicas = 1;
	serviceSpec := &corev1.Service{
		TypeMeta: metav1.TypeMeta{

		},
		ObjectMeta:metav1.ObjectMeta{
			Name: "url",
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "url",
					Port: 9090, 
				},
			},
			Selector: labels,
		},
	}
	p.clientset.CoreV1().Services(namespace).Create(serviceSpec)
	statefulSpec, err := p.clientset.AppsV1().StatefulSets(namespace).Create(statefulSpec)
	if err != nil {
		return err
	}
	// create service
	return nil	
}

// ScaleUp for http server
func ScaleUp(response http.ResponseWriter, request *http.Request) {
	var operator Operator;
	proxy := NewProxy()
	operator = proxy
	_ = request.ParseForm()
	index := request.FormValue("index")
	if (index == "") {
		http.Error(response, "index parameter must have", http.StatusBadRequest)
	}
	err := operator.ScaleUp(index)
	if (err != nil) {
		http.Error(response, err.Error(), http.StatusInternalServerError)	
	} else {
		response.WriteHeader(http.StatusOK)
	}
}