package browsers

//type Chromium struct {
//	MasterKey []byte
//}

type Chromium struct {
	Name        string
	Storage     string
	ProfilePath string
	MasterKey   []byte
	DataTypes   []int
	Paths       map[int]string
}

type Gecko struct {
	MasterKey []byte
}

type Browser struct {
	User          string
	BrowserType   string
	ProfileParent string
	Profiles      []string
}

type Profile struct {
	Name    string
	Path    string
	Browser Browser

	Logins      []Login
	Cookies     []Cookie
	CreditCards []CreditCard
	Downloads   []Download
	History     []History
}

type Login struct {
	Username string
	Password string
	LoginURL string
}

type Cookie struct {
	Host       string
	Name       string
	Path       string
	Value      string
	ExpireDate int64
}

type CreditCard struct {
	GUID            string
	Name            string
	ExpirationYear  string
	ExpirationMonth string
	Number          string
	Address         string
	Nickname        string
}

type Download struct {
	TargetPath string
	URL        string
}

type History struct {
	Title         string
	URL           string
	VisitCount    int
	LastVisitTime int64
}
