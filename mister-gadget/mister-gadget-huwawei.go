package mister_gadget_bot

import (
	"time"
	"regexp"
	"strconv"
	"strings"
	"net/url"
	"io/ioutil"
	"net/http"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/runtime"
	"google.golang.org/appengine/urlfetch"
	"github.com/giona69/http-helpers"
	"github.com/giona69/mister_gadget-commons"
)

//noinspection GoSnakeCaseUsage
type HuaweiStruct struct {
	Filter_body	string `json:"filter_body"`
	Brand_list string `json:"Brand_list"`
}

func ParseHuawei(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Parsing Brand Forum START")

		for i := 1; i <= 35; i++ {
			page := strconv.FormatInt(int64(i), 10)
			parseHuaweiListAjaxNoReg(c, "https://www.huawei.com/Brands/regular?ajax", "r", page)
		}

		if _, err := disactive.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive statement: %v: ", err)
		}

		if _, err := disacSite.Exec(); err != nil {
			log.Errorf(c, "Could not run disactive by site statement: %v: ", err)
		}

		if _, err := phoneA.Exec(); err != nil {
			log.Errorf(c, "Could not run phone active statement: %v: ", err)
		}

		log.Debugf(c, "Parsing Brand Forum END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch parse Brand forum process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Parsing Brand Forum ...")
}

func ParseHuaweiSingle(w http.ResponseWriter, r *http.Request) {
	urlp := r.URL.Query().Get("url")

	helpers.SetHeaderJSON(w)

	c := appengine.NewContext(r)

	log.Debugf(c, "Parsing Brand Forum SINGLE START")
	log.Debugf(c, "Url: " + urlp)

	Brand := parseHuaweiSingle(c, HuaweiURL + urlp)

	log.Debugf(c, "Parsing Brand Forum SINGLE END")

	helpers.ReturnJSON(w, r, Brand)
}

func ParseHuaweiSingleToScreen(w http.ResponseWriter, r *http.Request) {
	urlp := r.URL.Query().Get("url")

	helpers.SetHeaderJSON(w)

	c := appengine.NewContext(r)

	helpers.Debugf(c, "Parsing Brand Forum START")
	helpers.Debugf(c, "Url: " + urlp)
	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	parseHuaweiSingleToScreen(c, httpc, w, r, urlp)


	helpers.Debugf(c, "Parsing Brand Forum END")
}

func ParseHuaweiAjaxToScreen(w http.ResponseWriter, r *http.Request) {
	sort := r.URL.Query().Get("sort")
	reg  := r.URL.Query().Get("reg")
	page := r.URL.Query().Get("page")
	urlp := "https://www.huawei.com/models?ajax"

	helpers.SetHeaderJSON(w)

	c := appengine.NewContext(r)

	var eb HuaweiStruct
	helpers.Debugf(c, "Parsing Brand Ajax List START " + urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	form := url.Values{}
	form.Add("sort", sort)
	form.Add("reg", reg)
	form.Add("page", page)
	req, err := http.NewRequest("POST", urlp, strings.NewReader(form.Encode()))
	req.PostForm = form
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	cookie := http.Cookie{Name: "18", Value: "1"}
	req.AddCookie(&cookie)
	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting Brand list: %v", err)
		return
	}

	if err := helpers.DecodeJSON(res.Body, &eb); err != nil {
		log.Errorf(c, "cannot decode json: %v", err)
		return
	}

	body, err := ioutil.ReadAll(strings.NewReader(eb.Brand_list))
	res.Body.Close()
	if err != nil {
		log.Errorf(c, "Error reading ajax template: ", err)
		return
	}

	w.Write(body)

	helpers.Debugf(c, "Parsing Brand Forum Ajax END")
}

