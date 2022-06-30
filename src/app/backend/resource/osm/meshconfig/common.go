package meshconfig

import (
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"

	osmconfigv1alph2 "github.com/openservicemesh/osm/pkg/apis/config/v1alpha2"
)

// The code below allows to perform complex data section on []api.MeshConfig

type MeshConfigCell osmconfigv1alph2.MeshConfig

func (self MeshConfigCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
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

func toCells(std []osmconfigv1alph2.MeshConfig) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = MeshConfigCell(std[i])
	}
	return cells
}

func fromCells(cells []dataselect.DataCell) []osmconfigv1alph2.MeshConfig {
	std := make([]osmconfigv1alph2.MeshConfig, len(cells))
	for i := range std {
		std[i] = osmconfigv1alph2.MeshConfig(cells[i].(MeshConfigCell))
	}
	return std
}
