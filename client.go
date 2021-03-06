package main

import (
	"fmt"
	"bytes"
	"os"
	"io/ioutil"
	"net/http"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
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
		fmt.Println(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	return &Proxy {
		clientset: clientset,
	}
}

// ScaleUp for interface
func (p *Proxy) ScaleUp(index string) error {
	namespace := os.Getenv("NAME_SPACE")
	version := os.Getenv("VERSION")
	configMapName := "config-"+ index
	deploymentName := "url-"+ index
	mysqlName := "mysql-" + index
	mysqlServiceName := mysqlName
	serviceName := deploymentName
	labels := map[string]string {
		"name": "url",
		"shard": index,
	}
	// create configmap
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: map[string]string {
			"config.yml": "mysql:\n" +
			"  ipport: mysql-"+ index +".default:3306\n" +
			"  username: root\n" + 
			"  password:\n" + 
		    "redis:\n" +
			"  ipport: localhost:6379",
		},
	}
	fmt.Println("configmap")
	configmap, err := p.clientset.CoreV1().ConfigMaps(namespace).Create(configmap)
	if err != nil {
		fmt.Println("configmap")
		return err
	}
	// create deployment from yaml
	urlbytes,err := ioutil.ReadFile("/config/url-redis.yaml")
	if err != nil {
		fmt.Println("/config/url-redis.yaml")
		return err
	}
	if err != nil {
		fmt.Println("ioutil url-redis")
		return err
	}
	deployment := &v1.Deployment{}
	fmt.Println("deployment")
	err = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(urlbytes),len(urlbytes)).Decode(&deployment)
	if err != nil {
		fmt.Println("yaml decoder deployment")
		return err
	}
	deployment.ObjectMeta.Name = deploymentName
	deployment.Spec.Selector = &metav1.LabelSelector {
		MatchLabels: labels,
	}
	deployment.Spec.Template.ObjectMeta.Labels = labels
	deployment.Spec.Template.Spec.Volumes = []corev1.Volume {
		corev1.Volume{
			Name: configMapName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource {
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		},
	}
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		corev1.VolumeMount{
			Name: configMapName,
			MountPath: "/root/config",
		},
	}
	deployment.Spec.Template.Spec.Containers[0].Image = "ty0207/link-server:v"+ version 
	// create statefulset
	fmt.Println("containers")
	statefulSpec := &v1.StatefulSet{}
	statefulfile,err := ioutil.ReadFile("/config/mysql.yaml")
	if err != nil {
		fmt.Println("/config/mysql.yaml")
		return err
	}
	fmt.Println("mysql")
	err = yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(statefulfile),len(statefulfile)).Decode(statefulSpec)
	fmt.Println("statefulSpec")
	if err != nil {
		fmt.Println("yaml decoder mysql")
		return err
	}
	statefulSpec.Spec.Replicas = new (int32)
	*statefulSpec.Spec.Replicas = 1
	statefulSpec.ObjectMeta.Name = mysqlName
	
	fmt.Println("configmap")
	statefulSpec.Spec.ServiceName = mysqlServiceName
	mysqlLabels := map[string]string {
		"name": "mysql",
		"shard": index,
	}
	statefulSpec.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: mysqlLabels,
	}
	statefulSpec.Spec.Template.ObjectMeta.Labels = mysqlLabels
	serviceSpec := &corev1.Service{
		TypeMeta: metav1.TypeMeta{

		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
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
	mysqlserviceSpec := &corev1.Service{
		TypeMeta: metav1.TypeMeta{

		},
		ObjectMeta:metav1.ObjectMeta{
			Name: mysqlServiceName,
			Labels: mysqlLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "mysql",
					Port: 3306, 
				},
			},
			Selector: mysqlLabels,
		},
	}
	fmt.Println("deployment")
	deployment, err = p.clientset.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		fmt.Println("deployment create")
		return err
	}
	fmt.Println("serviceSpec")
	serviceSpec, err = p.clientset.CoreV1().Services(namespace).Create(serviceSpec)
	if err != nil {
		fmt.Println("service create")
		return err
	}
	// create service
	fmt.Println("mysqlserviceSpec")
	mysqlserviceSpec,err = p.clientset.CoreV1().Services(namespace).Create(mysqlserviceSpec)
	if err != nil {
		fmt.Println("mysql service create")
		return err
	}
	fmt.Println("statfulSpec")
	statefulSpec, err = p.clientset.AppsV1().StatefulSets(namespace).Create(statefulSpec)
	if err != nil {
		fmt.Println("statefulset create")
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
		return
	}
	err := operator.ScaleUp(index)
	if (err != nil) {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	} else {
		response.WriteHeader(http.StatusOK)
	}
}