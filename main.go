package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func get_config() (*rest.Config, error) {
	var config *rest.Config
	var err error
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		fmt.Println("Using in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		fmt.Println("Using local kubeconfig")
		kubeconfig := clientcmd.RecommendedHomeFile
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return config, err
}

type DeploymentLifetimeStat struct {
	TotalpodCount int
	CurrPodCount  int
	TotalRScount  int
}

type Event struct {
	objName   string
	eventType string
}

var deploydetails map[string]*DeploymentLifetimeStat

// func PodAdded(obj interface{}) {
// 	key, _ := cache.MetaNamespaceKeyFunc(obj)
// 	fmt.Println("[POD ADD]", key)
// 	var pe Event
// 	pe.eventType = "podAdd"
// 	pe.objName = key
// 	ch1 <- pe
// 	fmt.Println("Sent event for processing", key)
// }
// func PodUpdated(oldObj, newObj interface{}) {
// 	key, _ := cache.MetaNamespaceKeyFunc(newObj)
// 	fmt.Println("[POD UPDATE]", key)
// }
// func PodDeleted(obj interface{}) {
// 	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
// 	fmt.Println("[POD DELETE]", key)
// 	var pe Event
// 	pe.eventType = "podDelete"
// 	pe.objName = key
// 	ch1 <- pe
// 	fmt.Println("Sent event for processing", key)
// }

var clientset *kubernetes.Clientset
var ch1 chan Event

func main() {
	var config *rest.Config
	var err error

	config, err = get_config()
	if err != nil {
		panic(err)
	}

	deploydetails = make(map[string]*DeploymentLifetimeStat)

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	fmt.Println("Clientset DONE")

	ch1 = make(chan Event, 100)
	go process_events(ch1)
	defer close(ch1)

	factory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    PodAdded,
		UpdateFunc: PodUpdated,
		DeleteFunc: PodDeleted,
	})
	fmt.Println("Added POD handlers")

	rsInformer := factory.Apps().V1().ReplicaSets().Informer()
	rsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    RsAdded,
		UpdateFunc: RsUpdated,
		DeleteFunc: RsDeleted,
	})

	stopCh := make(chan struct{})
	defer close(stopCh)
	go factory.Start(stopCh)
	fmt.Println("Starting Sync")
	if !cache.WaitForCacheSync(stopCh, podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	fmt.Println("SyncDone")

	go serveclient()

	wait.Until(func() {
		fmt.Println("Main loop")
	}, time.Second*10, stopCh)

}

func deploystats(w http.ResponseWriter, r *http.Request) {
	jsonString, _ := json.Marshal(deploydetails)
	fmt.Fprintf(w, string(jsonString))
}

func serveclient() {
	http.HandleFunc("/deploytracker", deploystats)
	http.ListenAndServe(":8080", nil)
}

func process_rs_event_add(event Event) {

}

func process_rs_event_delete(event Event) {

}

// Receiver goroutine: continuously waiting for input
func process_events(ch1 chan Event) {
	for {
		select {
		case event := <-(ch1):
			fmt.Printf("Received event %s and objName %s\n", event.eventType, event.objName)
			// Handle data from ch1
			switch event.eventType {
			case "podAdd":
				{
					process_pod_event_add(event)
				}
			case "podDelete":
				{
					process_pod_event_delete(event)
				}
			case "podUpdate":
				{

				}

			case "rsAdd":
				{
					process_rs_event_add(event)
				}
			case "rsDelete":
				{
					process_rs_event_delete(event)
				}
			case "rsUpdate":
				{

				}

			}

		default:
		}
	}
}

func get_pod_name_ns(podname string) (string, string, error) {
	poddetail := strings.Split(podname, "/")
	if len(poddetail) < 1 {
		return "", "", fmt.Errorf("Invalid")
	}
	//fmt.Println("split pod details", poddetail[1])
	podinfo := poddetail[1]
	namespace := poddetail[0]
	return podinfo, namespace, nil
}
