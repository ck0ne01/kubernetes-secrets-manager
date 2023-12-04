package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	secretsMsg         *v1.SecretList
	namespacesItemsMsg []list.Item
	secretData         map[string]string
)

type secret struct {
	ApiVersion string
	Data       secretData
	Kind       string
	Metadata   metadata
}

type metadata struct {
	Name string
}

func fetchNamespaces() (*v1.NamespaceList, error) {
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// Use the kubeconfig files to create a Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/christian.kirmse/.kube/config_go")
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return namespaces, nil
}

// TODO: use generic to refactor to one xToListItems function
func namespacesToListItems(namespaces *v1.NamespaceList) namespacesItemsMsg {
	var items []list.Item

	for _, namespace := range namespaces.Items {
		items = append(items, item(namespace.Name))
	}

	return namespacesItemsMsg(items)
}

func fetchSecrets(namespace string) (*v1.SecretList, error) {
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// Use the kubeconfig files to create a Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/christian.kirmse/.kube/config_go")
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	return secretsMsg(secrets), nil
}

func secretNamesToListItems(secrets *v1.SecretList) []list.Item {
	var items []list.Item

	for _, secret := range secrets.Items {
		items = append(items, item(secret.Name))
	}

	return items
}

func stringToSecretData(secretData *string) secretData {
	orderedSecretData := make(map[string]string)
	secretDataLines := strings.Split(strings.TrimSpace(*secretData), "\n")
	for _, secretDataLine := range secretDataLines {
		secretDataLineElements := strings.Split(secretDataLine, ":")
		key := secretDataLineElements[0]
		value := strings.TrimSpace(secretDataLineElements[1])
		encodedValue := b64.StdEncoding.EncodeToString([]byte(value))
		orderedSecretData[key] = encodedValue
	}
	return orderedSecretData
}

func getSecretData(secrets *v1.SecretList, secretName string) string {
	var secretData string
	for _, secret := range secrets.Items {
		if secret.Name == secretName {
			for key, value := range secret.Data {
				secretData += fmt.Sprintf("%s: %s\n", key, string(value))
			}
		}
	}
	return secretData
}

func createSecretYamlContent(secretName string, secretData secretData) ([]byte, error) {
	newSecretContent := secret{
		ApiVersion: "v1",
		Kind:       "Secret",
		Data:       secretData,
		Metadata: metadata{
			Name: secretName,
		},
	}

	return yaml.Marshal(&newSecretContent)
}
