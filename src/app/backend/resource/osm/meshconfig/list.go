package meshconfig

import (
	"log"

	osmconfigv1alph2 "github.com/openservicemesh/osm/pkg/apis/config/v1alpha2"
	osmconfigclientset "github.com/openservicemesh/osm/pkg/gen/client/config/clientset/versioned"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

// MeshConfig is a representation of a meshconfig.
type MeshConfig struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
	// Spec is the MeshConfig specification.
	// +optional
	Spec osmconfigv1alph2.MeshConfigSpec `json:"spec,omitempty"`
}

// MeshConfigList contains a list of services in the cluster.
type MeshConfigList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of meshconfigs.
	MeshConfigs []MeshConfig `json:"meshconfigs"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetServiceList returns a list of all services in the cluster.
func GetMeshConfigList(osmConfigClient osmconfigclientset.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*MeshConfigList, error) {
	log.Print("Getting list of all meshconfigs in the cluster")

	channels := &common.ResourceChannels{
		MeshConfigList: common.GetMeshConfigListChannel(osmConfigClient, nsQuery, 1),
	}

	return GetMeshConfigListFromChannels(channels, dsQuery)
}

// GetMeshConfigListFromChannels returns a list of all services in the cluster.
func GetMeshConfigListFromChannels(channels *common.ResourceChannels,
	dsQuery *dataselect.DataSelectQuery) (*MeshConfigList, error) {
	meshConfigs := <-channels.MeshConfigList.List
	err := <-channels.MeshConfigList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return CreateMeshConfigList(meshConfigs.Items, nonCriticalErrors, dsQuery), nil
}

func toMeshConfig(meshConfig *osmconfigv1alph2.MeshConfig) MeshConfig {
	return MeshConfig{
		ObjectMeta: api.NewObjectMeta(meshConfig.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindService),
		Spec:       meshConfig.Spec,
	}
}

// CreateMeshConfigList returns paginated traffictarget list based on given traffictarget array and pagination query.
func CreateMeshConfigList(meshConfigs []osmconfigv1alph2.MeshConfig, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *MeshConfigList {
	meshConfigsList := &MeshConfigList{
		MeshConfigs: make([]MeshConfig, 0),
		ListMeta:    api.ListMeta{TotalItems: len(meshConfigs)},
		Errors:      nonCriticalErrors,
	}

	meshConfigCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(meshConfigs), dsQuery)
	meshConfigs = fromCells(meshConfigCells)
	meshConfigsList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, meshConfig := range meshConfigs {
		meshConfigsList.MeshConfigs = append(meshConfigsList.MeshConfigs, toMeshConfig(&meshConfig))
	}

	return meshConfigsList
}
