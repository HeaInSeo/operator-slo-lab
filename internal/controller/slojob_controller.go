/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	labv1alpha1 "github.com/HeaInSeo/operator-slo-lab/api/v1alpha1"
)

// SloJobReconciler reconciles a SloJob object
type SloJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=lab.heainseo.dev,resources=slojobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=lab.heainseo.dev,resources=slojobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=lab.heainseo.dev,resources=slojobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SloJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *SloJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("slojob", req.NamespacedName)

	// 1) Fetch object
	var slojob labv1alpha1.SloJob
	if err := r.Get(ctx, req.NamespacedName, &slojob); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 위의 코드와 동일, 학습을 위해 남겨둠.
	//if err := r.Get(ctx, req.NamespacedName, &slojob); err != nil {
	//	if apierrors.IsNotFound(err) {
	//		return ctrl.Result{}, nil
	//	}
	//	return ctrl.Result{}, err
	//}

	// 2) Read start-time annotation (Phase 2: more explicit logs)
	if slojob.Annotations == nil {
		log.Info("no annotations; skip convergence metric")
		return ctrl.Result{}, nil
	}

	//startStr := slojob.Annotations["test/start-time"]
	//if startStr == "" {
	//	log.Info("missing annotation test/start-time; skip convergence metric")
	//	return ctrl.Result{}, nil
	//}

	startStr, ok := slojob.Annotations["test/start-time"]
	if !ok {
		log.Info("annotation test/start-time not set; skip convergence metric")
		return ctrl.Result{}, nil
	}

	if startStr == "" {
		log.Info("annotation test/start-time is empty; skip convergence metric")
		return ctrl.Result{}, nil
	}

	startTime, err := time.Parse(time.RFC3339Nano, startStr)
	if err != nil {
		log.Error(err, "invalid annotation test/start-time; skip convergence metric", "value", startStr)
		return ctrl.Result{}, nil
	}

	// 3) Duplicate prevention:
	// If we already observed this exact start-time, skip observing again.
	if slojob.Status.ObservedStartTime != nil && *slojob.Status.ObservedStartTime == startStr {
		prev := ""
		if slojob.Status.ObservedResult != nil {
			prev = *slojob.Status.ObservedResult
		}
		log.Info("start-time already observed; skip convergence metric", "start-time", startStr, "previousResult", prev)
		return ctrl.Result{}, nil
	}

	dur := time.Since(startTime)
	seconds := dur.Seconds()            // // Prometheus Observe용 (float64)
	ms := int64(dur / time.Millisecond) // Status 저장용 (int64)

	// 4) Failure label + error path (test hook):
	// If annotation test/force-fail=="true", observe as error and return error.
	// 일단 주석 처리함.
	//resultLabel := "success"
	//if slojob.Annotations["test/force-fail"] == "true" {
	//	resultLabel = "error"
	//}

	// 아래로 교체
	resultLabel := "success"

	testHooksEnabled := os.Getenv("SLOLAB_TEST_HOOKS") == "1"
	forceFail := testHooksEnabled && slojob.Annotations["test/force-fail"] == "true"
	if forceFail {
		resultLabel = "error"
	}

	// --- 여기부터가 핵심: 먼저 status를 "내가 관측했다"로 기록 ---
	orig := slojob.DeepCopy()
	now := metav1.Now()

	slojob.Status.ObservedStartTime = &startStr
	slojob.Status.ObservedResult = &resultLabel
	slojob.Status.ObservedMillis = &ms
	slojob.Status.ObservedAt = &now

	patch := client.MergeFromWithOptions(orig, client.MergeFromWithOptimisticLock{})
	if err := r.Status().Patch(ctx, &slojob, patch); err != nil {
		if apierrors.IsConflict(err) {
			log.Info("status update conflict; likely already observed; skip metric", "start-time", startStr)
			// 캐시가 따라잡게 아주 짧게 requeue (선택)
			return ctrl.Result{RequeueAfter: 100 * time.Millisecond}, nil
		}
		return ctrl.Result{}, err
	}

	// --- status 기록이 성공한 경우에만 Observe ---
	E2EConvergenceSeconds.WithLabelValues(resultLabel).Observe(seconds)
	log.Info("observed e2e convergence time", "result", resultLabel, "seconds", seconds)

	// 주석 처리함
	//if resultLabel == "error" {
	//	return ctrl.Result{}, fmt.Errorf("forced failure for testing (test/force-fail=true)")
	//}

	if forceFail {
		// 1회성 훅: 다음 reconcile에서 계속 실패하지 않게 annotation 제거
		orig2 := slojob.DeepCopy()
		delete(slojob.Annotations, "test/force-fail")

		patch2 := client.MergeFromWithOptions(orig2, client.MergeFromWithOptimisticLock{})
		if err := r.Patch(ctx, &slojob, patch2); err != nil {
			// annotation 제거 실패는 테스트 훅이므로 로그만 남기고 실패는 그대로 반환
			log.Error(err, "failed to remove test/force-fail annotation")
		}

		return ctrl.Result{}, fmt.Errorf("forced failure for testing (SLOLAB_TEST_HOOKS=1)")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SloJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&labv1alpha1.SloJob{}).
		Named("slojob").
		Complete(r)
}
