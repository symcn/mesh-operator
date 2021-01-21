package option

// ControllerOption ...
type ControllerOption struct {
	HTTPAddress             string
	MetricsEnabled          bool
	SyncPeriod              int32
	LeaderElectionNamespace string
	LeaderElectionID        string
	EnableLeaderElection    bool
	GinLogEnabled           bool
	GinLogSkipPath          []string
	PprofEnabled            bool
	GoroutineThreshold      int
	MaxConcurrentReconciles int
	InitDefaultSidecars     bool

	// Dubbo proxy settings
	ProxyHost          string
	ProxyAttempts      int32
	ProxyPerTryTimeout int64
	ProxyRetryOn       string

	// MeshConfig Name
	MeshConfigName      string
	MeshConfigNamespace string

	// Custome some labels
	SelectLabel string

	// Controller Enabled
	EnableAppMeshConfig     bool
	EnableConfiguredService bool
	EnableIstioConfig       bool
	EnableMeshConfig        bool
	EnableServiceAccessor   bool
	EnableServiceConfig     bool
	WatchIstioCRD           bool
}

// DefaultControllerOption ...
func DefaultControllerOption() *ControllerOption {
	return &ControllerOption{
		HTTPAddress:             ":8080",
		SyncPeriod:              120,
		MetricsEnabled:          true,
		GinLogEnabled:           true,
		GinLogSkipPath:          []string{"/ready", "/live"},
		EnableLeaderElection:    true,
		LeaderElectionID:        "meshach-lock",
		LeaderElectionNamespace: "sym-admin",
		PprofEnabled:            true,
		GoroutineThreshold:      10000,
		MaxConcurrentReconciles: 1000,
		ProxyHost:               "mosn.io.dubbo.proxy",
		ProxyAttempts:           3,
		ProxyPerTryTimeout:      2,
		ProxyRetryOn:            "gateway-error,connect-failure,refused-stream",
		MeshConfigName:          "sym-meshconfig",
		MeshConfigNamespace:     "sym-admin",
		SelectLabel:             "service",
		EnableAppMeshConfig:     false,
		EnableConfiguredService: false,
		EnableIstioConfig:       false,
		EnableMeshConfig:        false,
		EnableServiceAccessor:   false,
		EnableServiceConfig:     false,
		WatchIstioCRD:           true,
		InitDefaultSidecars:     false,
	}
}
