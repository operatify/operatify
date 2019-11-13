package v1alpha1

// Status defines the desired state of resource
type Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Id         string `json:"id,omitempty"`
	StringData string `json:"stringData,omitempty"`
	IntData    int    `json:"intData,omitempty"`
}

// Status defines the observed state of resource
type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
}
