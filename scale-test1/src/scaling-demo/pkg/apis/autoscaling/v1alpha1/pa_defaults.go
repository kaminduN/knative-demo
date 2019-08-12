package v1alpha1

func (r *PodAutoscaler) SetDefaults() {
	r.Spec.SetDefaults()
	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}
}

func (rs *PodAutoscalerSpec) SetDefaults() {
}
