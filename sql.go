package tksql

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-gorp/gorp"
)

// SQLSession SQLセッション構造体.
type SQLSession struct {
	dbMap        *gorp.DbMap
	mapperConfig *MapperConfig
	mappers      map[string]*mapper
	logConfig    *LogConfig
	query        string
}

// NewSQLSession SQLセッション構造体を生成します.
func NewSQLSession(dBConfig *DBConfig, mapperConfig *MapperConfig) (*SQLSession, error) {
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
	}, nil
}

// SetLogConfig ログコンフィグを設定する.
func (s *SQLSession) SetLogConfig(config *LogConfig) {
	s.logConfig = config
}

// SelectOne 抽出条件に一致する1レコードを取得します.
func (s *SQLSession) SelectOne(parameter any, result any, mapper string, id string) error {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Select {
				if id == query.ID {
					sql = query.Value
					break
				}
			}
		}
	}

	if len(sql) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return err
	}

	s.log(LogLevelDebug, "Loaded SQL mapper: "+mapper+"."+id)

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	s.parseSQL(sql, paramMap)

	//クエリを初期化.
	defer s.clearQuery()

	//SQLを取得.
	switch r := result.(type) {
	case *int:
		i64, err := s.dbMap.SelectInt(s.query)
		if err != nil {
			s.logConfig.Error(fmt.Sprintf("%v", err))
			return err
		}
		*r = int(i64)
	case *time.Time:
		rows, err := s.dbMap.Query(s.query)
		if err != nil {
			s.logConfig.Error(fmt.Sprintf("%v", err))
			return err
		}
		if !rows.Next() {
			err := errors.New("query was empty")
			s.logConfig.Error(fmt.Sprintf("%v", err))
			return err
		}
		rows.Scan(r)
		if rows.Next() {
			err := errors.New("query was multiple")
			s.logConfig.Error(fmt.Sprintf("%v", err))
			return err
		}
	default:
		if err := s.dbMap.SelectOne(r, s.query); err != nil {
			s.logConfig.Error(fmt.Sprintf("%v", err))
			return err
		}
	}
	s.log(LogLevelDebug, fmt.Sprintf("Result: %v", result))
	return nil
}

// SelectList 抽出条件に一致する複数レコードを取得します.
func (s *SQLSession) SelectList(parameter any, resultList any, mapper, id string) error {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Select {
				if id == query.ID {
					sql = query.Value
					break
				}
			}
		}
	}

	if len(sql) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return err
	}

	s.log(LogLevelDebug, "Loaded SQL mapper: "+mapper+"."+id)

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	s.parseSQL(sql, paramMap)

	//SQLを取得.
	_, err := s.dbMap.Select(resultList, s.query)
	if err != nil {
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return err
	}

	s.log(LogLevelDebug, fmt.Sprintf("Result: %v", resultList))

	//クエリを初期化.
	s.clearQuery()

	return nil
}

// Insert レコードの挿入を行います.
func (s *SQLSession) Insert(parameter any, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Insert {
				if id == query.ID {
					sql = query.Value
					break
				}
			}
		}
	}

	if len(sql) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	s.log(LogLevelDebug, "Loaded SQL mapper: "+mapper+"."+id)

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	s.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := s.dbMap.Db.Exec(s.query)
	if err != nil {
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	//レコード登録数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	s.clearQuery()

	return int(num), err
}

// Update レコードの更新を行います.
func (s *SQLSession) Update(parameter any, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Update {
				if id == query.ID {
					sql = query.Value
					break
				}
			}
		}
	}

	s.log(LogLevelDebug, "Loaded SQL mapper: "+mapper+"."+id)

	if len(sql) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	s.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := s.dbMap.Db.Exec(s.query)
	if err != nil {
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	//レコード更新数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	s.clearQuery()

	return int(num), err
}

// Delete レコードの削除を行います.
func (s *SQLSession) Delete(parameter any, mapper string, id string) (int, error) {
	//SQL文.
	sql := ""

	//指定のSQL文を取得.
	if m, ok := s.mappers[mapper]; ok {
		if mapper == m.Name {
			for _, query := range m.Delete {
				if id == query.ID {
					sql = query.Value
					break
				}
			}
		}
	}

	if len(sql) == 0 {
		err := errors.New("specified mapper or id does not exist. mapper: " + mapper + ", id: " + id)
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	s.log(LogLevelDebug, "Loaded SQL mapper: "+mapper+"."+id)

	//構造体の全フィールドをマップに変換.
	paramMap := convFieldToMap(parameter)

	//SQLテンプレートを解析.
	s.parseSQL(sql, paramMap)

	//SQLを実行.
	result, err := s.dbMap.Db.Exec(s.query)
	if err != nil {
		s.logConfig.Error(fmt.Sprintf("%v", err))
		return -1, err
	}

	//レコード削除数を取得.
	num, err := result.RowsAffected()

	//クエリを初期化.
	s.clearQuery()

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
func (s *SQLSession) parseSQL(sql string, paramMap map[string]string) {
	// paramMapが未設定(nil or 要素0件)の場合は空のMapで初期化.
	if len(paramMap) == 0 {
		paramMap = map[string]string{}
	}

	// SQL文の整形を行う.
	sql = regexp.MustCompile(`(\r\n|\r|\n)`).ReplaceAllString(sql, "")
	sql = regexp.MustCompile(`( |\t)+`).ReplaceAllString(sql, " ")
	sql = strings.TrimSpace(sql)

	templ := template.Must(template.New("sql").Parse(sql))
	templ.Execute(s, paramMap)

	s.log(LogLevelDebug, "Parsed SQL: "+s.query)
}

// clearQuery クエリフィールドを初期化します.
func (session *SQLSession) clearQuery() {
	session.query = ""
}

// log ログ出力を行います.
func (s *SQLSession) log(level LogLevel, message string) {
	if s.logConfig == nil {
		return
	}
	if s.logConfig.Level > level {
		return
	}
	switch level {
	case LogLevelDebug:
		s.logConfig.Debug(message)
	case LogLevelInfo:
		s.logConfig.Info(message)
	case LogLevelWarn:
		s.logConfig.Warn(message)
	case LogLevelError:
		s.logConfig.Error(message)
	}
}

// DEBUGレベルのログ出力が有効かどうかを返す.
func (s *SQLSession) isLogEnableDebug() bool {
	return s.logConfig.Level <= LogLevelDebug
}

// INFOレベルのログ出力が有効かどうかを返す.
func (s *SQLSession) isLogEnableInfo() bool {
	return s.logConfig.Level <= LogLevelInfo
}

// WARNレベルのログ出力が有効かどうかを返す.
func (s *SQLSession) isLogEnableWarn() bool {
	return s.logConfig.Level <= LogLevelWarn
}

// ERRORレベルのログ出力が有効かどうかを返す.
func (s *SQLSession) isLogEnableError() bool {
	return s.logConfig.Level <= LogLevelError
}

// convFieldToMap 構造体のフィールドをマップに変換します.
func convFieldToMap(obj any) map[string]string {
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
			// TODO DBによりフォーマットを切り替える必要がある.
			fieldMap[fieldName] = val.Format("'2006-01-02 15:04:05.000'")
			continue
		}

		fieldMap[fieldName] = fmt.Sprintf("'%v'", fieldValue)

	}

	return fieldMap
}