func parseHuaweiList(c context.Context, urlp string) {
	tParsed := 0
	log.Debugf(c, "Parsing Brand Forum List " + urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)
	res, err := httpc.Get(urlp)
	if err != nil {
		log.Errorf(c, "Error getting Brand forum HP: ", err)
		return
	}
	defer res.Body.Close()

	z := html.NewTokenizer(res.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			helpers.Debugf(c, "Parsing Brand Forum END")
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if isDiv := t.Data == "div"; isDiv {
				for _, pw := range t.Attr {
					if pw.Key == "class" && (pw.Val == "photo_wrap" || pw.Val == "gotd-wrp") {
						helpers.Debugf(c, "Found class: %v", pw)
						if pw.Val == "gotd-wrp" {
							z.Next(); z.Next()
						}
						z.Next(); z.Next();	at := z.Token()
						helpers.Debugf(c, "Tag a: %v", at)

						for _, a := range at.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found href: %v", a)
								parseHuaweiSingle(c, HuaweiURL + a.Val)
								tParsed++
								break
							}
						}
						break
					}
				}
			}
		}
	}

	log.Debugf(c, "Parsing Brand Forum List: TOTAL %d", tParsed)
}

func parseHuaweiListAjax(c context.Context, urlp string, sort string, reg string, page string) {
	var eb HuaweiStruct
	log.Debugf(c, "Parsing Brand Forum Ajax List " + urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	form := url.Values{}
	form.Add("sort", sort)
	form.Add("reg", reg)
	form.Add("page", page)
	req, err := http.NewRequest("POST", urlp, strings.NewReader(form.Encode()))
	req.PostForm = form
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	cookie := http.Cookie{Name: "18", Value: "1"}
	req.AddCookie(&cookie)
	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting Brand list: %v", err)
		return
	}
	defer res.Body.Close()

	if err := helpers.DecodeJSON(res.Body, &eb); err != nil {
		log.Errorf(c, "cannot decode json: %v", err)
		return
	}

	z := html.NewTokenizer(strings.NewReader(eb.Brand_list))

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			helpers.Debugf(c, "Parsing Brand Forum Ajax Token END")
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if isDiv := t.Data == "div"; isDiv {
				for _, pw := range t.Attr {
					if pw.Key == "class" && (pw.Val == "photo_wrap" || pw.Val == "gotd-wrp") {
						helpers.Debugf(c, "Found class: %v", pw)
						if pw.Val == "gotd-wrp" {
							z.Next(); z.Next()
						}
						z.Next(); z.Next();	at := z.Token()
						helpers.Debugf(c, "Tag a: %v", at)

						for _, a := range at.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found href: %v", a)
								parseHuaweiSingle(c, HuaweiURL + a.Val)
								break
							}
						}
						break
					}
				}
			}
		}
	}

	helpers.Debugf(c, "Parsing Brand Forum Ajax END")
}

func parseHuaweiListAjaxNoReg(c context.Context, urlp string, sort string, page string) {
	tParsed := 0
	var eb HuaweiStruct
	log.Debugf(c, "Parsing Brand Ajax Forum List " + urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	form := url.Values{}
	form.Add("sort", sort)
	form.Add("page", page)
	req, err := http.NewRequest("POST", urlp, strings.NewReader(form.Encode()))
	req.PostForm = form
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	cookie := http.Cookie{Name: "18", Value: "1"}
	req.AddCookie(&cookie)
	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting Brand list: %v", err)
		return
	}
	defer res.Body.Close()

	if err := helpers.DecodeJSON(res.Body, &eb); err != nil {
		log.Errorf(c, "cannot decode json: %v", err)
		return
	}

	z := html.NewTokenizer(strings.NewReader(eb.Brand_list))

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			helpers.Debugf(c, "Parsing Brand Forum Ajax Token END")
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if isDiv := t.Data == "div"; isDiv {
				for _, pw := range t.Attr {
					if pw.Key == "class" && (pw.Val == "photo_wrap" || pw.Val == "gotd-wrp") {
						helpers.Debugf(c, "Found class: %v", pw)
						if pw.Val == "gotd-wrp" {
							z.Next(); z.Next()
						}
						z.Next(); z.Next();	at := z.Token()
						helpers.Debugf(c, "Tag a: %v", at)

						for _, a := range at.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found href: %v", a)
								parseHuaweiSingle(c, HuaweiURL + a.Val)
								tParsed++
								break
							}
						}
						break
					}
				}
			}
		}
	}

	log.Debugf(c, "Parsing Brand Forum Ajax: TOTAL %d", tParsed)
}

