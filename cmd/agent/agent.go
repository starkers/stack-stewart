package main

import (
	"fmt"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	cfgKubeConfig = kingpin.Flag(
		"kubeconfig", "Full path to your KUBECONFIG file").
		Default("nope").
		Envar("KUBECONFIG").
		Short('k').String()
	cfgAgentSecretName = kingpin.Flag(
		"secret-name", "secret for agent").
		Default("fudge").
		Envar("SECRET_NAME").
		Short('s').String()
	cfgLogLevel = kingpin.Flag(
		"log-level", "Debug, Info").
		Default("Info").
		Envar("LOG_LEVEL").
		Short('v').String()
	// filter namespaces for this
	namespaceFilter = "fudge=yes"
	CfgAPIServer    = kingpin.Flag(
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
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	if *cfgLogLevel == "Debug" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	log := logrus.WithFields(logrus.Fields{
		//"common": "common field",
	})

	// connect to kubernetes and retrieve a client-go "clientset"
	// TODO: also return err?
	cs := GetClientSet(log, *cfgKubeConfig)

	// determines the namespace (string).. either from pod ENV vars or $NAMESPACE
	// falls back on "default" if unable to determine.
	namespace := GetMyNamespace()
	log.Debugf("detected namespace: %s", namespace)

	// creates a UUID (in a k8s secret) if not already there
	err := SecretKeyBootstrap(cs, namespace, *cfgAgentSecretName, "token", log)
	if err != nil {
		panic(err.Error())
	}

	secretTokenString, err := GetSecretKeyData(cs, namespace, *cfgAgentSecretName, "token", log)
	if err != nil {
		panic(err.Error())
	}
	log.Infof("secretToken: %s", secretTokenString)

	// we will only look for deployments inside a namespace matching these labels
	listOpts := metav1.ListOptions{LabelSelector: namespaceFilter}

	for {
		// make a list of namespaces to parse for deployments
		list, err := cs.CoreV1().Namespaces().List(listOpts)
		if err != nil {
			log.Fatal(err.Error())
		}
		if len(list.Items) == 0 {
			log.Warnf("matched 0 namespaces with %s", namespaceFilter)
		} else {
			// loop over namespaces
			for _, i := range list.Items {
				// convert the namespace Name into a string
				namespaceString := fmt.Sprintf(i.Name)
				log.Debugf("processing namespace: %s", namespaceString)
				SentDeployments(
					cs,
					namespaceString,
					log,
					secretTokenString,
					*CfgAPIServer,
				)
			}
		}
		time.Sleep(30 * time.Second)
	}

}
