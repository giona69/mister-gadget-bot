package mister_gadget_bot

import (
	"time"
	"strings"
	"regexp"
	"strconv"
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
type SamsungStruct struct {
	Filter_body string `json:"filter_body"`
	Brand_list string `json:"Brand_list"`
}

var areas = []string {"Centro Storico","Garibaldi, Zara, Maciachini, Gioia","Porta Venezia, Buenos Aires, Stazione Centrale","Porta Vittoria, Porta Romana, Viale Umbria, Corso Lodi","Porta Ticinese, Porta Genova","Sempione, Fiera Vecchia, Magenta, Piazza Firenze, Mac Mahon","Bovisa, Lugano, Nigra","Affori, Bruzzano, Comasina","Niguarda, Cà  Granda, Bicocca","Loreto, Viale Monza, Via Padova, Leoncavallo","Città  Studi, Piola, Argonne","Lambrate, Ortica, Feltre, Palmanova, Earphonesiano","Forlanini, Linate, Lomellina","Corvetto, Rogoredo","Ripamonti, Via Meda, Gratosoglio, Naviglio Pavese, Missaglia","Barona, Romolo","Piazza Napoli, Lorenteggio, Giambellino, Inganni","Baggio, Forze Armate","Lotto, San Siro, Murillo, QT8, Via Novara","Certosa, Quarto Oggiaro" }

func ParseSamsung(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "Parsing Rosa Rossa START")

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

		log.Debugf(c, "Parsing Rosa Rossa END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch parse Brand forum process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "Parsing Rosa Rossa ...")
}

func ParseSamsungSingle(w http.ResponseWriter, r *http.Request) {
	urlp := r.URL.Query().Get("url")

	helpers.SetHeaderJSON(w)

	c := appengine.NewContext(r)

	helpers.Debugf(c, "Parsing Rosa Rossa START")
	helpers.Debugf(c, "Url: "+urlp)

	Brand := parseSamsungSingle(c, SamsungURL + urlp)

	log.Debugf(c, "Parsing Rosa Rossa Single")

	helpers.ReturnJSON(w, r, Brand)
}

func parseSamsungList(c context.Context, urlp string) {
	tParsed := 0
	log.Debugf(c, "Parsing Rosa Rossa List "+urlp)

	ctxWithDeadline, _ := context.WithTimeout(c, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)
	res, err := httpc.Get(urlp)
	if err != nil {
		log.Errorf(c, "Error getting rosa rossa HP: ", err)
		return
	}
	defer res.Body.Close()

	z := html.NewTokenizer(res.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			helpers.Debugf(c, "Parsing Rosa Rossa END")
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if isDiv := t.Data == "div"; isDiv {
				helpers.Debugf(c, "Found div: %v", t)
				for _, pw := range t.Attr {
					if pw.Key == "class" && pw.Val == "box-preview" {
						helpers.Debugf(c, "Found girl: %v", pw)

						z.Next(); at := z.Token()
						helpers.Debugf(c, "Tag a: %v", at)
						for _, a := range at.Attr {
							if a.Key == "href" {
								helpers.Debugf(c, "Found href: %v", a)
								parseSamsungSingle(c, SamsungURL + a.Val)
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

	log.Debugf(c, "Parsing Rosa Rossa List: TOTAL %d", tParsed)
}

func parseSamsungSingle(c context.Context, urlp string) (Brand *mister_gadget.Brand) {
	var BrandData mister_gadget.Brand
	var gallery []string
	var idBrand = "n/a"

	log.Debugf(c, "Parsing Rosa Rossa url " + urlp)

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

	idBrand = "0"

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf(c, "Error reading body Brand site: ", err)
		return
	}

	z := html.NewTokenizer(strings.NewReader(string(body[:])))
	tt := html.StartTagToken

	for tt != html.ErrorToken {
		tt = z.Next()

		switch {
		case tt == html.ErrorToken:
		case tt == html.SelfClosingTagToken:
		case tt == html.StartTagToken:
			t := z.Token()

			if isH2 := t.Data == "h2"; isH2 {
				z.Next(); t = z.Token()
				helpers.Debugf(c, "Tag name: %v", t)
				BrandData.Name = t.Data
			}
			if isP := t.Data == "p"; isP {
				for _, nd := range t.Attr {
					if nd.Key == "id" && nd.Val == "main-image" {
						helpers.Debugf(c, "Found id main image: %v", nd)
						z.Next(); z.Next(); t = z.Token()
						for _, a := range t.Attr {
							if a.Key == "src" {
								helpers.Debugf(c, "Found src img: %v", a)
								gallery = append(gallery, a.Val)
								if !BrandData.Image.Valid {
									BrandData.Image.String = a.Val
									BrandData.Image.Valid = true
								}
								break
							}
						}
					}
					break
				}
			}
			if isA := t.Data == "a"; isA {
				isRel := false
				thisHref := "n/a"
				for _, nd := range t.Attr {
					if nd.Key == "rel" && nd.Val == "lightbox-slide" {
						helpers.Debugf(c, "Found lightbox-slide: %v", nd)
						isRel = true
					}
					if nd.Key == "href" {
						thisHref = nd.Val
					}
				}
				if isRel && thisHref != "n/a" {
					helpers.Debugf(c, "Appended gallery: %v", thisHref)
					r := strings.NewReplacer("img/clients", "wide")
					gallery = append(gallery, r.Replace(thisHref))
				}
			}
			if isStrong := t.Data == "strong"; isStrong {
				helpers.Debugf(c, "Found strong: %v", t)
				z.Next(); t = z.Token()
				helpers.Debugf(c, "Found strong: %v", t)
				if strings.TrimSpace(t.Data) == "Città:" {
					helpers.Debugf(c, "Found citta: %v", t)
					z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag city: %v", t)
					BrandData.City = strings.TrimSpace(t.Data)
				}
				if t.Data == "Telefono" {
					helpers.Debugf(c, "Found telefono: %v", t)
					z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); z.Next(); t = z.Token()
					helpers.Debugf(c, "Tag Telefono: %v", t)
					re := regexp.MustCompile("[^0-9]+")
					if t.Data[0:1] == "+" {
						BrandData.Phone = "+" + re.ReplaceAllString(t.Data,"")
					} else {
						BrandData.Phone = "+39" + re.ReplaceAllString(t.Data,"")
					}
					helpers.Debugf(c, "Tag Telefono regexp: %v", BrandData.Phone)
					if strings.Contains(BrandData.Phone, "3202264517") {
						helpers.Debugf(c, "Daniela bom bom: %v", BrandData.Phone)
						BrandData.Phone = "+393202264517"
					}
					helpers.Debugf(c, "Tag Telefono: %v", BrandData.Phone)
				}
			}

			if isDiv := t.Data == "div"; isDiv {
				for _, nd := range t.Attr {
					if nd.Key == "id" && nd.Val == "xad-text" {
						helpers.Debugf(c, "Found class xad-text: %v", nd)
						tt = z.Next()
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
				}
			}
		}
	}

	if !(BrandData.Phone != "" && idBrand != "") {
		log.Debugf(c, "Phone or ID empty: return with no save")
		return &BrandData
	}

	re := regexp.MustCompile("current_zone = [0-9]+")
	if re != nil {
		barea := re.Find(body)
		if barea != nil {
			area := string(barea[:])
			areai, _ := strconv.ParseInt(strings.TrimLeft(area, "current_zone = "), 10, 64)
			BrandData.Area.String = areas[areai-1]
			BrandData.Area.Valid = true
			helpers.Debugf(c, "Brand zone: %v", BrandData.Area)
		}
	}

	BrandData.Country = "italy"
	BrandData.Id = idBrand
	helpers.Debugf(c, "Brand: %v", BrandData)
	helpers.Debugf(c, "Brand id: %v", idBrand)
	helpers.Debugf(c, "Brand gallery: %v", gallery)

	if BrandData.Image.Valid && BrandData.Phone != "" && BrandData.Id != "" && BrandData.City != "" {
		insertBrand(c, SamsungSite, httpc, urlp, BrandData, gallery)
	} else {
		log.Debugf(c, "Brand phone, id, city or image empty")
	}

	return &BrandData
}

