package shared

// // Common types to be shared across the project

// Containers ..
type Containers struct {
	Image string `json:"image" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Tag   string `json:"tag"` //only server needs this
}

// Replicas ..
type Replicas struct {
	Available int32 `json:"available" validate:"required"`
	//Desired   int32 `json:"desired" validate:"required"`
	Ready   int32 `json:"ready" validate:"required"`
	Updated int32 `json:"updated" validate:"required"`
}

// Stack ..
type Stack struct {
	Agent         string       `json:"agent"` //not required on the agent, the server will use it however
	ContainerList []Containers `json:"containers" validate:"required"`
	Kind          string       `json:"kind" validate:"required"`
	Lane          string       `json:"lane" validate:"required"`
	Name          string       `json:"name" validate:"required"`
	Namespace     string       `json:"namespace" validate:"required"`
	Replicas      Replicas     `json:"replicas"`
	Trace         string       `json:"trace"` //not required.. only the server uses this
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
	LogLevel string `yaml:"log_level"`
}
