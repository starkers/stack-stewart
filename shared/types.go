package shared

// // Common types to be shared across the project

// Containers ..
type Containers struct {
	Image string `json:"image" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

// Replicas ..
type Replicas struct {
	Available int `json:"available" validate:"required"`
	Desired   int `json:"desired" validate:"required"`
	Ready     int `json:"ready" validate:"required"`
	Updated   int `json:"updated" validate:"required"`
}

// Stack ..
type Stack struct {
	Agent         string       `json:"agent" validate:"required"`
	ContainerList []Containers `json:"containers" validate:"required"`
	Kind          string       `json:"kind" validate:"required"`
	Lane          string       `json:"lane" validate:"required"`
	Name          string       `json:"name" validate:"required"`
	Namespace     string       `json:"namespace" validate:"required"`
	Replicas      Replicas     `json:"replicas"`
}

// StackList  list of stacks
type StackList struct {
	Stacks []Stack `json:"stacks"`
}

// ServerConfig to be created from config.yaml
type ServerConfig struct {
	Agents []struct {
		Name  string `yaml:"name"`
		Token string `yaml:"token"`
	} `yaml:"agents"`
}
