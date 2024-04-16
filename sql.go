package tksql

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-gorp/gorp"
)

const (
	// 変換日時フォーマット.
	convDateTimeFormat = "2006-01-02 15:04:05.0000000"
)

// SQLSession SQLセッション構造体.
type SQLSession struct {
	dbMap        *gorp.DbMap
	mapperConfig *MapperConfig
	mappers      map[string]*mapper
	query        string
	paramMap     map[string]any
	params       []any
}

// dynamicSQL 動的SQL構造体.
type dynamicSQL struct {
	If []struct {
		Test string `xml:"test,attr"`
		Data string `xml:",innerxml"`
	} `xml:"if"`
	Foreach []struct {
		Item       string `xml:"item,attr"`
		Collection string `xml:"collection,attr"`
		Open       string `xml:"open,attr"`
		Separator  string `xml:"separator,attr"`
		Close      string `xml:"close,attr"`
		Data       string `xml:",innerxml"`
	} `xml:"foreach"`
}

type mapWrapper struct {
	Item map[string]any
}

// NewSQLSession SQLセッション構造体を生成します.
func NewSQLSession(dBConfig *DBConfig, mapperConfig *MapperConfig) (*SQLSession, error) {
	// DBとのコネクションを取得.
	dbMap, err := conn(dBConfig)

	if err != nil {
		return nil, err
	}

	// マッパーMap.
	mappers := map[string]*mapper{}

	// 指定されたすべてのXMLファイルをマッパー構造体に変換しリストに格納.
	for _, filename := range mapperConfig.mapperNames {
		mapper, err := parseMapper(mapperConfig.mappersDir, filename)

		if err != nil {
			return nil, err
		}

		mappers[mapper.Name] = mapper
	}

	// SQLSessionを返す.
	return &SQLSession{
		dbMap:        dbMap,
		mapperConfig: mapperConfig,
		mappers:      mappers,
	}, nil
}

// SetLogConfig ログコンフィグを設定する.
func (s *SQLSession) SetLogConfig(config LogConfig) {
	logger = &Logger{config}
}

// SelectOne 抽出条件に一致する1レコードを取得します.
func (s *SQLSession) SelectOne(parameter any, result any, mapper string, id string) error {
	// 初期化.
	defer s.clean()

	// SQL文.
	sqlQuery := ""

	// 指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Select {
				if id == query.ID {
					sqlQuery = query.Value
					break
				}
			}
		}
	}

	if len(sqlQuery) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		logger.Error(err)
		return err
	}

	// クエリを設定.
	s.query = sqlQuery

	logger.Debug("Loaded SQL mapper: " + mapper + "." + id)

	// 構造体の全フィールドをマップに変換.
	if parameter != nil {
		s.convertStructFieldsToMap(parameter)
	}

	// SQLを解析.
	s.parseSQLQuery()

	// SQLを取得.
	switch r := result.(type) {
	case *int:
		i64, err := s.dbMap.SelectInt(s.query, s.params...)
		if err != nil {
			logger.Error(err)
			return err
		}
		*r = int(i64)
		if logger.IsDebagEnable() {
			logger.Debug("Result:", parseAnyValue(r))
		}
	case *time.Time:
		rows, err := s.dbMap.Query(s.query, s.params...)
		if err != nil {
			logger.Error(err)
			return err
		}
		if !rows.Next() {
			return NewNoResultError("0 result records.")
		}
		rows.Scan(r)
		if rows.Next() {
			err := errors.New("multiple result records")
			logger.Error(err)
			return err
		}
		if logger.IsDebagEnable() {
			logger.Debug("Result:", parseAnyValue(r))
		}
	default:
		if err := s.dbMap.SelectOne(r, s.query, s.params...); err != nil {
			logger.Error(err)
			return err
		}
		if logger.IsDebagEnable() {
			logger.Debug("Result:", parseAnyValue(r))
		}
	}

	return nil
}

