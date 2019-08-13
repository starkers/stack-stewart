package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" //GCP support

	"os"
	"strings"

	"github.com/starkers/stack-stewart/api"
	"github.com/starkers/stack-stewart/shared"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientSet returns a kubernetes clientset
func GetClientSet(cfgKubeConfig string) *kubernetes.Clientset {
	if cfgKubeConfig == "nope" {
		log.Info("loading in-cluster")

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
		return clientset
	}
	if _, err := os.Stat(cfgKubeConfig); os.IsNotExist(err) {
		log.Errorf("%s does not exist", cfgKubeConfig)
		os.Exit(1)
	}
	log.Info("using KUBECONFIG: ", cfgKubeConfig)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", cfgKubeConfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

// GetMyNamespace determines the namespace (string) the pod is running in
func GetMyNamespace() string {
	// if this code is running in k8s it will return the namespace as a string
	// otherwise it will fallback on the $NAMESPACE env or default namespace
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	// giving up on native kubernetes.. maybe this env var is set?
	if ns := os.Getenv("NAMESPACE"); ns != "" {
		return ns
	}
	// no idea.. lets just use the default namespace.. probably "default"
	return metav1.NamespaceDefault
}

// SecretKeyBootstrap creates an initial k8s secret if one isn't found
func SecretKeyBootstrap(client *kubernetes.Clientset, namespace string, secretName string, key string) error {
	// creates an initial secret (for agent mode) if it is not present
	log.Info("searching for existing secret key")
	secretList, err := client.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	matched := false
	for i, j := range secretList.Items {
		// loops over secrets in the namespace, looking to set matched = true
		// if it finds a secret matching the name of secretName
		name := j.GetName()
		log.Debugf("secrets (in ns: %s): [%d] %s", namespace, i, name)

		// if we have a secret which name matches the one we expect...
		if name == secretName {
			matched = true
		}
	}
	if matched {
		log.Infof("found a secret called '%s', no need to create a new one", secretName)
		return nil
	}
	// no secret found
	// create a secret
	// TODO: generate something random in the data
	uuidRaw := uuid.NewV4()
	uuidString := fmt.Sprintf("%s", uuidRaw)
	_, err = client.CoreV1().Secrets(namespace).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		StringData: map[string]string{
			key: uuidString,
		},
	})
	if err != nil {
		return err
	}
	log.Infof("successfully created secret/%s", secretName)
	return nil
}

// GetSecretKeyData returns a string from secret['token']
func GetSecretKeyData(client *kubernetes.Clientset, namespace string, secretName string, key string) (string, error) {
	log.Debugf("attempting to get the secret from: secret/%s (in namespace: %s) data key: %s", secretName, namespace, key)
	secret, err := client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	//convert the secret data into a string
	valueByte := secret.Data[key]
	valueString := string(valueByte)
	log.Debugf("read token '%s' from secret", valueString)
	return valueString, err
}

//// GetDeployments gets some stuff from deployments
//func GetDeployments(client *kubernetes.Clientset, namespace string) {
//	deploymentClient := client.AppsV1().Deployments(namespace)
//	list, err := deploymentClient.List(metav1.ListOptions{})
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//
//	for _, d := range list.Items {
//		// fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
//		name := d.Name
//		replicas := *d.Spec.Replicas
//		log.Printf("%s/%s (replicas: %d)", namespace, name, replicas)
//		for _, c := range d.Spec.Template.Spec.Containers {
//			log.Printf("container name: %s, image: %v", c.Name, c.Image)
//		}
//
//	}
//
//}

// SendDeployments sends intel about a Deployment to the PostStack function
func SendDeployments(
	client *kubernetes.Clientset,
	namespace,
	lane,
	tokenString,
	apiServer string,
	CfgClusterNickname string,
) {
	log.Debugf("searching ns/%s for deployments", namespace)
	deploymentClient := client.AppsV1().Deployments(namespace)
	list, err := deploymentClient.List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	log.Debugf("found %d deployments in namespace: %s", len(list.Items), namespace)

	// Loop over deployments
	for _, d := range list.Items {
		s := shared.Stack{}
		// Loop over the containers in the deployment
		for _, c := range d.Spec.Template.Spec.Containers {
			log.Debugf("ns/%s deployment/%s found container/image: %s/%v",
				namespace, d.Name, c.Name, c.Image)
			ContainerData := shared.Containers{
				Name:  c.Name,
				Image: c.Image,
			}
			s.ContainerList = append(s.ContainerList, ContainerData)
		}
		s = shared.Stack{
			Agent:     CfgClusterNickname,
			Name:      d.Name,
			Kind:      "deployment",
			Lane:      lane,
			Namespace: namespace,
			Replicas: shared.Replicas{
				Available: d.Status.AvailableReplicas,
				Ready:     d.Status.ReadyReplicas,
				Updated:   d.Status.ReadyReplicas,
				//Desired: d.Spec.Replicas,
			},
			ContainerList: s.ContainerList,
		}
		log.Debug(s)
		err := api.PostStack(apiServer, s, tokenString)
		if err != nil {
			// TODO: metrics!
			log.Error(err)
			return
		}
	}
	countDeploys := len(list.Items)
	log.Infof("lane/%s, ns/%s sent %d deployments", lane, namespace, countDeploys)

}


// GetValueFromLabelKey returns bool if if matched and the value as a string from a set of Labels
func GetValueFromLabelKey(inputList map[string]string, NamespaceLaneKey string) (bool, string) {
	log.Debugf("searching for %s in a namespace\n", NamespaceLaneKey)
	for k, v := range inputList {
		//log.Debugf("searching :::: [%s ---- %s ]\n", k, v)
		if k == NamespaceLaneKey {
			log.Debugf("ns/%s matched a label value: %v\n", v)
			return true, v
		}
	}
	return false, "unknown"
}