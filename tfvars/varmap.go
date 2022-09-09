package tfvars

// VariableMap is a collection of variable key value pairs stored within a map.
type VariableMap map[string]string

// Merge combines the contents of two VariableMap structs. If a key exists in both variable maps,
// then the value from the other VariableMap is used in the output map
func (vm *VariableMap) Merge(other VariableMap) VariableMap {
	combinationMap := VariableMap{}

	for k, v := range *vm {
		combinationMap[k] = v
	}

	for k, v := range other {
		combinationMap[k] = v
	}

	return combinationMap
}