// SelectList 抽出条件に一致する複数レコードを取得します.
func (s *SQLSession) SelectList(parameter any, resultList any, mapper, id string) error {
	// 初期化.
	defer s.clean()

	// SQL文.
	sqlQuery := ""

	// 指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Select {
				if id == query.ID {
					sqlQuery = query.Value
					break
				}
			}
		}
	}

	if len(sqlQuery) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		logger.Error(err)
		return err
	}

	// クエリを設定.
	s.query = sqlQuery

	logger.Debug("Loaded SQL mapper: " + mapper + "." + id)

	// 構造体の全フィールドをマップに変換.
	if parameter != nil {
		s.convertStructFieldsToMap(parameter)
	}

	// SQLを解析.
	s.parseSQLQuery()

	// SQLを取得.
	_, err := s.dbMap.Select(resultList, s.query, s.params...)
	if err != nil {
		logger.Error(err)
		return err
	}

	// デバッグ時結果ログ出力.
	if logger.IsDebagEnable() {
		result := parseAnyValue(resultList)
		rv := reflect.ValueOf(result)
		logger.Debug("Result row count:", rv.Len())
		switch {
		case rv.Len() < 2:
			logger.Debug("Results:", result)
		default:
			sb := strings.Builder{}
			sb.WriteString("Results:\n")
			for i := 0; i < rv.Len(); i++ {
				v := rv.Index(i).Interface()
				sb.WriteString(fmt.Sprintf("Row[%d]: %v\n", i+1, v))
			}
			logger.Debug(sb.String())
		}
	}

	return nil
}

// Insert レコードの挿入を行います.
func (s *SQLSession) Insert(parameter any, mapper string, id string) (int, error) {
	// 初期化.
	defer s.clean()

	// SQL文.
	sqlQuery := ""

	// 指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Insert {
				if id == query.ID {
					sqlQuery = query.Value
					break
				}
			}
		}
	}

	if len(sqlQuery) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		logger.Error(err)
		return -1, err
	}

	// クエリを設定.
	s.query = sqlQuery

	logger.Debug("Loaded SQL mapper: " + mapper + "." + id)

	// 構造体の全フィールドをマップに変換.
	if parameter != nil {
		s.convertStructFieldsToMap(parameter)
	}

	// SQLを解析.
	s.parseSQLQuery()

	// SQLを実行.
	result, err := s.dbMap.Db.Exec(s.query, s.params...)
	if err != nil {
		logger.Error(err)
		return -1, err
	}

	// レコード登録数を取得.
	num, err := result.RowsAffected()

	logger.Debug("Insert count:", num)

	return int(num), err
}

// Update レコードの更新を行います.
func (s *SQLSession) Update(parameter any, mapper string, id string) (int, error) {
	// 初期化.
	defer s.clean()

	// SQL文.
	sqlQuery := ""

	// 指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Update {
				if id == query.ID {
					sqlQuery = query.Value
					break
				}
			}
		}
	}

	logger.Debug("Loaded SQL mapper: " + mapper + "." + id)

	if len(sqlQuery) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		logger.Error(err)
		return -1, err
	}

	// クエリを設定.
	s.query = sqlQuery

	// 構造体の全フィールドをマップに変換.
	if parameter != nil {
		s.convertStructFieldsToMap(parameter)
	}

	// SQLを解析.
	s.parseSQLQuery()

	// SQLを実行.
	var result sql.Result
	var err error
	if len(s.params) == 0 {
		result, err = s.dbMap.Db.Exec(s.query)
	} else {
		result, err = s.dbMap.Db.Exec(s.query, s.params...)
	}
	if err != nil {
		logger.Error(err)
		return -1, err
	}

	// レコード更新数を取得.
	num, err := result.RowsAffected()

	logger.Debug("Update count:", num)

	return int(num), err
}

