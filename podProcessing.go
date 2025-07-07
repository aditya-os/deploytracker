package main

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func process_pod_event_add(event Event) {
	podname, namespace, err := get_pod_name_ns(event.objName)
	if err != nil {
		fmt.Printf("Invalid podname")
	}
	podobj, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podname, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(" PodName[%s]", podname)
	for k, v := range podobj.Labels {
		fmt.Println("  " + k + ":" + v)
		if k == "pod-template-hash" {
			deploymentName := strings.Split(podname, "-"+v+"-")[0]
			key := namespace + "/" + deploymentName
			fmt.Println("Deployment name : ", deploymentName)
			if _, ok := deploydetails[key]; !ok {
				deploydetails[key] = new(DeploymentLifetimeStat)
			}
			deploydetails[key].TotalpodCount++
			deploydetails[key].CurrPodCount++
		}
	}
	// for i, container := range podobj.Spec.Containers {
	// 	fmt.Printf("Container[%d]: %s | Image: %s\n", i, container.Name, container.Image)
	// }
}

func process_pod_event_delete(event Event) {
	podname, namespace, err := get_pod_name_ns(event.objName)
	if err != nil {
		fmt.Printf("Invalid podname")
	}
	fmt.Println(" PodName[", podname, "]")
	podinfo := strings.Split(podname, "-")
	if len(podinfo) <= 2 {
		return
	}
	podhash := podinfo[len(podinfo)-1]
	rshash := podinfo[len(podinfo)-2]
	deploymentName := strings.Split(podname, "-"+rshash+"-"+podhash)[0]
	key := namespace + "/" + deploymentName
	fmt.Println("deployment Name : ", deploymentName)
	if _, ok := deploydetails[key]; !ok {
		fmt.Println("No deployment found for the pod")
	}
	deploydetails[key].CurrPodCount--

}
