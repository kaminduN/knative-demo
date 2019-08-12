package v1alpha1

func (r *PodAutoscaler) SetDefaults() {
	if r.Annotations == nil {
		r.Annotations = make(map[string]string)
	}
}

func (rs *PodAutoscalerSpec) SetDefaults() {
}
