package controller

import (
	"context"
	"time"

	apiv1 "k8s.io/api/core/v1"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	"github.com/argoproj/argo-workflows/v3/workflow/controller/indexes"
)

func (woc *wfOperationCtx) getPodGCDelay(ctx context.Context, podGC *wfv1.PodGC) time.Duration {
	delay := woc.controller.Config.GetPodGCDeleteDelayDuration()
	podGCDelay, err := podGC.GetDeleteDelayDuration()
	if err != nil {
		woc.log.WithError(err).Warn(ctx, "failed to parse podGC.deleteDelayDuration")
	} else if podGCDelay >= 0 {
		delay = podGCDelay
	}
	return delay
}

func (woc *wfOperationCtx) queuePodsForCleanup(ctx context.Context) {
	podGC := woc.execWf.Spec.PodGC
	delay := woc.getPodGCDelay(ctx, podGC)
	strategy := podGC.GetStrategy()
	selector, _ := podGC.GetLabelSelector()
	workflowPhase := woc.wf.Status.Phase
	objs, _ := woc.controller.PodController.GetPodsByIndex(indexes.WorkflowIndex, woc.wf.Namespace+"/"+woc.wf.Name)
	for _, obj := range objs {
		pod := obj.(*apiv1.Pod)
		if _, ok := pod.Labels[common.LabelKeyComponent]; ok { // for these types we don't want to do PodGC
			continue
		}
		nodeID := woc.nodeID(pod)
		node, err := woc.wf.Status.Nodes.Get(nodeID)
		if err != nil {
			woc.log.WithField("nodeID", nodeID).Error(ctx, "was unable to obtain node for nodeID")
			continue
		}
		nodePhase := node.Phase
		if !nodePhase.Fulfilled(node.TaskResultSynced) {
			continue
		}
		woc.controller.PodController.EnactAnyPodCleanup(ctx, selector, pod, strategy, workflowPhase, delay)
	}
}
