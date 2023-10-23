package importer

import (
	"fmt"
	"reflect"
	"sort"
)

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func variableRef(name string, k string) string {
	return fmt.Sprintf("var.%s", variableName(name, k))
}

func variableName(name string, k string) string {
	return fmt.Sprintf("%s_%s", name, k)
}

func isEmpty(in any) bool {
	return reflect.DeepEqual(in, reflect.Zero(reflect.TypeOf(in)).Interface())

}
