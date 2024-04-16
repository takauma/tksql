package tksql

import (
	"reflect"
	"time"
)

// isNil ポインタがnilであるか判定します.
func isNil(obj any) bool {
	if isKindNilable(reflect.ValueOf(obj)) {
		return obj == nil || reflect.ValueOf(obj).IsNil()
	}
	return false
}

// isKindOrPointerKind 型もしくはポインタが参照している型の種類が一致するか判定します.
func isKindOrPointerKind(obj any, targetKind reflect.Kind) bool {
	kind := reflect.ValueOf(obj).Kind()
	if kind == reflect.Pointer {
		kind = reflect.ValueOf(obj).Elem().Kind()
	}
	return kind == targetKind
}

// getInstanceReflectValue ポインタが刺す実体の情報を取得します.
func getInstanceReflectValue(rv reflect.Value) reflect.Value {
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		newRv := rv.Elem()
		if newRv.Kind() == reflect.Invalid {
			return rv
		}
		rv = newRv
	}
	return rv
}

// getInstanceValue 実体の値を取得します.
func getInstanceValue(rv reflect.Value) any {
	v := rv.Interface()
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05.0000000")
	}
	return v
}

// getStructInstanceFieldValues 構造体実体のフィールド値スライスを取得します.
func getStructInstanceFieldValues(rv reflect.Value) []any {
	// フィールド値スライス.
	fieldValues := []any{}

	// フィールドのループ.
	for i := 0; i < rv.NumField(); i++ {
		// フィールドがnil.
		if isKindNilable(rv) {
			if rv.Field(i).IsNil() {
				fieldValues = append(fieldValues, nil)
				continue
			}
		}

		// フィールドの実体情報取得.
		rvField := getInstanceReflectValue(rv.Field(i))

		// 取得フィールド値をスライスに追加.
		fieldValues = append(fieldValues, getInstanceValue(rvField))
	}

	return fieldValues
}

// getSliceInstance スライス実体を取得します.
func getSliceInstance(rvSlice reflect.Value) []any {
	// 解析後スライス.
	parsedSlice := []any{}

	// スライスのループ.
	for i := 0; i < rvSlice.Len(); i++ {
		// 要素がnilの場合.
		if isKindNilable(rvSlice) {
			if rvSlice.Index(i).IsNil() {
				parsedSlice = append(parsedSlice, nil)
				continue
			}
		}

		// スライス要素の実体情報取得.
		rv := getInstanceReflectValue(rvSlice.Index(i))

		// 構造体の場合.
		if rv.Kind() == reflect.Struct {
			// 日時型の場合はフィールド値を取得しない.
			if _, ok := rv.Interface().(time.Time); ok {
				parsedSlice = append(parsedSlice, getInstanceValue(rv))
			}

			parsedSlice = append(parsedSlice, getStructInstanceFieldValues(rv))
			continue
		}

		// 構造体以外の場合.
		parsedSlice = append(parsedSlice, rv.Interface())
	}

	return parsedSlice
}

// parseAnyValue Any型変数の解析を行います.
func parseAnyValue(anyValue any) any {
	// nilの場合.
	if isNil(anyValue) {
		return nil
	}

	rv := getInstanceReflectValue(reflect.ValueOf(anyValue))

	// 構造体の場合.
	if rv.Kind() == reflect.Struct {
		// 日時型の場合はフィールド値を取得しない.
		if _, ok := rv.Interface().(time.Time); ok {
			return getInstanceValue(rv)
		}
		return getStructInstanceFieldValues(rv)
	}

	// スライス以外の場合.
	if rv.Kind() != reflect.Slice {
		return getInstanceValue(rv)
	}

	// スライスの場合.
	return getSliceInstance(rv)
}

// isKindNilable nil許容する型種別か判定します.
func isKindNilable(rv reflect.Value) bool {
	return rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface
}
