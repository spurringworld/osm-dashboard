package trafficsplit

import (
	"log"

	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	smisplitclientset "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/split/clientset/versioned"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

// TrafficSplit is a representation of a httpgroup.
type TrafficSplit struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// TrafficSplitList contains a list of trafic split in the cluster.
type TrafficSplitList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of trafficsplits.
	TrafficSplits []TrafficSplit `json:"trafficsplits"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetTrafficSplitList returns a list of all traffice split in the cluster.
func GetTrafficSplitList(smiSplitClient smisplitclientset.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*TrafficSplitList, error) {
	log.Print("=== === === >>> Getting list of all traffic slip in the cluster")

	channels := &common.ResourceChannels{
		TrafficSplitList: common.GetTrafficSplitListChannel(smiSplitClient, nsQuery, 1),
	}

	return GetTrafficSplitListFromChannels(channels, dsQuery)
}

// GetTrafficSplitListFromChannels returns a list of all traffic split in the cluster.
func GetTrafficSplitListFromChannels(channels *common.ResourceChannels,
	dsQuery *dataselect.DataSelectQuery) (*TrafficSplitList, error) {
	trafficSplits := <-channels.TrafficSplitList.List
	err := <-channels.TrafficSplitList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		println(criticalError)
		return nil, criticalError
	}

	return CreateTrafficSplitList(trafficSplits.Items, nonCriticalErrors, dsQuery), nil
}

func toTrafficSplit(trafficSplit *smisplitv1alpha2.TrafficSplit) TrafficSplit {
	return TrafficSplit{
		ObjectMeta: api.NewObjectMeta(trafficSplit.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindTrafficSplit),
	}
}

// CreateTrafficSplitList returns paginated traffic split list based on given traffic split array and pagination query.
func CreateTrafficSplitList(trafficSplits []smisplitv1alpha2.TrafficSplit, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *TrafficSplitList {
	trafficSplitsList := &TrafficSplitList{
		TrafficSplits: make([]TrafficSplit, 0),
		ListMeta:      api.ListMeta{TotalItems: len(trafficSplits)},
		Errors:        nonCriticalErrors,
	}

	trafficSplitCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(trafficSplits), dsQuery)
	trafficSplits = fromCells(trafficSplitCells)
	trafficSplitsList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, trafficSplit := range trafficSplits {
		trafficSplitsList.TrafficSplits = append(trafficSplitsList.TrafficSplits, toTrafficSplit(&trafficSplit))
	}

	return trafficSplitsList
}
