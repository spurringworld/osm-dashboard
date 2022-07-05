package osmcli

type OsmInstallSpec struct {
	MeshName string `json:"meshName"`

	Namespace string `json:"namespace"`

	EnforceSingleMesh bool `json:"enforceSingleMesh"`
}

func NewOsmInstallSpec() OsmInstallSpec {
	osmInstallSpec := OsmInstallSpec{}
	osmInstallSpec.MeshName = "osm"
	osmInstallSpec.Namespace = "osm-system"
	osmInstallSpec.EnforceSingleMesh = true
	return osmInstallSpec
}

type OsmUninstallSpec struct {
	MeshName string `json:"meshName"`

	Namespace string `json:"namespace"`
}

func NewOsmUninstallSpec() OsmUninstallSpec {
	osmUninstallSpec := OsmUninstallSpec{}
	osmUninstallSpec.MeshName = "osm"
	osmUninstallSpec.Namespace = "osm-system"
	return osmUninstallSpec
}
