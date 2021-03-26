package mister_gadget_bot

import (
	"io"
	"time"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"cloud.google.com/go/storage"
	"github.com/giona69/http-helpers"
	"github.com/giona69/mister_gadget-commons"
	"google.golang.org/appengine/urlfetch"
)

func Filer(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone")
	id := r.FormValue("id")
	site := r.FormValue("site")

	c := appengine.NewContext(r)

	log.Debugf(c, "Filer START "+phone+"-"+id)

	helpers.Debugf(c, "Phone: "+phone)
	helpers.Debugf(c, "Id: "+id)
	helpers.Debugf(c, "Site: "+site)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	if _, err := r2gallery.Exec(phone, id, site); err != nil {
		log.Errorf(c, "Could not delete pictures %v: ", err)
	}

	processPictures(c, httpc, phone, id, site)

	log.Debugf(c, "Filer END")
	helpers.ReturnOkJSON(w, r, "gallery processed")
}

func SaveRenderedPage(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("filename")
	urls     := r.FormValue("url")

	c := appengine.NewContext(r)

	log.Debugf(c, "SaveBrandRendered START "+filename)

	helpers.Debugf(c, "filename: " + filename)
	helpers.Debugf(c, "urls: " + urls)

	urlr := "https://rendertron-187115.appspot.com/render/" + urls + "%3Fsitemap=true?wc-inject-shadydom=true"
	helpers.Debugf(c, "url: " + urlr)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	req, err := http.NewRequest("GET", urlr, nil)
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
		helpers.Errorf(w, r, http.StatusInternalServerError, "Error new request: %v", err)
		return
	}

	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting rendertron response: %v", err)
		helpers.Errorf(w, r, http.StatusInternalServerError, "Error getting rendertron response: %v", err)
		return
	}
	defer res.Body.Close()

	helpers.Debugf(c, "Response rendertron: %v", res.Status)
	helpers.Debugf(c, "Response rendertron: %v", res.StatusCode)
	helpers.Debugf(c, "Response rendertron: %v", res.Header)

	if res.StatusCode != 200 {
		log.Errorf(c, "Rendertron status not 200: %v", res)
		helpers.Errorf(w, r, http.StatusInternalServerError, "Rendertron status not 200: %v", err)
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf(c, "Error reading rendertron body: %v", err)
		helpers.Errorf(w, r, http.StatusInternalServerError, "Error reading rendertron body: %v", err)
		return
	}

	mister_gadget.DeleteFile(c, filename)
	if err != nil {
		log.Errorf(c, "Error deleting file: %v", err)
	}

	mister_gadget.SaveFile(c, filename, bytes.NewReader(body))
	if err != nil {
		log.Errorf(c, "Error saving file: %v", err)
		helpers.Errorf(w, r, http.StatusInternalServerError, "Error saving file: %v", err)
		return
	}

	mister_gadget.SetMemcache(c, filename, body)
	if err != nil {
		log.Errorf(c, "Error setting memcache: %v", err)
	}

	log.Debugf(c, "SaveBrandRendered END")
	helpers.ReturnOkJSON(w, r, "rendertron render saved")
}

//noinspection GoUnusedParameter
func Resizer(w http.ResponseWriter, r *http.Request) {
	urli := r.FormValue("url")
	typei := r.FormValue("typei")

	c := appengine.NewContext(r)

	log.Debugf(c, "Resizer START "+urli)

	helpers.Debugf(c, "Phone: "+urli)
	helpers.Debugf(c, "Id: "+typei)

	client, err := storage.NewClient(c)
	if err != nil {
		log.Errorf(c, "Error opening storage: %v", err)
		return
	}
	defer client.Close()

	name := urli
	helpers.Debugf(c, "Pic to copy location: ", name)

	re, err := client.Bucket(bucket).Object(name).NewReader(c)
	if err != nil {
		log.Errorf(c, "Error reading url image: ", err)
		return
	}

	body, err := ioutil.ReadAll(re)
	if err != nil {
		log.Errorf(c, "Error reading Brand site: ", err)
		return
	}
	if err := re.Close(); err != nil {
		log.Errorf(c, "Error closing file: ", err)
	}

	wr := client.Bucket(bucket).Object(name).NewWriter(c)
	defer wr.Close()
	wr.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	wr.ContentType = "image/jpeg"
	wr.CacheControl = "public, max-age=2592000"

	var ibytes []byte
	if typei == "MAIN" {
		ibytes = mister_gadget.ResizeThumb(body)
	} else {
		ibytes = mister_gadget.ResizePicture(body)
	}

	readbody := bytes.NewReader(ibytes)
	if _, err := io.Copy(wr, readbody); err != nil {
		log.Errorf(c, "Error reading resize image: ", err)
		return
	}

	log.Debugf(c, "Resizer END")
}

func processPictures(c context.Context, httpc *http.Client, phone string, BrandId string, site string) {
	row := lpic.QueryRow(phone, BrandId)
	var imain string
	if err := row.Scan(&imain); err != nil {
		log.Errorf(c, "Error scan main picture: %v ", err)
	} else {
		if !strings.Contains(imain, "storage.googleapis.com") {
			urlmain, err := mister_gadget.CopyPicture(c, httpc, imain, phone, BrandId, "MAIN")
			if err != nil {
				log.Errorf(c, "Error copying main picture: %v", err)
			} else {
				if _, err := upic.Exec(urlmain, phone, BrandId); err != nil {
					log.Errorf(c, "Could not update main pic url: %v: ", err)
				}
			}
		}
	}

	rows, err := sgallery.Query(phone, BrandId, site)
	if err != nil {
		log.Errorf(c, "Error querying pictures: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pic string
		var id string

		if err := rows.Scan(&pic, &id); err != nil {
			log.Errorf(c, "Error scan row picture: %v ", err)
			continue
		}

		urlp, err := mister_gadget.CopyPicture(c, httpc, pic, phone, BrandId, id)
		if err != nil {
			log.Errorf(c, "Error copying picture: %v", err)
			continue
		}

		if _, err := ugallery.Exec(urlp, phone, BrandId, site, id); err != nil {
			log.Errorf(c, "Could not update pic url: %v: ", err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error row.next pictures: %v", err)
	}
}
