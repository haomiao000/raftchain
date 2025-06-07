package ext

type MySQLConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	DBName    string
	Charset   string
	ParseTime bool
	Loc       string
}

var (
	MySQL MySQLConfig
)
