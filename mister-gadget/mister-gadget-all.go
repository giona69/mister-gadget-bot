package mister_gadget_bot

import (
	"strings"
	"strconv"
	"net/url"
	"net/http"
	"database/sql"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/runtime"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
	"github.com/giona69/http-helpers"
	"github.com/giona69/mister_gadget-commons"
	"time"
)

func ParseAllBrands(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Parsing Brand STANDARD START")

		for i := 1; i <= 35; i++ {
			page := strconv.FormatInt(int64(i), 10)
			parseHuaweiListAjaxNoReg(c, "https://www.huawei.com/Brands/regular?ajax", "r", page)
		}

		parseSamsungList(c, "https://www.samsung.com/models/")
		parseSamsungList(c, "https://www.samsung.com/models/all/page:2/")
		parseSamsungList(c, "https://www.samsung.com/models/all/page:3/")
		parseSamsungList(c, "https://www.samsung.com/models/all/page:4/")
		parseSamsungList(c, "https://www.samsung.com/models/all/page:5/")

		if _, err := disactive.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive statement: %v: ", err)
		}

		if _, err := disacSite.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive by site statement: %v: ", err)
		}

		if _, err := phoneA.Exec(); err != nil {
			log.Errorf(c, "Could not run phone active statement: %v: ", err)
		}

		log.Debugf(c, "Parsing Brand STANDARD END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch parse Brand process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Parsing Brand Forum ...")
}

func ParseReinforce(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Parsing Brand REINFORCE START")

		parseSamsungList(c, "https://www.samsung.com/models/milano/")
		parseSamsungList(c, "https://www.samsung.com/models/milano/page:2/")
		parseSamsungList(c, "https://www.samsung.com/models/milano/page:3/")
		parseSamsungList(c, "https://www.samsung.com/models/milano/page:4/")
		parseSamsungList(c, "https://www.samsung.com/models/milano/page:5/")

		parseSamsungList(c, "https://www.samsung.com/models/roma/")

		parseHuaweiList(c, "https://www.huawei.com/")

		parseHuaweiList(c, "https://www.huawei.com/models")
		parseHuaweiListAjax(c, "https://www.huawei.com/models?ajax", "r", "0", "1")
		parseHuaweiListAjax(c, "https://www.huawei.com/models?ajax", "r", "1", "2")
		parseHuaweiListAjax(c, "https://www.huawei.com/models?ajax", "r", "1", "3")

		parseHuaweiList(c, "https://www.huawei.com/forms")
		parseHuaweiListAjax(c, "https://www.huawei.com/forms?ajax", "r", "0", "1")
		parseHuaweiListAjax(c, "https://www.huawei.com/forms?ajax", "r", "1", "2")
		parseHuaweiListAjax(c, "https://www.huawei.com/forms?ajax", "r", "1", "3")

		parseHuaweiList(c, "https://www.huawei.com/other/arg-other-2354")
		parseHuaweiList(c, "https://www.huawei.com/other/Brandofitaly-1369")
		parseHuaweiList(c, "https://www.huawei.com/other/Brands-in-italy-2433")
		parseHuaweiList(c, "https://www.huawei.com/other/tv-1600")

		if _, err := disactive.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive statement: %v: ", err)
		}

		if _, err := disacSite.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive by site statement: %v: ", err)
		}

		log.Debugf(c, "Parsing Brand REINFORCE END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch parse Brand process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Parsing Brand Forum ...")
}

func CheckReadyBrands(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Brand CHECK START")

		rows, err := listOrigin.Query()
		if err != nil {
			log.Errorf(c,"Error in db.Query: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var urle, ide string
			if err := rows.Scan(&urle, &ide); err != nil {
				log.Errorf(c, "Error in rows.Scan: %v", err)
				continue
			}

			if ide == "1" {
				helpers.Debugf(c, urle)
				parseHuaweiSingle(c, urle)
			} else if ide == "2" {
				helpers.Debugf(c, urle)
				parseSamsungSingle(c, urle)
			}
		}
		if err := rows.Err(); err != nil {
			log.Errorf(c, "Error in Row error: %v", err)
		}

		log.Debugf(c, "Brand CHECK END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch parse Brand process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Checking Brand Lists ...")
}

func OncePerDay(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "ONCE PER DAY START")

		if _, err := disactive.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive statement: %v: ", err)
		}

		if _, err := disacSite.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive by site statement: %v: ", err)
		}

		if _, err := phoneA.Exec(); err != nil {
			log.Errorf(c, "Could not run phone active statement: %v: ", err)
		}

		if _, err := uDuration.Exec(); err != nil {
			log.Errorf(c, "Could not run update duration statement: %v: ", err)
		}

		if err := decreaseCredit(c); err != nil {
			log.Errorf(c, "Could not run decrease credit statement: %v: ", err)
		}

		// comment to save money
		// createAllSitemap(c)

		log.Debugf(c, "ONCE PER DAY END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch once per day process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Once per day process ...")
}

func BonusPost(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "BONUS POST START")

		if err := bonusCreatePost(c); err != nil {
			log.Errorf(c, "Could not run bonus create post statement: %v: ", err)
		}

		log.Debugf(c, "BONUS POST END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch bonus post process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "BONUS process ...")
}

