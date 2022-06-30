package httproutegroup

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	smispecsv1alpha4 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha4"
)

// The code below allows to perform complex data section on []api.TrafficTarget

type HttpRouteGroupCell smispecsv1alpha4.HTTPRouteGroup

func (self HttpRouteGroupCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(self.ObjectMeta.CreationTimestamp.Time)
	case dataselect.NamespaceProperty:
		return dataselect.StdComparableString(self.ObjectMeta.Namespace)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

func toCells(std []smispecsv1alpha4.HTTPRouteGroup) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = HttpRouteGroupCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []smispecsv1alpha4.HTTPRouteGroup {
	std := make([]smispecsv1alpha4.HTTPRouteGroup, len(cells))
	for i := range std {
		std[i] = smispecsv1alpha4.HTTPRouteGroup(cells[i].(HttpRouteGroupCell))
	}
	return std
}