func parseHuaweiSingle(c context.Context, urlp string) (Brand *mister_gadget.Brand) {
	var BrandData mister_gadget.Brand
	var gallery []string
	var idBrand = "n/a"

	log.Debugf(c, "Parsing Brand Forum url " + urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	req, err := http.NewRequest("GET", urlp, nil)
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
	}

	cookie := http.Cookie{Name: "18", Value: "1"}
	req.AddCookie(&cookie)
	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting Brand href: %v %v", urlp, err)
		return
	}
	defer res.Body.Close()

	helpers.Debugf(c, "Response Brand site: %v", res.Status)
	helpers.Debugf(c, "Response Brand site: %v", res.StatusCode)
	helpers.Debugf(c, "Response Brand site: %v", res.Header)
	helpers.Debugf(c, "Response Brand site: %v", res.Body)

	z  := html.NewTokenizer(res.Body)
	tt := html.StartTagToken

	for tt != html.ErrorToken {
		tt = z.Next()

		switch {
		case tt == html.ErrorToken:
		case tt == html.SelfClosingTagToken:
			t := z.Token()
			if isInput := t.Data == "input"; isInput {
				id    := "n/a"
				value := "n/a"
				for _, itag := range t.Attr {
					if itag.Key == "id" {
						helpers.Debugf(c, "Found input tag: %v", itag)
						id = itag.Val
					}
					if itag.Key == "value" {
						helpers.Debugf(c, "Found input tag value: %v", itag.Val)
						value = itag.Val
					}
				}
				if id == "Brand_id" {
					idBrand = strings.TrimSpace(value)
				}
			}
		case tt == html.StartTagToken:
			t := z.Token()

			if isDiv := t.Data == "div"; isDiv {
				for _, nd := range t.Attr {
					if nd.Key == "class" && nd.Val == "head info" {
						helpers.Debugf(c, "Found class head info: %v", nd)
						z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag name: %v", t)
						BrandData.Name = t.Data
					}
					if nd.Key == "class" && nd.Val == "about-text" {
						helpers.Debugf(c, "Found class about-text: %v", nd)
						BrandData.Text.Valid  = true
						for tt != html.ErrorToken && (tt != html.EndTagToken || t.Data != "div") {
							tt = z.Next()
							raw := z.Raw()
							t = z.Token()
							helpers.Debugf(c, "Tag type: %v", t.Type)
							helpers.Debugf(c, "Raw text: %v", string(raw[:]))
							helpers.Debugf(c, "Tag text: %v", t)
							helpers.Debugf(c, "Tag data: %v", t.Data)
							if t.Type == html.TextToken {
								BrandData.Text.String += t.Data
							}
						}
					}
					if nd.Key == "class" && nd.Val == "w400" {
						helpers.Debugf(c, "Found class w400: %v", nd)
						z.Next(); z.Next(); t = z.Token()
						for _, a := range t.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found w400 href: %v", a)
								gallery = append(gallery, a.Val)
								if !BrandData.Image.Valid {
									BrandData.Image.String = a.Val
									BrandData.Image.Valid = true
								}
								break
							}
						}
					}
					if nd.Key == "class" && nd.Val == "w500" {
						helpers.Debugf(c, "Found class w500: %v", nd)
						z.Next(); z.Next(); t = z.Token()
						for _, a := range t.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found w500 href: %v", a)
								gallery = append(gallery, a.Val)
								break
							}
						}
					}
				}
			}
			if isTh := t.Data == "th"; isTh {
				z.Next(); t = z.Token()
				if t.Data == "Città base:" {
					z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag city: %v", t)
					BrandData.City = strings.TrimSpace(t.Data)
				} else if t.Data == "Zone città:" {
					z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag area: %v", t)
					BrandData.Area.String = strings.TrimSpace(t.Data)
					BrandData.Area.Valid = true
				} else if t.Data == "Telefono:" {
					z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag phone: %v", t)
					re := regexp.MustCompile("[^0-9]+")
					if t.Data[0:1] == "+" {
						BrandData.Phone = "+" + re.ReplaceAllString(t.Data,"")
					} else {
						BrandData.Phone = "+39" + re.ReplaceAllString(t.Data,"")
					}
				} else if t.Data == "Etnia:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag etnia: %v", t)
					BrandData.Details.String = t.Data
					BrandData.Details.Valid  = true
				} else if t.Data == "Nazionalità:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Nazionalità: %v", t)
					BrandData.Screen.String = t.Data
					BrandData.Screen.Valid    = true
				} else if t.Data == "Età:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Età: %v", t)
					BrandData.Age.Int64, _ = strconv.ParseInt(t.Data, 10, 64)
					BrandData.Age.Valid    = true
				} else if t.Data == "Occhi:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Occhi: %v", t)
					BrandData.System.String = t.Data
					BrandData.System.Valid  = true
				} else if t.Data == "Altezza:" {
					z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Altezza: %v", t)
					BrandData.Height.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " cm"), 10, 64)
					BrandData.Height.Valid    = true
				} else if t.Data == "Fumatore/rice:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Fumatore: %v", t)
					if t.Data == "No" {
						BrandData.Height.Bool = false
					} else {
						BrandData.Height.Bool = true
					}
					BrandData.Height.Valid  = true
				} else if t.Data == "Bere:" {
					z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Bere: %v", t)
					if t.Data == "No" {
						BrandData.Battery.Bool = false
					} else {
						BrandData.Battery.Bool = true
					}
					BrandData.Battery.Valid  = true
				}
			}
		}
	}

	if !(BrandData.Phone != "" && idBrand != "") {
		log.Debugf(c, "Phone & Id not found")
		return &BrandData
	}

	resps, err := httpc.Get("https://www.huawei.com/it/Brands/ajax-get-personal-site?Brand_id="+idBrand+"&has_rates=0&has_services=1")
	if err != nil {
		helpers.Debugf(c, "Error getting ajax personal site: %v", err)
	} else {
		defer resps.Body.Close()
		z = html.NewTokenizer(resps.Body)
		tt = html.StartTagToken
		for tt != html.ErrorToken {
			tt = z.Next()

			switch {
			case tt == html.ErrorToken:
				// End of the document, we're done
				break
			case tt == html.StartTagToken:
				t := z.Token()

				if isTh := t.Data == "th"; isTh {
					z.Next(); t = z.Token()
					if t.Data == "Personal site:" {
						z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag Personal: %v", t)
						for _, a := range t.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found href: %v", a)
								BrandData.Detail_site.String = a.Val
								BrandData.Detail_site.Valid = true
								break
							}
						}
					}
				}
			}
		}
	}

	resds, err := httpc.Get(BrandData.Detail_site.String)
	if err != nil {
		helpers.Debugf(c, "Error getting Brand services site: %v", err)
	} else {
		defer resds.Body.Close()
		z = html.NewTokenizer(resds.Body)
		tt = html.StartTagToken
		for tt != html.ErrorToken {
			tt = z.Next()

			switch {
			case tt == html.ErrorToken:
				// End of the document, we're done
				break
			case tt == html.StartTagToken:
				t := z.Token()

				if isLi := t.Data == "li"; isLi {
					z.Next(); t = z.Token()
					if t.Data == "Cum in Mouth" {
						BrandData.Earphones.Bool  = true
						BrandData.Earphones.Valid = true
					} else if strings.Contains(t.Data, "Cum in Mouth") {
						BrandData.Earphones.Bool  = true
						BrandData.Earphones.Valid = true
						BrandData.EarphonesExtra.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " Cum in Mouth / EUR"), 10, 64)
						BrandData.EarphonesExtra.Valid = true
					} else if t.Data == "Blowjob without Condom" {
						BrandData.Ram.Bool  = true
						BrandData.Ram.Valid = true
					} else if strings.Contains(t.Data, "Blowjob without Condom") && t.Data != "Blowjob without Condom to Completion" {
						BrandData.Ram.Bool  = true
						BrandData.Ram.Valid = true
						BrandData.BatteryExtra.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " Blowjob without Condom / EUR"), 10, 64)
						BrandData.BatteryExtra.Valid = true
					} else if t.Data == "Cum on Face" {
						BrandData.Rom.Bool  = true
						BrandData.Rom.Valid = true
					} else if strings.Contains(t.Data, "Cum on Face") {
						BrandData.Rom.Bool  = true
						BrandData.Rom.Valid = true
						BrandData.RomExtra.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " Cum on Face / EUR"), 10, 64)
						BrandData.RomExtra.Valid = true
					} else if t.Data == "Kissing if good chemistry" {
						BrandData.Cables.Bool  = true
						BrandData.Cables.Valid = true
					} else if t.Data == "Kissing" {
						BrandData.Storage.Bool  = true
						BrandData.Storage.Valid = true
					} else if t.Data == "French Kissing" {
						BrandData.Storage.Bool  = true
						BrandData.Storage.Valid = true
					} else if strings.Contains(t.Data, "French Kissing") {
						BrandData.Storage.Bool  = true
						BrandData.Storage.Valid = true
						BrandData.StorageExtra.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " French Kissing / EUR"), 10, 64)
						BrandData.StorageExtra.Valid = true
					} else if t.Data == "Size Sex" {
						BrandData.Size.Bool  = true
						BrandData.Size.Valid = true
					} else if strings.Contains(t.Data, "Size Sex") {
						BrandData.Size.Bool  = true
						BrandData.Size.Valid = true
						BrandData.SizeExtra.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " Size Sex / EUR"), 10, 64)
						BrandData.SizeExtra.Valid = true
					} else if t.Data == "Ball Licking and Sucking" {
						BrandData.Weight.Bool  = true
						BrandData.Weight.Valid = true
					} else if t.Data == "Girlfriend Experience (Material)" {
						BrandData.Material.Bool  = true
						BrandData.Material.Valid = true
					} else if t.Data == "Pornstar Experience (Gsm)" {
						BrandData.Gsm.Bool  = true
						BrandData.Gsm.Valid = true
					} else if t.Data == "Colors/uniforms" {
						BrandData.Colors.Bool  = true
						BrandData.Colors.Valid = true
					}
				}

				if isTh := t.Data == "th"; isTh {
					z.Next(); t = z.Token()
					if t.Data == "30 minute" {
						z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag 30: %v", t)
						if !BrandData.P30Min.Valid {
							BrandData.P30Min.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P30Min.Valid = true
						} else {
							BrandData.P30MinOut.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P30MinOut.Valid = true
						}
					} else if t.Data == "40 minute" {
						z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag 40: %v", t)
						BrandData.P30Min.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
						BrandData.P30Min.Valid    = true
					} else if t.Data == "60 minute" {
						z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag 60: %v", t)
						if !BrandData.P60Min.Valid {
							BrandData.P60Min.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P60Min.Valid = true
						} else {
							BrandData.P60MinOut.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P60MinOut.Valid = true
						}
					} else if t.Data == "1 hour" {
						z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
						helpers.Debugf(c, "Tag 1 hour: %v", t)
						if !BrandData.P60Min.Valid {
							BrandData.P60Min.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P60Min.Valid = true
						} else {
							BrandData.P60MinOut.Int64, _ = strconv.ParseInt(strings.Trim(t.Data, " EUR"), 10, 64)
							BrandData.P60MinOut.Valid = true
						}
					}
				}
			}
		}
	}

	BrandData.Country = "italy"
	BrandData.Id = idBrand
	helpers.Debugf(c, "Brand: %v", BrandData)
	helpers.Debugf(c, "Brand id: %v", idBrand)
	helpers.Debugf(c, "Brand gallery: %v", gallery)

	if BrandData.Image.Valid && BrandData.Phone != "" && BrandData.Id != "" && BrandData.City != "" {
		insertBrand(c, HuaweiSite, httpc, urlp, BrandData, gallery)
	} else {
		log.Debugf(c, "Brand phone, id, city or image empty")
	}

	return &BrandData
}

//noinspection GoUnusedParameter
func parseHuaweiSingleToScreen(c context.Context, httpc *http.Client, w http.ResponseWriter, r *http.Request, href string) {
	urlp := "https://www.huawei.com" + href

	req, err := http.NewRequest("GET", urlp, nil)
	if err != nil {
		log.Errorf(c, "Error new request: %v", err)
	}

	cookie := http.Cookie{Name: "18", Value: "1"}
	req.AddCookie(&cookie)
	res, err := httpc.Do(req)
	if err != nil {
		log.Errorf(c, "Error getting Brand href: %v %v", href, err)
		return
	}
	defer res.Body.Close()

	//res, err := httpc.Get("https://www.huawei.com" + href)
	helpers.Debugf(c, "Response Brand site: %v", res.Status)
	helpers.Debugf(c, "Response Brand site: %v", res.StatusCode)
	helpers.Debugf(c, "Response Brand site: %v", res.Header)
	helpers.Debugf(c, "Response Brand site: %v", res.Body)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Errorf(c, "Error reading Brand site: ", err)
		return
	}

	w.Write(body)
}
