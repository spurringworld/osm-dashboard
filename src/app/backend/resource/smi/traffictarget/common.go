package traffictarget

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"

	smiaccessv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha3"
)

// The code below allows to perform complex data section on []api.TrafficTarget

type TrafficTargetCell smiaccessv1alpha3.TrafficTarget

func (self TrafficTargetCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []smiaccessv1alpha3.TrafficTarget) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = TrafficTargetCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []smiaccessv1alpha3.TrafficTarget {
	std := make([]smiaccessv1alpha3.TrafficTarget, len(cells))
	for i := range std {
		std[i] = smiaccessv1alpha3.TrafficTarget(cells[i].(TrafficTargetCell))
	}
	return std
}
