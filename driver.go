package tksql

// TODO
// 現状はmysqlのみの対応.
// 後々他のDBも対応予定.

// Driver Driver型.
type Driver int

const (
	//MYSQL MySQL.
	MYSQL Driver = iota
)

// Driver名を取得します.
func (d Driver) String() string {
	switch d {
	case MYSQL:
		return "mysql"
	default:
		return ""
	}
}
