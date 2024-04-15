package tksql

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	operatorEq    = "=="
	operatorNe    = "!="
	operatorLt    = "<"
	operatorGt    = ">"
	operatorLe    = "<="
	operatorGe    = ">="
	operatorLtXml = "&lt;"
	operatorGtXml = "&gt;"
)

var (
	OperatorReplacer = strings.NewReplacer(
		operatorLt, operatorLtXml,
		operatorGt, operatorGtXml,
	)
)

// replaceTagIf if文を解析結果で置き換えます.
func replaceTagIf(query, test, data, replaceStr string) string {

	match := regexp.MustCompile(`<if +test *= *"`+OperatorReplacer.Replace(test)+`" *>`+data+`</if>`).FindAllString(query, 1)[0]
	return strings.Replace(query, match, replaceStr, 1)
}

// replaceTagForeach foreach文を解析結果で置き換えます.
func replaceTagForeach(query, data, open, separator, close string, length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s += open + data + close
		s += separator
	}
	s = s[:len(s)-len(separator)]

	match := regexp.MustCompile(`<foreach.*>`+data+`</foreach>`).FindAllString(query, 1)[0]
	return strings.Replace(query, match, s, 1)
}

// parseAttrInlineValue 属性内で出現する値を解析します.
func parseAttrInlineValue(paramMap map[string]any, txt string) any {
	// nil.
	if txt == "nil" {
		return nil
	}
	// 文字列.
	if txt == `''` || txt == `""` {
		return ""
	}
	start := txt[:1]
	end := txt[len(txt)-1:]
	if (start == `"` || start == `'`) && (end == `"` || end == `'`) {
		return txt[1 : len(txt)-1]
	}
	// パラメータ指定.
	if v, ok := paramMap[txt]; ok {
		return v
	}
	// その他.
	return txt
}

// evaluation 評価を行います.
func evaluation(operator string, a any, b any) bool {
	// a, bがnilの場合.
	if a == nil && b == nil {
		switch operator {
		case operatorEq:
			return true
		case operatorNe:
			return false
		}
	}

	// aがnilの場合.
	if a == nil && b != nil {
		if reflect.ValueOf(b).Kind() != reflect.Pointer {
			b = &b
		}
		switch operator {
		case operatorEq:
			return b == nil
		case operatorNe:
			return b != nil
		}
	}
	// bがnilの場合.
	if a != nil && b == nil {
		if reflect.ValueOf(a).Kind() != reflect.Pointer {
			a = &a
		}
		switch operator {
		case operatorEq:
			return a == nil
		case operatorNe:
			return a != nil
		}
	}

	// ポインタ型参照している実体を取得.
	rvA := reflect.ValueOf(a)
	if rvA.Kind() == reflect.Pointer {
		a = rvA.Elem().Interface()
	}
	rvB := reflect.ValueOf(b)
	if rvB.Kind() == reflect.Pointer {
		b = rvB.Elem().Interface()
	}

	// a, bが文字列の場合.
	if isKindOrPointerKind(a, reflect.String) && isKindOrPointerKind(b, reflect.String) {
		switch operator {
		case operatorEq:
			return a == b
		case operatorNe:
			return a != b
		}
	}

	// float64に変換を試みる.
	f64A := parseFloat64(a)
	f64B := parseFloat64(b)
	if !isNil(f64A) && !isNil(f64B) {
		return evaluationNum(operator, *f64A, *f64B)
	}

	// float32に変換を試みる.
	f32A := parseFloat32(a)
	f32B := parseFloat32(b)
	if !isNil(f32A) && !isNil(f32B) {
		return evaluationNum(operator, *f32A, *f32B)
	}

	return false
}

// evaluationNum 数値の評価を行います.
func evaluationNum[T int | float32 | float64](operator string, a T, b T) bool {
	switch operator {
	case operatorEq:
		return a == b
	case operatorNe:
		return a != b
	case operatorLt:
		return a < b
	case operatorGt:
		return a > b
	case operatorLe:
		return a <= b
	case operatorGe:
		return a >= b
	}
	return false
}

// parseFloat32 値を解析しflaot64に変換します.
func parseFloat32(value any) *float32 {
	switch v := value.(type) {
	case int:
		r := float32(v)
		return &r
	case *int:
		r := float32(*v)
		return &r
	case float32:
		return &v
	case *float32:
		return v
	case float64:
		r := float32(v)
		return &r
	case *float64:
		r := float32(*v)
		return &r
	case string:
		r, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil
		}
		r2 := float32(r)
		return &r2
	case *string:
		r, err := strconv.ParseFloat(*v, 32)
		if err != nil {
			return nil
		}
		r2 := float32(r)
		return &r2
	}
	return nil
}

// parseFloat64 値を解析しflaot64に変換します.
func parseFloat64(value any) *float64 {
	switch v := value.(type) {
	case int:
		r := float64(v)
		return &r
	case *int:
		r := float64(*v)
		return &r
	case float64:
		return &v
	case *float64:
		return v
	case string:
		r, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil
		}
		return &r
	case *string:
		r, err := strconv.ParseFloat(*v, 64)
		if err != nil {
			return nil
		}
		return &r
	}
	return nil
}

// queryOptimize SQL文の整形を行います.
func queryOptimize(query string) string {
	query = regexp.MustCompile(`(\r\n|\r|\n)`).ReplaceAllString(query, "")
	query = regexp.MustCompile(`( |\t)+`).ReplaceAllString(query, " ")
	return strings.TrimSpace(query)
}

// isKindOrPointerKind 型もしくはポインタが参照している型の種類が一致するか判定します.
func isKindOrPointerKind(obj any, targetKind reflect.Kind) bool {
	kind := reflect.ValueOf(obj).Kind()
	if kind == reflect.Pointer {
		kind = reflect.ValueOf(obj).Elem().Kind()
	}
	return kind == targetKind
}

// isNil ポインタがnilであるか判定します.
func isNil(obj any) bool {
	if reflect.ValueOf(obj).Kind() == reflect.Pointer {
		return obj == nil || reflect.ValueOf(obj).IsNil()
	}
	return false
}
