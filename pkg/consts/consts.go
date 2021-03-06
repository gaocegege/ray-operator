package consts

const (
	EventNormal  = "Normal"
	EventWarning = "Warning"

	ReasonValidationFailed = "ValidationFailedOrNotImplemented"
	ReasonCreate           = "SuccessfullyCreate"
	ReasonUpdate           = "SuccessfullyUpdate"

	LabelRayWorker   = "ray-worker"
	LabelRayHead     = "ray-head"
	LabelRayLauncher = "ray-launcher"
	LabelRay         = "ray"

	DefaultRayImage              = "rayproject/autoscaler"
	DefaultRayLauncherName       = "ray-launcher"
	DefaultRayConfigMapFile      = "config.yaml"
	DefaultRayConfigMapVolume    = "ray-config-volume"
	DefaultRayConfigMapMountPath = "/etc/config/"
	DefaultRayConfigMapFilePath  = DefaultRayConfigMapMountPath +
		DefaultRayConfigMapFile
)
