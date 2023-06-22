package fieldmap

import "fmt"

// Mapper ...
type Mapper[F1 Field, T1 MapType[F1], F2 Field, T2 MapType[F2]] struct {
	parentOf func(source F1) F1
	fieldMap map[F1][][]F2
	mappings []MappingData[F1, F2]
}

// MappingData ...
type MappingData[F1, F2 Field] struct {
	from   F1
	toList []F2
}

// MappingOption ...
type MappingOption[F1 Field, T1 MapType[F1], F2 Field, T2 MapType[F2]] func(
	mappings []MappingData[F1, F2],
) []MappingData[F1, F2]

// NewMapping ...
func NewMapping[F1, F2 Field](
	from F1, toList ...F2,
) MappingData[F1, F2] {
	if len(toList) == 0 {
		panic("missing destination fields")
	}
	return MappingData[F1, F2]{from: from, toList: toList}
}

// WithSimpleMapping ...
func WithSimpleMapping[F1 Field, T1 MapType[F1], F2 Field, T2 MapType[F2]](
	_ *FieldMap[F1, T1], _ *FieldMap[F2, T2],
	mappings ...MappingData[F1, F2],
) MappingOption[F1, T1, F2, T2] {
	return func(result []MappingData[F1, F2]) []MappingData[F1, F2] {
		return append(result, mappings...)
	}
}

// WithInheritMapping ...
func WithInheritMapping[F1 Field, T1 MapType[F1], F2 Field, T2 MapType[F2], SubT1 MapType[F1], SubT2 MapType[F2]](
	source *FieldMap[F1, T1], dest *FieldMap[F2, T2],
	inherit *Mapper[F1, SubT1, F2, SubT2],
	sourceFunc func(m T1) SubT1, destFunc func(m T2) SubT2,
) MappingOption[F1, T1, F2, T2] {
	return func(mappings []MappingData[F1, F2]) []MappingData[F1, F2] {
		sourceDiff := sourceFunc(source.GetMapping()).GetRoot() - 1
		destDiff := destFunc(dest.GetMapping()).GetRoot() - 1

		for _, subMapping := range inherit.mappings {
			newToList := make([]F2, 0, len(subMapping.toList))
			for _, to := range subMapping.toList {
				newToList = append(newToList, to+destDiff)
			}

			mappings = append(mappings, MappingData[F1, F2]{
				from:   subMapping.from + sourceDiff,
				toList: newToList,
			})
		}
		return mappings
	}
}

type emptyStruct struct{}

// NewMapper ...
func NewMapper[F1 Field, T1 MapType[F1], F2 Field, T2 MapType[F2]](
	source *FieldMap[F1, T1], dest *FieldMap[F2, T2],
	mappings ...MappingOption[F1, T1, F2, T2],
) *Mapper[F1, T1, F2, T2] {
	fieldMap := map[F1][][]F2{}
	dedupSets := map[F1]map[F2]emptyStruct{}

	getDedupSet := func(source F1) map[F2]emptyStruct {
		s, ok := dedupSets[source]
		if !ok {
			s = map[F2]emptyStruct{}
		}
		dedupSets[source] = s
		return s
	}

	var mappingDataList []MappingData[F1, F2]
	for _, option := range mappings {
		mappingDataList = option(mappingDataList)
	}

	for _, m := range mappingDataList {
		set := getDedupSet(m.from)
		if len(m.toList) == 1 {
			for _, to := range m.toList {
				_, existed := set[to]
				if existed {
					panic(fmt.Sprintf(
						"duplicated destination field %q for source field %q",
						dest.GetFullFieldName(to),
						source.GetFullFieldName(m.from),
					))
				}
				set[to] = emptyStruct{}
			}
		}
		fieldMap[m.from] = append(fieldMap[m.from], m.toList)
	}

	return &Mapper[F1, T1, F2, T2]{
		parentOf: source.ParentOf,
		fieldMap: fieldMap,
		mappings: mappingDataList,
	}
}
func (m *Mapper[F1, T1, F2, T2]) findMappedFieldsForSourceField(
	sourceField F1, resultSet map[F2]emptyStruct, result []F2,
) []F2 {
	var empty F1

	for {
		for _, destFields := range m.fieldMap[sourceField] {
			for _, f := range destFields {
				_, existed := resultSet[f]
				if existed {
					continue
				}
				resultSet[f] = emptyStruct{}
				result = append(result, f)
			}
			if len(destFields) > 0 {
				return result
			}
		}

		sourceField = m.parentOf(sourceField)
		if sourceField == empty {
			return result
		}
	}
}

// FindMappedFields ...
func (m *Mapper[F1, T1, F2, T2]) FindMappedFields(sourceFields []F1) []F2 {
	var result []F2
	resultSet := map[F2]emptyStruct{}

	for _, sourceField := range sourceFields {
		result = m.findMappedFieldsForSourceField(sourceField, resultSet, result)
	}

	return result
}