func decreaseCredit(c context.Context) (err error) {
	rows, err := cDuration.Query()
	if err != nil {
		log.Errorf(c,"Error in db.Query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user, id string
		if err := rows.Scan(&user, &id); err != nil {
			log.Errorf(c, "Error in rows.Scan: %v", err)
			continue
		}

		helpers.Debugf(c, "decreasing %v", user, id)

		var tx *sql.Tx
		if tx, err = db.Begin(); err != nil {
			log.Errorf(c, "Error in db.Begin: %v", err)
			tx.Rollback()
			continue
		}

		if _, err := tx.Stmt(decCredit).Exec(user); err != nil {
			log.Errorf(c, "Could not run decrease credit statement: %v: ", err)
			tx.Rollback()
			continue
		}

		if _, err := resetD.Exec(user, id); err != nil {
			log.Errorf(c, "Could not run decrease credit statement: %v: ", err)
			tx.Rollback()
			continue
		}

		tx.Commit()
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error in Row error: %v", err)
		return err
	}

	return nil
}

func bonusCreatePost(c context.Context) (err error) {
	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	rows, err := activeByU.Query()
	if err != nil {
		log.Errorf(c,"Error in db.Query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user, id string
		if err := rows.Scan(&user, &id); err != nil {
			log.Errorf(c, "Error in rows.Scan: %v", err)
			continue
		}

		helpers.Debugf(c, "generating post %v", user, id)

		var etu mister_gadget.Brand
		if err := getBrandSQL(c, user, id, &etu); err != nil {
			log.Errorf(c, "Brand not found %v", err)
			continue
		}

		mister_gadget.InsertPost(c, httpc, &etu, etu.Image.String, mister_gadget.PostNowActive)
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error in Row error: %v", err)
		return err
	}

	return nil
}

func ResizeAllPictures(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Resizing START")

		rows, err := agallery.Query()
		if err != nil {
			log.Errorf(c, "Error in db.Query: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var urlp string
			if err := rows.Scan(&urlp); err != nil {
				log.Errorf(c, "Error in rows.Scan: %v", err)
				continue
			}

			urlp = images + "/" + strings.TrimLeft(urlp, "https://storage.googleapis.com/mister_gadget-175215.appspot.com/" + images + "/")
			helpers.Debugf(c, "url: " + urlp)

			t := taskqueue.NewPOSTTask("/resizer", url.Values{"url": {urlp}, "typei": {"0"}})
			if _, err := taskqueue.Add(c, t, "resizer"); err != nil {
				log.Errorf(c, "Could not insert task: %v: ", err)
			}

		}
		if err := rows.Err(); err != nil {
			log.Errorf(c, "Error in Row error: %v", err)
		}

		log.Debugf(c, "Resizing END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch resizing process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Resize All Pictures ...")
}

func ResizeMainPictures(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Resizing MAIN START")

		rows, err := aBrand.Query()
		if err != nil {
			log.Errorf(c, "Error in db.Query: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var urlp string
			if err := rows.Scan(&urlp); err != nil {
				log.Errorf(c, "Error in rows.Scan: %v", err)
				continue
			}

			urlp = images + "/" + strings.TrimLeft(urlp, "https://storage.googleapis.com/mister_gadget-175215.appspot.com/" + images + "/")
			helpers.Debugf(c, "url: " + urlp)

			t := taskqueue.NewPOSTTask("/resizer", url.Values{"url": {urlp}, "typei": {"MAIN"}})
			if _, err := taskqueue.Add(c, t, "resizer"); err != nil {
				log.Errorf(c, "Could not insert task: %v: ", err)
			}

		}
		if err := rows.Err(); err != nil {
			log.Errorf(c, "Error in Row error: %v", err)
		}

		log.Debugf(c, "Resizing MAIN END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch resizing process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Resize Main Pictures ...")
}

func ResizePostPictures(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Resizing POST START")

		ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
		httpc := urlfetch.Client(ctxWithDeadline)

		rows, err := apost.Query()
		if err != nil {
			log.Errorf(c, "Error in db.Query: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var urlp, phone, ides string
			if err := rows.Scan(&urlp, &phone, &ides); err != nil {
				log.Errorf(c, "Error in rows.Scan: %v", err)
				continue
			}

			mister_gadget.UpdatePost(c, httpc, phone, ides, urlp)
		}
		if err := rows.Err(); err != nil {
			log.Errorf(c, "Error in Row error: %v", err)
		}

		log.Debugf(c, "Resizing POST END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch resizing process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Resize Post Pictures ...")
}

func ResizeOnePicture(w http.ResponseWriter, r *http.Request) {
	urlp := r.URL.Query().Get("url")

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {

		t := taskqueue.NewPOSTTask("/resizer", url.Values{"url": {urlp}, "typei": {"0"}})
		if _, err := taskqueue.Add(c, t, "resizer"); err != nil {
			log.Errorf(c, "Could not insert task: %v: ", err)
		}

		log.Debugf(c, "Resizing PICTURE")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch resizing process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Resize All Pictures ...")
}

