package mister_gadget_bot

import (
	"fmt"
	"strings"
	"net/url"
	"net/http"
	"database/sql"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/runtime"
	"google.golang.org/appengine/taskqueue"
	"github.com/giona69/http-helpers"
	"github.com/giona69/mister_gadget-commons"
)


func CreateSitemap(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "create sitemap START")

		createAllSitemap(c)

		log.Debugf(c, "create sitemap END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch create sitemap process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Creating sitemap ...")
}

func createAllSitemap(c context.Context) {
	var sitemapf string

	sitemapf = createCitiesSitemap(c, sitemapf, false)
	sitemapf = createBrandSitemap(c, sitemapf, false)
	sitemapf = createPostSitemap(c, sitemapf, false)
	sitemapf = createFixedPages(c, sitemapf, false)

	mister_gadget.SaveFile(c, sitemapfile, strings.NewReader(sitemapf))
	if err != nil {
		log.Errorf(c, "Error saving sitemap: %v", err)
	}

	mister_gadget.SetMemcache(c, sitemapfile, []byte(sitemapf))
	if err != nil {
		log.Errorf(c, "Error setting memcache: %v", err)
	}
}

func createBrandSitemap(c context.Context, sitemape string, force bool) (sitemapf string) {

		log.Debugf(c, "create Brand sitemap START")

		rows, err := sitemapE.Query()
		if err != nil {
			log.Errorf(c,"Error in db.Query: %v", err)
			return sitemape
		}
		defer rows.Close()

		for rows.Next() {
			var city, phone, id, name string
			if err := rows.Scan(&city, &phone, &id, &name); err != nil {
				log.Errorf(c, "Error in rows.Scan: %v", err)
				continue
			}

			phone     = mister_gadget.FixPhone(phone)
			name      = url.QueryEscape(strings.ToLower(name))
			city      = url.QueryEscape(strings.ToLower(city))
			urls     := fmt.Sprintf("%s/Brand/%s/girl/%s/%s/%s", mainSite, city, phone, id, name)
			filename := rendered + "/girl" + phone + id + city

			sitemape += urls + "\n"

			if !mister_gadget.FileExists(c, filename) || force {
				postRenderTask(c, filename, urls)
			}
		}
		if err := rows.Err(); err != nil {
			log.Errorf(c, "Error in Row error: %v", err)
		}

		log.Debugf(c, "create Brand sitemap END")
		return sitemape
}

func createPostSitemap(c context.Context, sitemape string, force bool) (sitemapf string) {

	log.Debugf(c, "create POST sitemap START")

	rows, err := postE.Query()
	if err != nil {
		log.Errorf(c,"Error in db.Query: %v", err)
		return sitemape
	}
	defer rows.Close()

	for rows.Next() {
		var headline, id string
		var city, name, agency sql.NullString
		if err := rows.Scan(&city, &id, &headline, &name, &agency); err != nil {
			log.Errorf(c, "Error in rows.Scan: %v", err)
			continue
		}

		if city.Valid {
			city.String = url.QueryEscape(strings.ToLower(city.String))
		} else {
			city.String = "milano"
		}

		newTail      := url.QueryEscape(strings.ToLower(headline + " " + name.String + " " + agency.String))
		urls         := fmt.Sprintf("%s/news/%s/post/%s/%s", mainSite, city.String, id, newTail)
		filename     := rendered + "/post" + id

		sitemape += urls + "\n"

		if !mister_gadget.FileExists(c, filename) || force {
			postRenderTask(c, filename, urls)
		}
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error in Row error: %v", err)
	}

	log.Debugf(c, "create POST sitemap END")
	return sitemape
}

func createCitiesSitemap(c context.Context, sitemape string, force bool) (sitemapf string) {

	log.Debugf(c, "create CITIES sitemap START")

	rows, err := citiesE.Query()
	if err != nil {
		log.Errorf(c,"Error in db.Query: %v", err)
		return sitemape
	}
	defer rows.Close()

	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			log.Errorf(c, "Error in rows.Scan: %v", err)
			continue
		}

		city          = url.QueryEscape(strings.ToLower(city))

		urlNews            := fmt.Sprintf("%s/news/%s", mainSite, city)
		urlBrand          := fmt.Sprintf("%s/Brand/%s", mainSite, city)
		urlBrandStorage        := fmt.Sprintf("%s/Brand/%s/Storage", mainSite, city)
		urlBrandBbj       := fmt.Sprintf("%s/Brand/%s/bbj", mainSite, city)
		urlBrandSize      := fmt.Sprintf("%s/Brand/%s/Size", mainSite, city)
		urlBrandStorageBbj     := fmt.Sprintf("%s/Brand/%s/Storage+bbj", mainSite, city)
		urlBrandStorageBbjSize := fmt.Sprintf("%s/Brand/%s/Storage+bbj+Size", mainSite, city)

		filenameNews            := rendered + "/news" + city
		filenameBrand          := rendered + "/Brand" + city
		filenameBrandStorage        := rendered + "/BrandStorage" + city
		filenameBrandBbj       := rendered + "/Brandbbj" + city
		filenameBrandSize      := rendered + "/BrandSize" + city
		filenameBrandStorageBbj     := rendered + "/BrandStoragebbj" + city
		filenameBrandStorageBbjSize := rendered + "/BrandStoragebbjSize" + city

		sitemape += urlNews + "\n"
		sitemape += urlBrand + "\n"
		sitemape += urlBrandStorage + "\n"
		sitemape += urlBrandBbj + "\n"
		sitemape += urlBrandSize + "\n"
		sitemape += urlBrandStorageBbj + "\n"
		sitemape += urlBrandStorageBbjSize + "\n"

		if !mister_gadget.FileExists(c, filenameNews) || force {
			postRenderTask(c, filenameNews, urlNews)
		}

		if !mister_gadget.FileExists(c, filenameBrand) || force {
			postRenderTask(c, filenameBrand, urlBrand)
		}

		if !mister_gadget.FileExists(c, filenameBrandStorage) || force {
			postRenderTask(c, filenameBrandStorage, urlBrandStorage)
		}

		if !mister_gadget.FileExists(c, filenameBrandBbj) || force {
			postRenderTask(c, filenameBrandBbj, urlBrandBbj)
		}

		if !mister_gadget.FileExists(c, filenameBrandSize) || force {
			postRenderTask(c, filenameBrandSize, urlBrandSize)
		}

		if !mister_gadget.FileExists(c, filenameBrandStorageBbj) || force {
			postRenderTask(c, filenameBrandStorageBbj, urlBrandStorageBbj)
		}

		if !mister_gadget.FileExists(c, filenameBrandStorageBbjSize) || force {
			postRenderTask(c, filenameBrandStorageBbjSize, urlBrandStorageBbjSize)
		}
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error in Row error: %v", err)
	}

	log.Debugf(c, "create CITIES sitemap END")
	return sitemape
}

