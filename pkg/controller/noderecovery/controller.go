/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package noderecovery

import (
	"time"

	"github.com/golang/glog"

	"kubevirt.io/node-recovery/pkg/controller"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// maxRetries is the number of times a deployment will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a deployment is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

type NodeRecoveryController struct {
	queue             workqueue.RateLimitingInterface
	nodeInformer      cache.SharedIndexInformer
	configMapInformer cache.SharedIndexInformer
	jobInformer       cache.SharedIndexInformer
}

// NewNodeRecoveryController returns new NodeRecoveryController instance
func NewNodeRecoveryController(
	nodeInformer cache.SharedIndexInformer,
	configMapInformer cache.SharedIndexInformer,
	jobInformer cache.SharedIndexInformer) *NodeRecoveryController {

	c := &NodeRecoveryController{
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nodeInformer:      nodeInformer,
		configMapInformer: configMapInformer,
		jobInformer:       jobInformer,
	}

	c.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addNode,
		DeleteFunc: c.deleteNode,
		UpdateFunc: c.updateNode,
	})

	c.configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addConfigMap,
		DeleteFunc: c.deleteConfigMap,
		UpdateFunc: c.updateConfigMap,
	})

	c.jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addJob,
		DeleteFunc: c.deleteJob,
		UpdateFunc: c.updateJob,
	})

	return c
}

// TODO: I think we will need only part of this handlers,
// addNode, deleteNode, updateNode and maybe 
func (c *NodeRecoveryController) addNode(obj interface{}) {
}

func (c *NodeRecoveryController) deleteNode(obj interface{}) {
}

func (c *NodeRecoveryController) updateNode(old, curr interface{}) {
	currNode := curr.(*apiv1.Node)
	glog.Warningf("Node %s updated", currNode.Name)
}

func (c *NodeRecoveryController) addConfigMap(obj interface{}) {
}

func (c *NodeRecoveryController) deleteConfigMap(obj interface{}) {
}

func (c *NodeRecoveryController) updateConfigMap(old, curr interface{}) {
}

func (c *NodeRecoveryController) addJob(obj interface{}) {
}

func (c *NodeRecoveryController) deleteJob(obj interface{}) {
}

func (c *NodeRecoveryController) updateJob(old, curr interface{}) {
}

// Run begins watching and syncing.
func (c *NodeRecoveryController) Run(threadiness int, stopCh chan struct{}) {
	defer controller.HandlePanic()
	defer c.queue.ShutDown()
	glog.Info("Starting node-recovery controller.")

	// Wait for cache sync before we start the pod controller
	if !controller.WaitForCacheSync("node-recovery", stopCh, c.nodeInformer.HasSynced, c.configMapInformer.HasSynced, c.jobInformer.HasSynced) {
		return
	}

	// Start the actual work
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping node-recovery controller.")
}

func (c *NodeRecoveryController) worker() {
	for c.processNextWorkItem() {
	}
}

// Execute runs a worker thread that just dequeues items
func (c *NodeRecoveryController) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)
	err := c.syncNode(key.(string))

	c.handleErr(err, key)
	return true
}

func (c *NodeRecoveryController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < maxRetries {
		glog.V(2).Infof("Error syncing node %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	glog.V(2).Infof("Dropping node %q out of the queue: %v", key, err)
	c.queue.Forget(key)
}

func (c *NodeRecoveryController) syncNode(key string) error {
	// TODO: add remediation logic
	return nil
}