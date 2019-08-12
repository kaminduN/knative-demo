package v1alpha1

import (
	"strconv"
	"time"

	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scaling-demo/pkg/apis/autoscaling"

)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodAutoscaler is a Knative abstraction that encapsulates the interface by which Knative
// components instantiate autoscalers.  This definition is an abstraction that may be backed
// by multiple definitions.  For more information, see the Knative Pluggability presentation:
// https://docs.google.com/presentation/d/10KWynvAJYuOEWy69VBa6bHJVCqIsz1TNdEKosNvcpPY/edit
type PodAutoscaler struct {
	meta_v1.TypeMeta `json:",inline"`
	// +optional
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the PodAutoscaler (from the client).
	// +optional
	Spec PodAutoscalerSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the PodAutoscaler (from the controller).
	// +optional
	Status PodAutoscalerStatus `json:"status,omitempty"`
}

// Check that PodAutoscaler can be validated, can be defaulted, and has immutable fields.
var _ apis.Validatable = (*PodAutoscaler)(nil)
var _ apis.Defaultable = (*PodAutoscaler)(nil)
var _ apis.Immutable = (*PodAutoscaler)(nil)

// Check that ConfigurationStatus may have its conditions managed.
var _ duckv1alpha1.ConditionsAccessor = (*PodAutoscalerStatus)(nil)

const (
	// PodAutoscalerConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	PodAutoscalerConditionReady = duckv1alpha1.ConditionReady
	// PodAutoscalerConditionActive is set when the PodAutoscaler's ScaleTargetRef is receiving traffic.
	PodAutoscalerConditionActive duckv1alpha1.ConditionType = "Active"
)

var podCondSet = duckv1alpha1.NewLivingConditionSet(PodAutoscalerConditionActive)

// PodAutoscalerStatus communicates the observed state of the PodAutoscaler (from the controller).
type PodAutoscalerStatus struct {
	// Conditions communicates information about ongoing/complete
	// reconciliation processes that bring the "spec" inline with the observed
	// state of the world.
	// +optional
	Conditions duckv1alpha1.Conditions `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodAutoscalerList is a list of PodAutoscaler resources
type PodAutoscalerList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []PodAutoscaler `json:"items"`
}

func (pa *PodAutoscaler) Class() string {
	if c, ok := pa.Annotations[autoscaling.ClassAnnotationKey]; ok {
		return c
	}
	// Default to "pa" class for backward compatibility.
	return autoscaling.KPA
}

func (pa *PodAutoscaler) scaleBoundInt32(key string) int32 {
	if s, ok := pa.Annotations[key]; ok {
		// no error check: relying on validation
		i, _ := strconv.ParseInt(s, 10, 32)
		return int32(i)
	}
	return 0
}

// ScaleBounds returns scale bounds annotations values as a tuple:
// `(min, max int32)`. The value of 0 for any of min or max means the bound is
// not set
func (pa *PodAutoscaler) ScaleBounds() (min, max int32) {
	min = pa.scaleBoundInt32(autoscaling.MinScaleAnnotationKey)
	max = pa.scaleBoundInt32(autoscaling.MaxScaleAnnotationKey)
	return
}

func (pa *PodAutoscaler) MetricTarget() (target int32, ok bool) {
	if s, ok := pa.Annotations[autoscaling.TargetAnnotationKey]; ok {
		if i, err := strconv.Atoi(s); err == nil {
			return int32(i), true
		}
	}
	return 0, false
}

// IsReady looks at the conditions and if the Status has a condition
// PodAutoscalerConditionReady returns true if ConditionStatus is True
func (rs *PodAutoscalerStatus) IsReady() bool {
	return podCondSet.Manage(rs).IsHappy()
}

// IsActivating assumes the pod autoscaler is Activating if it is neither
// Active nor Inactive
func (rs *PodAutoscalerStatus) IsActivating() bool {
	cond := rs.GetCondition(PodAutoscalerConditionActive)

	return cond != nil && cond.Status == corev1.ConditionUnknown
}

func (rs *PodAutoscalerStatus) GetCondition(t duckv1alpha1.ConditionType) *duckv1alpha1.Condition {
	return podCondSet.Manage(rs).GetCondition(t)
}

func (rs *PodAutoscalerStatus) InitializeConditions() {
	podCondSet.Manage(rs).InitializeConditions()
}

func (rs *PodAutoscalerStatus) MarkActive() {
	podCondSet.Manage(rs).MarkTrue(PodAutoscalerConditionActive)
}

func (rs *PodAutoscalerStatus) MarkActivating(reason, message string) {
	podCondSet.Manage(rs).MarkUnknown(PodAutoscalerConditionActive, reason, message)
}

func (rs *PodAutoscalerStatus) MarkInactive(reason, message string) {
	podCondSet.Manage(rs).MarkFalse(PodAutoscalerConditionActive, reason, message)
}

// CanScaleToZero checks whether the pod autoscaler has been in an inactive state
// for at least the specified grace period.
func (rs *PodAutoscalerStatus) CanScaleToZero(gracePeriod time.Duration) bool {
	if cond := rs.GetCondition(PodAutoscalerConditionActive); cond != nil {
		switch cond.Status {
		case corev1.ConditionFalse:
			// Check that this PodAutoscaler has been inactive for
			// at least the grace period.
			return time.Now().After(cond.LastTransitionTime.Inner.Add(gracePeriod))
		}
	}
	return false
}

// CanMarkInactive checks whether the pod autoscaler has been in an active state
// for at least the specified idle period.
func (rs *PodAutoscalerStatus) CanMarkInactive(idlePeriod time.Duration) bool {
	if cond := rs.GetCondition(PodAutoscalerConditionActive); cond != nil {
		switch cond.Status {
		case corev1.ConditionTrue:
			// Check that this PodAutoscaler has been active for
			// at least the grace period.
			return time.Now().After(cond.LastTransitionTime.Inner.Add(idlePeriod))
		}
	}
	return false
}

// GetConditions returns the Conditions array. This enables generic handling of
// conditions by implementing the duckv1alpha1.Conditions interface.
func (rs *PodAutoscalerStatus) GetConditions() duckv1alpha1.Conditions {
	return rs.Conditions
}

// SetConditions sets the Conditions array. This enables generic handling of
// conditions by implementing the duckv1alpha1.Conditions interface.
func (rs *PodAutoscalerStatus) SetConditions(conditions duckv1alpha1.Conditions) {
	rs.Conditions = conditions
}

// PodAutoscalerSpec is the spec for a PodAutoscaler resource
type PodAutoscalerSpec struct {
	// Message and SomeValue are example custom spec fields
	//
	// this is where you would put your custom resource data
	Message   string `json:"message"`
	SomeValue *int32 `json:"someValue"`
}