// Delete レコードの削除を行います.
func (s *SQLSession) Delete(parameter any, mapper string, id string) (int, error) {
	// 初期化.
	defer s.clean()

	// SQL文.
	sqlQuery := ""

	// 指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Delete {
				if id == query.ID {
					sqlQuery = query.Value
					break
				}
			}
		}
	}

	if len(sqlQuery) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		logger.Error(err)
		return -1, err
	}

	// クエリを設定.
	s.query = sqlQuery

	logger.Debug("Loaded SQL mapper: " + mapper + "." + id)

	// 構造体の全フィールドをマップに変換.
	if parameter != nil {
		s.convertStructFieldsToMap(parameter)
	}

	// SQLを解析.
	s.parseSQLQuery()

	// SQLを実行.
	var result sql.Result
	var err error
	if len(s.params) == 0 {
		result, err = s.dbMap.Db.Exec(s.query)
	} else {
		result, err = s.dbMap.Db.Exec(s.query, s.params...)
	}
	if err != nil {
		logger.Error(err)
		return -1, err
	}

	// レコード削除数を取得.
	num, err := result.RowsAffected()

	logger.Debug("Delete count:", num)

	return int(num), err
}

// GetDbMap DbMapを取得します.
func (session *SQLSession) GetDbMap() *gorp.DbMap {
	return session.dbMap
}

// Write queryフィールドに書き込みを行います.
// io.Writerインターフェースの実装メソッド.
// テンプレートのマッピング実行(parseSQL)で利用.
func (session *SQLSession) Write(p []byte) (n int, err error) {
	session.query += string(p)
	n = len(p)

	return n, nil
}

// clean フィールドの初期化を行います.
func (s *SQLSession) clean() {
	s.query = ""
	s.paramMap = map[string]any{}
	s.params = []any{}
}

// convertStructFieldsToMap 構造体のフィールドをマップに変換します.
func (s *SQLSession) convertStructFieldsToMap(obj any) {
	// 単一構造体のマップ変換処理.
	convert := func(v reflect.Value) map[string]any {
		// インターフェースに格納された構造体情報を取得.
		t := v.Type()

		// 構造体のフィールド数取得.
		num := t.NumField()

		// フィールド名と値のマッパー.
		fieldMap := map[string]any{}

		//フィールド名と値を取得しMapに格納.
		for i := 0; i < num; i++ {
			// フィールド名取得.
			fieldName := t.Field(i).Name
			// フィールド値取得.
			fieldValue := v.FieldByName(fieldName).Interface()

			// フィールド値がnilの場合はnilで設定し後続処理をスキップ.
			if reflect.ValueOf(fieldValue).Kind() == reflect.Pointer {
				if isNil(fieldValue) {
					fieldMap[fieldName] = nil
					continue
				}
			}

			switch value := fieldValue.(type) {
			case time.Time, *time.Time:
				switch v := value.(type) {
				case time.Time:
					// TODO DBによりフォーマットを切り替える必要がある.
					s := v.Format(convDateTimeFormat)
					fieldMap[fieldName] = s
				case *time.Time:
					// TODO DBによりフォーマットを切り替える必要がある.
					s := v.Format(convDateTimeFormat)
					fieldMap[fieldName] = &s
				}
			default:
				fieldMap[fieldName] = fieldValue
			}
		}

		return fieldMap
	}

	// 構造体情報取得.
	v := reflect.ValueOf(obj)

	// ポインタ型の場合は実体のValueを取得.
	if reflect.ValueOf(obj).Kind() == reflect.Pointer {
		v = reflect.ValueOf(obj).Elem()
	}

	// インターフェースの実態がスライス型の場合.
	if v.Kind() == reflect.Slice {
		wrappers := []mapWrapper{}
		for i := 0; i < v.Len(); i++ {
			m := convert(v.Index(i))
			wrapper := mapWrapper{m}
			wrappers = append(wrappers, wrapper)
		}
		s.paramMap = map[string]any{"slice": wrappers}
		return
	}

	// インターフェースの実態がスライス型でない場合.
	s.paramMap = convert(v)
}

