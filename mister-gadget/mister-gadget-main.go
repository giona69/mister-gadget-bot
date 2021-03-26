package mister_gadget_bot

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"time"
)

//noinspection GoSnakeCaseUsage
type Phones struct {
	Phone        string
	Id           string
	Phone_active time.Time
}

const BrandForumSite = 1
const SamsungSite = 2
const SamsungURL = "https://www.samsung.com"
const BrandForumURL = "https://www.huawei.com"

var db *sql.DB
var err error
var images string
var bucket string
//noinspection ALL
var env_filer string
var filer string
var render string
var rendered string
var sitemapfile string
var mainSite string
var insert *sql.Stmt
var update *sql.Stmt
var agallery *sql.Stmt
var apost *sql.Stmt
var aBrand *sql.Stmt
var sgallery *sql.Stmt
var ugallery *sql.Stmt
var igallery *sql.Stmt
var u2gallery *sql.Stmt
var r2gallery *sql.Stmt
var iorig *sql.Stmt
var uorig *sql.Stmt
var u2orig *sql.Stmt
var phoneA *sql.Stmt
var disactive *sql.Stmt
var disacSite *sql.Stmt
var uDuration *sql.Stmt
var cDuration *sql.Stmt
var decCredit *sql.Stmt
var resetD *sql.Stmt
var listOrigin *sql.Stmt
var getBrand *sql.Stmt
var getBrandByPhoneOnly *sql.Stmt
var lphones *sql.Stmt
var uphones *sql.Stmt
var lpic *sql.Stmt
var upic *sql.Stmt
var activeByU *sql.Stmt
var sitemapE *sql.Stmt
var postE *sql.Stmt
var citiesE *sql.Stmt
var twilioQ *sql.Stmt
var twilioI *sql.Stmt

