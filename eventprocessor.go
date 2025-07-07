package main

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"
)

func PodAdded(obj interface{}) {
	key, _ := cache.MetaNamespaceKeyFunc(obj)
	fmt.Println("[POD ADD]", key)
	var pe Event
	pe.eventType = "podAdd"
	pe.objName = key
	ch1 <- pe
	fmt.Println("Sent event for processing", key)
}
func PodUpdated(oldObj, newObj interface{}) {
	key, _ := cache.MetaNamespaceKeyFunc(newObj)
	fmt.Println("[POD UPDATE]", key)
}
func PodDeleted(obj interface{}) {
	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	fmt.Println("[POD DELETE]", key)
	var pe Event
	pe.eventType = "podDelete"
	pe.objName = key
	ch1 <- pe
	fmt.Println("Sent event for processing", key)
}

func RsAdded(obj interface{}) {
	rs := obj.(*appsv1.ReplicaSet)
	fmt.Println("[RS ADD]", rs.Name)
	for k, v := range rs.Labels {
		if k == "pod-template-hash" {
			deploymentName := strings.Split(rs.Name, "-"+v)[0]
			key := rs.Namespace + "/" + deploymentName
			if _, ok := deploydetails[deploymentName]; !ok {

				deploydetails[key] = new(DeploymentLifetimeStat)
			}
			deploydetails[key].TotalRScount++
		}
	}
}
func RsUpdated(oldObj, newObj interface{}) {
	rs := newObj.(*appsv1.ReplicaSet)
	fmt.Println("[RS UPDATE]", rs.Name)
}
func RsDeleted(obj interface{}) {
	rs := obj.(*appsv1.ReplicaSet)
	fmt.Println("[RS DELETE]", rs.Name)
}