// parseSQLQuery SQL文の解析を行う.
func (s *SQLSession) parseSQLQuery() {
	// 解析後クエリ.
	query := queryOptimize(s.query)

	paramType := "normal"

	for {
		parsingQuery := query

		// 動的SQLの構造体マッピング.
		dynamics := []dynamicSQL{}
		xml.Unmarshal([]byte("<xml>"+parsingQuery+"</xml>"), &dynamics)

		// 動的SQLを1つずつ解析.
		for _, dynamic := range dynamics {
			// if文.
			for _, tagIf := range dynamic.If {
				test := tagIf.Test
				switch {
				case test == "true":
					// 真の場合.
					parsingQuery = replaceTagIf(parsingQuery, test, tagIf.Data, tagIf.Data)
				case test == "false":
					// 偽の場合.
					parsingQuery = replaceTagIf(parsingQuery, test, tagIf.Data, "")
				default:
					// 比較式の場合.
					matches := regexp.MustCompile(`(==|\!=|<=|>=|<|>)`).FindAllString(test, 1)
					if len(matches) != 1 {
						break
					}
					split := strings.Split(test, matches[0])
					for i, item := range split {
						split[i] = strings.TrimSpace(item)
					}
					if len(split) == 2 {
						operator := matches[0]
						valA := parseAttrInlineValue(s.paramMap, split[0])
						valB := parseAttrInlineValue(s.paramMap, split[1])
						if evaluation(operator, valA, valB) {
							parsingQuery = replaceTagIf(parsingQuery, test, tagIf.Data, tagIf.Data)
						} else {
							parsingQuery = replaceTagIf(parsingQuery, test, tagIf.Data, "")
						}
						break
					}

					panic("SQL mapper comparison expression is invalid.")
				}
			}
			// foreach文.
			for _, tagForeach := range dynamic.Foreach {
				data := tagForeach.Data
				open := tagForeach.Open
				separator := tagForeach.Separator
				close := tagForeach.Close

				collection := tagForeach.Collection
				if collection == "slice" {
					paramType = collection
				}
				a := s.paramMap[collection]
				if wrappers, ok := a.([]mapWrapper); ok {
					parsingQuery = replaceTagForeach(parsingQuery, data, open, separator, close, len(wrappers))
				}
			}
		}

		if query == parsingQuery {
			break
		}
		query = parsingQuery
	}

	// 動的パラメータを解析.
	r := regexp.MustCompile(`#{([A-z]|_)([A-z0-9]|_)*}`)
	keys := r.FindAllString(query, -1)
	values := []any{}

	if paramType == "slice" {
		// パラメータがスライスの場合.

		// マップラッパー取得.
		wrappers, _ := s.paramMap["slice"].([]mapWrapper)

		// マッピング済キーマップを初期化.
		mapedKeyMaps := map[int]map[string]bool{}
		for i := 0; i < len(wrappers); i++ {
			mapedKeyMaps[i] = map[string]bool{}
			for _, key := range keys {
				k := key[2 : len(key)-1]
				mapedKeyMaps[i][k] = false
			}
		}

		// パラメータ値をオーダーして取得.
		for i := 0; i < len(wrappers); i++ {
			for _, key := range keys {
				key = key[2 : len(key)-1]
				wrapper := wrappers[i]
				if value, ok := wrapper.Item[key]; ok {
					values = append(values, value)
					delete(wrapper.Item, key)
				}
			}
		}
	} else {
		// パラメータ値をオーダーして取得.
		for _, key := range keys {
			values = append(values, s.paramMap[key[2:len(key)-1]])
		}
	}

	// 解析後のクエリとパラメータを設定.
	query = r.ReplaceAllString(query, "?")
	s.query = queryOptimize(query)
	s.params = values

	// デバッグ時パラメータログ出力.
	if logger.IsDebagEnable() {
		logger.Debug("Parsed SQL:", s.query)
		logger.Debug("Params:", parseAnyValue(s.params))
	}
}
