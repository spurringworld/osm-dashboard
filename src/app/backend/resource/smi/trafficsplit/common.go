package trafficsplit

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
)

// The code below allows to perform complex data section on []api.TrafficTarget

type TrafficSplitCell smisplitv1alpha2.TrafficSplit

func (self TrafficSplitCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []smisplitv1alpha2.TrafficSplit) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = TrafficSplitCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []smisplitv1alpha2.TrafficSplit {
	std := make([]smisplitv1alpha2.TrafficSplit, len(cells))
	for i := range std {
		std[i] = smisplitv1alpha2.TrafficSplit(cells[i].(TrafficSplitCell))
	}
	return std
}
