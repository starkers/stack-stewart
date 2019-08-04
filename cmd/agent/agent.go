package main

import (
	"fmt"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	CfgKubeConfig = kingpin.Flag(
		"kubeconfig", "Full path to your KUBECONFIG file").
		Default("nope").
		Envar("KUBECONFIG").
		Short('k').String()
	CfgAgentSecretName = kingpin.Flag(
		"secret-name", "secret for agent").
		Default("fudge").
		Envar("SECRET_NAME").
		Short('s').String()
	CfgLogLevel = kingpin.Flag(
		"log-level", "Debug, Info").
		Default("Info").
		Envar("LOG_LEVEL").
		Short('v').String()
	// filter namespaces for this
	NamespaceFilter = "fudge=yes"
	// CfgAPIServer ..
	CfgAPIServer = kingpin.Flag(
		"api-server", "Url of the api server.. eg 'https://api.example.com:3443'").
		Default("http://localhost:8080/stacks").
		Envar("API_SERVER").
		Short('a').String()
)

func main() {
	kingpin.New(filepath.Base(os.Args[0]), "Stack Stewart (agent)")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// start up log.. TODO: param for log level/json-text?
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if *CfgLogLevel == "Debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// connect to kubernetes and retrieve a client-go "clientset"
	// TODO: also return err?
	cs := GetClientSet(*CfgKubeConfig)

	// determines the namespace (string).. either from pod ENV vars or $NAMESPACE
	// falls back on "default" if unable to determine.
	namespace := GetMyNamespace()
	log.Debugf("detected namespace: %s", namespace)

	// creates a UUID (in a k8s secret) if not already there
	err := SecretKeyBootstrap(cs, namespace, *CfgAgentSecretName, "token")
	if err != nil {
		panic(err.Error())
	}

	secretTokenString, err := GetSecretKeyData(cs, namespace, *CfgAgentSecretName, "token")
	if err != nil {
		panic(err.Error())
	}
	log.Infof("secretToken: %s", secretTokenString)

	// we will only look for deployments inside a namespace matching these labels
	listOpts := metav1.ListOptions{LabelSelector: NamespaceFilter}

	for {
		// make a list of namespaces to parse for deployments
		list, err := cs.CoreV1().Namespaces().List(listOpts)
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(list.Items) == 0 {
			log.Warnf("matched 0 namespaces with %s", NamespaceFilter)
		} else {
			// loop over namespaces
			for _, i := range list.Items {
				// convert the namespace Name into a string
				namespaceString := fmt.Sprintf(i.Name)
				log.Debugf("processing namespace: %s", namespaceString)
				SendDeployments(
					cs,
					namespaceString,
					secretTokenString,
					*CfgAPIServer,
				)
			}
		}
		time.Sleep(30 * time.Second)
	}

}
