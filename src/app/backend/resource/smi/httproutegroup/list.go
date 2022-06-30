package httproutegroup

import (
	"log"

	smispecsv1alpha4 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha4"
	smispecsclientset "github.com/servicemeshinterface/smi-sdk-go/pkg/gen/client/specs/clientset/versioned"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
)

// HttpRouteGroup is a representation of a httpgroup.
type HttpRouteGroup struct {
	ObjectMeta api.ObjectMeta `json:"objectMeta"`
	TypeMeta   api.TypeMeta   `json:"typeMeta"`
}

// HttpRouteGroupList contains a list of services in the cluster.
type HttpRouteGroupList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of httpgroups.
	HttpRouteGroups []HttpRouteGroup `json:"httpgroups"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}

// GetServiceList returns a list of all services in the cluster.
func GetHttpRouteGroupList(smiSpecsClient smispecsclientset.Interface, nsQuery *common.NamespaceQuery,
	dsQuery *dataselect.DataSelectQuery) (*HttpRouteGroupList, error) {
	log.Print("Getting list of all http route group in the cluster")

	channels := &common.ResourceChannels{
		HttpRouteGroupList: common.GetHttpRouteGroupListChannel(smiSpecsClient, nsQuery, 1),
	}

	return GetHttpRouteGroupListFromChannels(channels, dsQuery)
}

// GetHttpRouteGroupListFromChannels returns a list of all services in the cluster.
func GetHttpRouteGroupListFromChannels(channels *common.ResourceChannels,
	dsQuery *dataselect.DataSelectQuery) (*HttpRouteGroupList, error) {
	httpRouteGroups := <-channels.HttpRouteGroupList.List
	err := <-channels.HttpRouteGroupList.Error
	nonCriticalErrors, criticalError := errors.HandleError(err)
	if criticalError != nil {
		return nil, criticalError
	}

	return CreateHttpRouteGroupList(httpRouteGroups.Items, nonCriticalErrors, dsQuery), nil
}

func toHttpRouteGroup(httpRouteGroup *smispecsv1alpha4.HTTPRouteGroup) HttpRouteGroup {
	return HttpRouteGroup{
		ObjectMeta: api.NewObjectMeta(httpRouteGroup.ObjectMeta),
		TypeMeta:   api.NewTypeMeta(api.ResourceKindService),
	}
}

// CreateHttpRouteGroupList returns paginated httpgroup list based on given httpgroup array and pagination query.
func CreateHttpRouteGroupList(httpRouteGroups []smispecsv1alpha4.HTTPRouteGroup, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *HttpRouteGroupList {
	httpRouteGroupsList := &HttpRouteGroupList{
		HttpRouteGroups: make([]HttpRouteGroup, 0),
		ListMeta:        api.ListMeta{TotalItems: len(httpRouteGroups)},
		Errors:          nonCriticalErrors,
	}

	httpRouteGroupCells, filteredTotal := dataselect.GenericDataSelectWithFilter(toCells(httpRouteGroups), dsQuery)
	httpRouteGroups = fromCells(httpRouteGroupCells)
	httpRouteGroupsList.ListMeta = api.ListMeta{TotalItems: filteredTotal}

	for _, httpRouteGroup := range httpRouteGroups {
		httpRouteGroupsList.HttpRouteGroups = append(httpRouteGroupsList.HttpRouteGroups, toHttpRouteGroup(&httpRouteGroup))
	}

	return httpRouteGroupsList
}
