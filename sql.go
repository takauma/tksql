package tksql

import (
	"errors"
	"fmt"
	"reflect"
	"text/template"

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

func (session *SQLSession) DbMap() *gorp.DbMap {
	return session.dbMap
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
func (session *SQLSession) SelectOne(parameter, result interface{}, mapper, id string) error {
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
	//session.logger.Debug("動的SQLパラメータ:", paramMap)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)
	//session.logger.Debug("実行SQL文:", session.query)

	//SQLを取得.
	//session.logger.Debug("実行前結果オブジェクト:", fmt.Sprintf("%v", result))
	session.dbMap.SelectOne(result, session.query)
	//session.logger.Debug("実行後結果オブジェクト:", fmt.Sprintf("%v", result))

	//クエリを初期化.
	session.clearQuery()
	//session.logger.Debug("初期化後SQL文格納フィールド:", session.query)

	return nil
}

// SelectList 抽出条件に一致する複数レコードを取得します.
func (session *SQLSession) SelectList(parameter, resultList interface{}, mapper, id string) error {
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
	//session.logger.Debug("動的SQLパラメータ:", paramMap)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)
	//session.logger.Debug("実行SQL文:", session.query)

	//SQLを取得.
	//session.logger.Debug("実行前結果リスト:", fmt.Sprintf("%v", resultList))
	session.dbMap.Select(resultList, session.query)
	//session.logger.Debug("実行後結果リスト:", fmt.Sprintf("%v", resultList))

	//クエリを初期化.
	session.clearQuery()
	//session.logger.Debug("初期化後SQL文格納フィールド:", session.query)

	return nil
}

// Insert レコードの挿入を行います.
func (session *SQLSession) Insert(parameter interface{}, mapper, id string) (int, error) {
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
	//session.logger.Debug("動的SQLパラメータ:", paramMap)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)
	//session.logger.Debug("実行SQL文:", session.query)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード登録数を取得.
	num, err := result.RowsAffected()
	//session.logger.Debug("レコード登録数:", num)

	//クエリを初期化.
	session.clearQuery()
	//session.logger.Debug("初期化後SQL文格納フィールド:", session.query)

	return int(num), err
}

// Update レコードの更新を行います.
func (session *SQLSession) Update(parameter interface{}, mapper, id string) (int, error) {
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
	//session.logger.Debug("動的SQLパラメータ:", paramMap)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)
	//session.logger.Debug("実行SQL文:", session.query)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード更新数を取得.
	num, err := result.RowsAffected()
	//session.logger.Debug("レコード更新数:", num)

	//クエリを初期化.
	session.clearQuery()
	//session.logger.Debug("初期化後SQL文格納フィールド:", session.query)

	return int(num), err
}

// Delete レコードの削除を行います.
func (session *SQLSession) Delete(parameter interface{}, mapper, id string) (int, error) {
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
	//session.logger.Debug("動的SQLパラメータ:", paramMap)

	//SQLテンプレートを解析.
	session.parseSQL(sql, paramMap)
	//session.logger.Debug("実行SQL文:", session.query)

	//SQLを実行.
	result, err := session.dbMap.Db.Exec(session.query)
	if err != nil {
		return 0, err
	}

	//レコード削除数を取得.
	num, err := result.RowsAffected()
	//session.logger.Debug("レコード削除数:", num)

	//クエリを初期化.
	session.clearQuery()
	//session.logger.Debug("初期化後SQL文格納フィールド:", session.query)

	return int(num), err
}

// parseSQL SQL文を解析します.
func (session *SQLSession) parseSQL(sql string, paramMap map[string]string) {
	if len(paramMap) == 0 {
		return
	}

	templ := template.Must(template.New("sql").Parse(sql))
	templ.Execute(session, paramMap)
}

// Write queryフィールドに書き込みを行います.
func (session *SQLSession) Write(p []byte) (n int, err error) {
	session.query += string(p)
	n = len(p)

	return n, nil
}

// clearQuery クエリフィールドを初期化します.
func (session *SQLSession) clearQuery() {
	session.query = ""
}

// convFieldToMap 構造体のフィールドをマップに変換します.
func convFieldToMap(obj interface{}) map[string]string {
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

		if fieldValue == "NULL" {
			fieldMap[fieldName] = fmt.Sprintf("%v", fieldValue)
		} else {
			fieldMap[fieldName] = fmt.Sprintf("'%v'", fieldValue)
		}

	}

	return fieldMap
}
