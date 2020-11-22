package tksql

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
