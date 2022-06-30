package traffictarget

import (
	"log"

	smiaccessv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha3"
	smiaccessclientset "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/access/clientset/versioned"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

// TrafficTarget is a representation of a traffictarget.
type TrafficTarget struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// TrafficTargetList contains a list of services in the cluster.
type TrafficTargetList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of traffictargets.
	TrafficTargets []TrafficTarget `json:"traffictargets"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetServiceList returns a list of all services in the cluster.
func GetTrafficTargetList(smiAccessClient smiaccessclientset.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*TrafficTargetList, error) {
	log.Print("Getting list of all traffictargets in the cluster")

	channels := &common.ResourceChannels{
		TrafficTargetList: common.GetTrafficTargetListChannel(smiAccessClient, nsQuery, 1),
	}

	return GetTrafficTargetListFromChannels(channels, dsQuery)
}

// GetTrafficTargetListFromChannels returns a list of all services in the cluster.
func GetTrafficTargetListFromChannels(channels *common.ResourceChannels,
	dsQuery *dataselect.DataSelectQuery) (*TrafficTargetList, error) {
	trafficTargets := <-channels.TrafficTargetList.List
	err := <-channels.TrafficTargetList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return CreateTrafficTargetList(trafficTargets.Items, nonCriticalErrors, dsQuery), nil
}

func toTrafficTarget(trafficTarget *smiaccessv1alpha3.TrafficTarget) TrafficTarget {
	return TrafficTarget{
		ObjectMeta:        api.NewObjectMeta(trafficTarget.ObjectMeta),
		TypeMeta:          api.NewTypeMeta(api.ResourceKindTrafficTarget),
	}
}

// CreateTrafficTargetList returns paginated traffictarget list based on given traffictarget array and pagination query.
func CreateTrafficTargetList(trafficTargets []smiaccessv1alpha3.TrafficTarget, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *TrafficTargetList {
	trafficTargetsList := &TrafficTargetList{
		TrafficTargets: make([]TrafficTarget, 0),
		ListMeta: api.ListMeta{TotalItems: len(trafficTargets)},
		Errors:   nonCriticalErrors,
	}

	trafficTargetCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(trafficTargets), dsQuery)
	trafficTargets = fromCells(trafficTargetCells)
	trafficTargetsList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, trafficTarget := range trafficTargets {
		trafficTargetsList.TrafficTargets = append(trafficTargetsList.TrafficTargets, toTrafficTarget(&trafficTarget))
	}

	return trafficTargetsList
}
