package tksql

import (
	"errors"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/takauma/logging"
)

// SQLSession SQLセッション構造体.
type SQLSession struct {
	dbMap        *gorp.DbMap
	mapperConfig *MapperConfig
	mappers      map[string]*mapper
	logger       *logging.Logger
	query        string
}

// NewSQLSession SQLセッション構造体を生成します.
func NewSQLSession(dBConfig *DBConfig, mapperConfig *MapperConfig, logger *logging.Logger) (*SQLSession, error) {
	//DBとのコネクションを取得.
	dbMap, err := conn(dBConfig)

	if err != nil {
		return nil, err
	}

	//マッパーMap.
	mappers := map[string]*mapper{}

	//指定されたすべてのXMLファイルをマッパー構造体に変換しリストに格納.
	for _, filename := range mapperConfig.mapperNames {
		mapper, err := parseMapper(mapperConfig.mappersDir, filename)

		if err != nil {
			return nil, err
		}

		mappers[mapper.Name] = mapper
	}

	//SQLSessionを返す.
	return &SQLSession{
		dbMap:        dbMap,
		mapperConfig: mapperConfig,
		mappers:      mappers,
		logger:       logger,
	}, nil
}

// SelectOne 抽出条件に一致する1レコードを取得します.
func (session *SQLSession) SelectOne(parameter interface{}, result interface{}, mapper, id string) error {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := session.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, s := range m.Select {
				if id == s.ID {
					sql = s.Value
					break
				}
			}
		}
	}

	if sql == "" {
		return errors.New("指定されたSQL文が存在しません。")
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)

	//SQLを取得.
	if _, ok := result.(int); ok {
		r, err := session.dbMap.SelectInt(session.query)
		if err != nil {
			return err
		}
		result = &r
	}

	session.dbMap.SelectOne(result, session.query)

	//クエリを初期化.
	session.clearQuery()

	return nil
}

// SelectList 抽出条件に一致する複数レコードを取得します.
func (session *SQLSession) SelectList(parameter interface{}, resultList interface{}, mapper, id string) error {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := session.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, s := range m.Select {
				if id == s.ID {
					sql = s.Value
					break
				}
			}
		}
	}

	if sql == "" {
		return errors.New("指定されたSQL文が存在しません。")
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)

	//SQLを取得.
	session.dbMap.Select(resultList, session.query)

	//クエリを初期化.
	session.clearQuery()

	return nil
}

// Insert レコードの挿入を行います.
func (session *SQLSession) Insert(parameter interface{}, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := session.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, s := range m.Insert {
				if id == s.ID {
					sql = s.Value
					break
				}
			}
		}
	}

	if sql == "" {
		return 0, errors.New("指定されたSQL文が存在しません。")
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード登録数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	session.clearQuery()

	return int(num), err
}

// Update レコードの更新を行います.
func (session *SQLSession) Update(parameter interface{}, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := session.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, s := range m.Update {
				if id == s.ID {
					sql = s.Value
					break
				}
			}
		}
	}

	if sql == "" {
		return 0, errors.New("指定されたSQL文が存在しません。")
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード更新数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	session.clearQuery()

	return int(num), err
}

// Delete レコードの削除を行います.
func (session *SQLSession) Delete(parameter interface{}, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := session.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, s := range m.Delete {
				if id == s.ID {
					sql = s.Value
					break
				}
			}
		}
	}

	if sql == "" {
		return 0, errors.New("指定されたSQL文が存在しません。")
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード削除数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	session.clearQuery()

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

// parseSQL SQL文を解析します.
func (session *SQLSession) parseSQL(sql string, paramMap map[string]string) {
	// paramMapが未設定(nil or 要素0件)の場合は空のMapで初期化.
	if len(paramMap) == 0 {
		paramMap = map[string]string{}
	}

	templ := template.Must(template.New("sql").Parse(sql))
	templ.Execute(session, paramMap)
}

// clearQuery クエリフィールドを初期化します.
func (session *SQLSession) clearQuery() {
	session.query = ""
}

// convFieldToMap 構造体のフィールドをマップに変換します.
func convFieldToMap(obj interface{}) map[string]string {
	// 構造体がnilの場合は後続処理を行わずnilを返す.
	if obj == nil {
		return nil
	}

	//インターフェースの実態をValue型に変換(フィールドの値を格納する構造体).
	v := reflect.ValueOf(obj).Elem()

	//Value型からタイプ型を取得(構造体の型情報を格納する構造体).
	t := v.Type()

	//フィールド数取得.
	num := t.NumField()

	//フィールドMap.
	fieldMap := map[string]string{}

	//フィールド名と値を取得しMapに格納.
	for i := 0; i < num; i++ {
		fieldName := t.Field(i).Name
		fieldValue := v.FieldByName(fieldName).Interface()

		// 値がNULL(文字列)の場合.
		if fieldValue == "NULL" {
			fieldMap[fieldName] = fmt.Sprintf("%v", fieldValue)
			continue
		}

		// 日時型の場合.
		val, ok := fieldValue.(time.Time)
		if ok {
			fieldMap[fieldName] = val.Format("'2006-01-02 15:04:05.000'")
			continue
		}

		fieldMap[fieldName] = fmt.Sprintf("'%v'", fieldValue)

	}

	return fieldMap
}