func InitAPI() {
	connectionName := os.Getenv("CLOUDSQL_CONNECTION_NAME")
	user           := os.Getenv("CLOUDSQL_USER")
	password       := os.Getenv("CLOUDSQL_PASSWORD")
	database       := os.Getenv("CLOUDSQL_DATABASE")
	images          = os.Getenv("CLOUDSQL_IMAGES")
	bucket          = os.Getenv("BUCKET")
	rendered        = os.Getenv("RENDERED")
	sitemapfile     = os.Getenv("SITEMAP")
	env_filer       = os.Getenv("HMI_ENV")
	mainSite        = os.Getenv("SITE")

	if env_filer == "prod" {
		filer = "filer"
		render = "render"
	} else {
		filer = "filerTEST"
		render = "renderTEST"
	}

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@cloudsql(%s)/%s?parseTime=true", user, password, connectionName, database))
	if err != nil {
		return
	}

	//noinspection SyntaxError
	insertSQL := "INSERT INTO Brand (phone, id, description, image, city, area, country, detail_site, height, age, Battery, Height,"
	insertSQL += "Screen, Details, System, text, Earphones, Rom, Ram, Storage, Cables, Size, Weight, Material,"
	insertSQL += "Gsm, Colors, p30min, p60min, p30minout, p60minout, BatteryExtra, Earphonesextra, Romextra, Storageextra, Sizeextra, "
	insertSQL += "created, updated, active, active_by_site, reviews_site, activated, phone_active) "
	insertSQL += "VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,true,true,?,?,NOW())"
	insert, err    = db.Prepare(insertSQL)
	agallery, err  = db.Prepare("SELECT url FROM Brand_gallery")
	aBrand, err   = db.Prepare("SELECT image FROM Brand")
	apost, err     = db.Prepare("SELECT image, phone_Brand, id_Brand FROM post")
	sgallery, err  = db.Prepare("SELECT origin_url, id FROM Brand_gallery WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=? AND url IS NULL")
	u2gallery, err = db.Prepare("UPDATE Brand_gallery SET id_Brand=? WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=?")
	r2gallery, err = db.Prepare("DELETE FROM Brand_gallery WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=? AND origin_url is null")
	ugallery, err  = db.Prepare("UPDATE Brand_gallery SET url=? WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=? AND id=?")
	igallery, err  = db.Prepare("INSERT INTO Brand_gallery (phone_Brand, id_Brand, id_origin_site, origin_url) VALUES (?, ?, ?, ?)")
	iorig, err     = db.Prepare("INSERT INTO Brand_origin_site (phone_Brand, id_Brand, id_origin_site, url) VALUES (?, ?, ?, ?)")
	uorig, err     = db.Prepare("UPDATE Brand_origin_site SET url=? WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=?")
	u2orig, err    = db.Prepare("UPDATE Brand_origin_site SET id_Brand=? WHERE phone_Brand=? AND id_Brand=? AND id_origin_site=?")
	lphones, err   = db.Prepare("SELECT phone_Brand, id_Brand FROM Brand_phone WHERE phone=?")
	uphones, err   = db.Prepare("UPDATE Brand_phone SET phone_active=? WHERE phone=?")
	lpic, err      = db.Prepare("SELECT image FROM Brand WHERE phone=? AND id=?")
	upic, err      = db.Prepare("UPDATE Brand SET image=? WHERE phone=? AND id=?")

	phoneA, err    = db.Prepare("UPDATE Brand SET phone_active = NOW() WHERE Brand.active = TRUE AND active_by_user = TRUE")
	disactive, err = db.Prepare("UPDATE Brand SET active = FALSE, active_by_site = FALSE WHERE updated <  NOW() - INTERVAL 1 DAY AND active = TRUE AND active_by_site = TRUE AND active_by_user = FALSE")
	disacSite, err = db.Prepare("UPDATE Brand SET active_by_site = FALSE WHERE updated <  NOW() - INTERVAL 1 DAY AND active_by_site = TRUE")
	uDuration, err = db.Prepare("UPDATE Brand SET duration=duration+TIMESTAMPDIFF(HOUR, pay_enabled, NOW()), pay_enabled=NOW() WHERE active_by_user = TRUE")
	cDuration, err = db.Prepare("SELECT user.email, Brand.id from user, Brand WHERE user.email = Brand.phone AND duration > 1")
	activeByU, err = db.Prepare("SELECT user.email, Brand.id from user, Brand WHERE user.email = Brand.phone AND Brand.active_by_user = TRUE")
	decCredit, err = db.Prepare("UPDATE user SET credit=credit-1 WHERE user.email = ?")
	decCredit, err = db.Prepare("UPDATE user SET credit=IF(credit>0, credit-1, 0) WHERE user.email = ?")
	resetD, err    = db.Prepare("UPDATE Brand SET duration=0 WHERE Brand.phone = ? AND Brand.id = ?")

	sitemapE, err  = db.Prepare("SELECT city, phone, id, description FROM Brand where phone != '+41765135635' and blacklist != TRUE")
	postE, err     = db.Prepare("SELECT post.city, post.id, post.headline, Brand.description, agency.description FROM post LEFT JOIN Brand ON post.phone_Brand = Brand.phone AND post.id_Brand = Brand.id LEFT JOIN agency ON post.phone_Brand = agency.phone")
	citiesE, err   = db.Prepare("SELECT DISTINCT city FROM Brand WHERE Brand.city != '' ORDER BY city")

	twilioQ, err  = db.Prepare("SELECT phone, Screen FROM Brand where active = TRUE AND phone != '+39123321123321' and Brand.phone like '%+39%' and phone not in (select phone from Brand_sms) LIMIT 25")
	twilioI, err  = db.Prepare("INSERT INTO Brand_sms (phone, date) VALUES (?, NOW())")

	//noinspection SyntaxError
	updateSQL := "UPDATE Brand SET id=?, description=?, image=?, city=?, area=?, country=?, detail_site=?, height=?, age=?, Battery=?, Height=?,"
	updateSQL += "Screen=?, Details=?, System=?, text=?, Earphones=?, Rom=?, Ram=?, Storage=?, Cables=?, Size=?, Weight=?, Material=?,"
	updateSQL += "Gsm=?, Colors=?, p30min=?, p60min=?, p30minout=?, p60minout=?, BatteryExtra=?, Earphonesextra=?, Romextra=?, Storageextra=?, Sizeextra=?, "
	updateSQL += "updated=?, active=true, active_by_site=true, reviews_site=?, activated=?, phone_active=? "
	updateSQL += "WHERE phone=? AND id=?"
	update, err = db.Prepare(updateSQL)

	listOrigin, err = db.Prepare("SELECT url, id_origin_site FROM Brand_origin_site")

	selectSQL := "SELECT Brand.phone, id, Brand.description, image, city, area, Brand.country, detail_site, height, age, Battery, Height,"
	selectSQL += "Screen, Details, System, text, Earphones, Rom, Ram, Storage, Cables, Size, Weight, Material,"
	selectSQL += "Gsm, Colors, p30min, p60min, p30minout, p60minout, BatteryExtra, Earphonesextra, Romextra, Storageextra, Sizeextra, "
	selectSQL += "created, updated, active, reviews_site, agency.description, agency.site, (created >  NOW() - INTERVAL 7 DAY) as new, activated, phone_active "
	selectSQL += "FROM Brand LEFT JOIN agency ON Brand.phone = agency.phone "
	BrandSQL := selectSQL + "WHERE Brand.phone=? AND id=?"
	getBrand, err = db.Prepare(BrandSQL)

	BrandBPOSQL := selectSQL + "WHERE Brand.phone=?"
	getBrandByPhoneOnly, err = db.Prepare(BrandBPOSQL)
}