func createFixedPages(c context.Context, sitemape string, force bool) (sitemapf string) {

	log.Debugf(c, "create FIXED sitemap START")

	sitemape += "https://www.mister_gadget.net/news" + "\n"
	sitemape += "https://www.mister_gadget.net/Brand" + "\n"
	url404      := "https://www.mister_gadget.net/not-found"
	filename404 := rendered + "/page404"
	if !mister_gadget.FileExists(c, filename404) {
		postRenderTask(c, filename404, url404)
	}

	filenameName           := rendered + "/pagename"
	filenamePhone          := rendered + "/pagephone"
	filenameContacts       := rendered + "/pagecontacts"
	filenameLogin          := rendered + "/pagelogin"
	filenameLoginBrand    := rendered + "/pagelogin-Brand"
	filenameLoginBrandSms := rendered + "/pagelogin-Brand-sms"
	filenamePay            := rendered + "/pagepay"
	filenameVerify         := rendered + "/pageverify"
	filenamePassword       := rendered + "/pagepassword"

	urlPhone          := fmt.Sprintf("%s/phone", mainSite)
	urlName           := fmt.Sprintf("%s/name", mainSite)
	urlContacts       := fmt.Sprintf("%s/contacts", mainSite)
	urlLogin          := fmt.Sprintf("%s/login", mainSite)
	urlLoginBrand    := fmt.Sprintf("%s/login-Brand", mainSite)
	urlLoginBrandSms := fmt.Sprintf("%s/login-Brand-sms", mainSite)
	urlPay            := fmt.Sprintf("%s/pay", mainSite)
	urlVerify         := fmt.Sprintf("%s/verify", mainSite)
	urlPassword       := fmt.Sprintf("%s/password", mainSite)

	sitemape += urlContacts + "\n"
	sitemape += urlPhone + "\n"
	sitemape += urlName + "\n"
	sitemape += urlLogin + "\n"
	sitemape += urlLoginBrand + "\n"
	sitemape += urlLoginBrandSms + "\n"
	sitemape += urlPay + "\n"
	sitemape += urlVerify + "\n"
	sitemape += urlPassword + "\n"

	if !mister_gadget.FileExists(c, filenameContacts) || force {
		postRenderTask(c, filenameContacts, urlContacts)
	}

	if !mister_gadget.FileExists(c, filenamePhone) || force {
		postRenderTask(c, filenamePhone, urlPhone)
	}

	if !mister_gadget.FileExists(c, filenameName) || force {
		postRenderTask(c, filenameName, urlName)
	}

	if !mister_gadget.FileExists(c, filenameLogin) || force {
		postRenderTask(c, filenameLogin, urlLogin)
	}

	if !mister_gadget.FileExists(c, filenameLoginBrand) || force {
		postRenderTask(c, filenameLoginBrand, urlLoginBrand)
	}

	if !mister_gadget.FileExists(c, filenameLoginBrandSms) || force {
		postRenderTask(c, filenameLoginBrandSms, urlLoginBrandSms)
	}

	if !mister_gadget.FileExists(c, filenamePay) || force {
		postRenderTask(c, filenamePay, urlPay)
	}

	if !mister_gadget.FileExists(c, filenameVerify) || force {
		postRenderTask(c, filenameVerify, urlVerify)
	}

	if !mister_gadget.FileExists(c, filenamePassword) || force {
		postRenderTask(c, filenamePassword, urlPassword)
	}

	log.Debugf(c, "create FIXED sitemap END")
	return sitemape
}

func postRenderTask(c context.Context, filename string, urls string) {
	t := taskqueue.NewPOSTTask("/render", url.Values{"filename": {filename}, "url": {urls}})
	if _, err := taskqueue.Add(c, t, render); err != nil {
		log.Errorf(c, "Could not insert task: %v: ", err)
	}
}
