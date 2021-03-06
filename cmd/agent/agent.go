package main

import (
	"fmt"
	"strconv"
	"time"

	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
)

var (
	CfgKubeConfig = kingpin.Flag(
		"kubeconfig", "(optional) full path to your KUBECONFIG file").
		Default("nope").
		Envar("KUBECONFIG").
		Short('k').String()
	CfgAgentSecretName = kingpin.Flag(
		"secret-name", "(optional) secret for agent").
		Default("stack-stewart").
		Envar("SECRET_NAME").
		String()
	CfgLogLevel = kingpin.Flag(
		"log-level", "Debug, Info").
		Default("Info").
		Envar("LOG_LEVEL").
		Short('v').String()
	// NamespaceLaneKey tells the agent what the "lane" is
	NamespaceLaneKey = "lane"

	CfgAgentName = kingpin.Flag(
		"agent-nickname", "nickname of the agent.. I recommend a short human-readable name for the cluster").
		Default("unconfigured-agent-name").
		Envar("AGENT_NAME").
		Short('a').String()

	// CfgServerAddress ..
	CfgServerAddress = kingpin.Flag(
		"api-server", "Url of the central API server.. eg 'https://server.example.com/stacks'").
		Default("http://localhost:8080/stacks").
		Envar("SERVER_ADDRESS").
		Short('s').String()

	CfgTick = kingpin.Flag(
		"tick", "how often to 'tick' over and send metrics").
		Default("15").
		Envar("TICK").
		Short('t').String()
)

func main() {
	kingpin.New(filepath.Base(os.Args[0]), "Stack Stewart (agent)")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// start up log.. TODO: param for json/text?
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{
		//DisableColors: true,
		FullTimestamp: true,
	})
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

	// convert tick time to int64
	tickTime, err := strconv.ParseInt(*CfgTick, 10, 64)
	log.Infof("will send metrics every %d seconds", tickTime)
	duration := time.Duration(tickTime) * time.Second
	tick := time.Tick(duration)
	for {
		select {
		case <-tick:
			list, err := cs.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				log.Fatal(err)
			}
			if len(list.Items) == 0 {
				log.Warn("ohh no.. we found 0 namespaces?")
			}
			log.Debugf("found %d namespaces", len(list.Items))

			// Loop over all namespaces
			for _, ns := range list.Items {

				namespaceString := fmt.Sprintf(ns.Name)
				// Match namespaces with the lane label
				matched, lane := GetValueFromLabelKey(ns.Labels, NamespaceLaneKey)
				if matched {
					log.Debugf("processing namespace: %s", namespaceString)
					SendDeployments(
						cs,
						namespaceString,
						lane,
						secretTokenString,
						*CfgServerAddress,
						*CfgAgentName,
					)
				}

			}

		}

	}

}
