package mister_gadget_bot

import (
	"time"
	"strconv"
	"net/url"
	"net/http"
	"database/sql"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"github.com/giona69/http-helpers"
	"google.golang.org/appengine/taskqueue"
	"github.com/giona69/mister_gadget-commons"
)

// da Brandforum
// se esiste con phone + id usare quello per update
// else se esiste quella di Samsung (query con id = "0"), usare quel phone e sovrascrivere id con BrandID (sovrascrivere anche Brand_origin_site e gallery)
// else creare nuovo record con phone + id
//
// da Samsung
// cercare solo per phone, se esiste usare quello con id pre-esistente (BrandForum)
// else usare nuovo record con id = "0"
func insertBrand(c context.Context, site int, httpc *http.Client, urlp string, Brand mister_gadget.Brand, gallery []string) {
	var etu mister_gadget.Brand
	Brand.Created = time.Now()
	Brand.Updated = time.Now()

	err := searchExistingBrand(c, site, &Brand, &etu)

	etu.Phone_active = time.Now()

	if err != nil {
		err = searchBrandByAlternatePhone(c, &Brand, &etu)
	}

	if err != nil {
		insertNewBrand(c, httpc, site, &Brand)
	} else {
		updateBrand(c, httpc, site, &Brand, &etu)
	}

	insertOrUpdateOrigin(c, site, &Brand, &etu, urlp)

	insertGallery(c, site, &Brand, gallery)

	postTask(c, site, &Brand)
}

func searchExistingBrand(c context.Context, site int, Brand *mister_gadget.Brand, etu *mister_gadget.Brand) (err error) {
	var rows *sql.Row

	if site == BrandForumSite {
		rows = getBrand.QueryRow(Brand.Phone, Brand.Id)
	} else {
		rows = getBrandByPhoneOnly.QueryRow(Brand.Phone)
	}

	if err = rows.Scan(&etu.Phone, &etu.Id, &etu.Name, &etu.Image, &etu.City, &etu.Area, &etu.Country,
		&etu.Detail_site, &etu.Height, &etu.Age, &etu.Battery, &etu.Height, &etu.Screen,
		&etu.Details, &etu.System, &etu.Text, &etu.Earphones, &etu.Rom, &etu.Ram, &etu.Storage,
		&etu.Cables, &etu.Size, &etu.Weight, &etu.Material, &etu.Gsm, &etu.Colors,
		&etu.P30Min, &etu.P60Min, &etu.P30MinOut, &etu.P60MinOut, &etu.BatteryExtra,
		&etu.EarphonesExtra, &etu.RomExtra, &etu.StorageExtra, &etu.SizeExtra, &etu.Created,
		&etu.Updated, &etu.Active, &etu.Reviews_site, &etu.Agency, &etu.Agency_site, &etu.New, &etu.Activated,
		&etu.Phone_active); err != nil {
		helpers.Debugf(c, "Error in rows.Scan: %v", err)

		// se BrandForum, allora riprovare con id di Samsung
		if site == BrandForumSite {
			rows = getBrand.QueryRow(Brand.Phone, "0")

			if err = rows.Scan(&etu.Phone, &etu.Id, &etu.Name, &etu.Image, &etu.City, &etu.Area, &etu.Country,
				&etu.Detail_site, &etu.Height, &etu.Age, &etu.Battery, &etu.Height, &etu.Screen,
				&etu.Details, &etu.System, &etu.Text, &etu.Earphones, &etu.Rom, &etu.Ram, &etu.Storage,
				&etu.Cables, &etu.Size, &etu.Weight, &etu.Material, &etu.Gsm, &etu.Colors,
				&etu.P30Min, &etu.P60Min, &etu.P30MinOut, &etu.P60MinOut, &etu.BatteryExtra,
				&etu.EarphonesExtra, &etu.RomExtra, &etu.StorageExtra, &etu.SizeExtra, &etu.Created,
				&etu.Updated, &etu.Active, &etu.Reviews_site, &etu.Agency, &etu.Agency_site, &etu.New, &etu.Activated,
				&etu.Phone_active); err != nil {
				helpers.Debugf(c, "Error in rows.Scan: %v", err)
			}
		}
	}

	helpers.Debugf(c, "searchExistingBrand err: %v", err)
	return err
}

func searchBrandByAlternatePhone(c context.Context, Brand *mister_gadget.Brand, etu *mister_gadget.Brand) (err error) {
	var ephone Phones
	var rows *sql.Row

	phoner := lphones.QueryRow(Brand.Phone)

	if err = phoner.Scan(&ephone.Phone, &ephone.Id); err != nil {
		helpers.Debugf(c, "Error in phones rows.Scan: %v", err)
	} else {
		rows = getBrand.QueryRow(ephone.Phone, ephone.Id)

		if err = rows.Scan(&etu.Phone, &etu.Id, &etu.Name, &etu.Image, &etu.City, &etu.Area, &etu.Country,
			&etu.Detail_site, &etu.Height, &etu.Age, &etu.Battery, &etu.Height, &etu.Screen,
			&etu.Details, &etu.System, &etu.Text, &etu.Earphones, &etu.Rom, &etu.Ram, &etu.Storage,
			&etu.Cables, &etu.Size, &etu.Weight, &etu.Material, &etu.Gsm, &etu.Colors,
			&etu.P30Min, &etu.P60Min, &etu.P30MinOut, &etu.P60MinOut, &etu.BatteryExtra,
			&etu.EarphonesExtra, &etu.RomExtra, &etu.StorageExtra, &etu.SizeExtra, &etu.Created,
			&etu.Updated, &etu.Active, &etu.Reviews_site, &etu.Agency, &etu.Agency_site, &etu.New, &etu.Activated,
			&etu.Phone_active); err != nil {
			helpers.Debugf(c, "Error in rows.Scan: %v", err)
		}

		if _, err := uphones.Exec(time.Now(), Brand.Phone); err != nil {
			log.Errorf(c, "Could not update Brand_phone: %v: ", err)
		}

		Brand.Id = etu.Id
		Brand.Phone = etu.Phone
	}

	helpers.Debugf(c, "searchByPhone err: %v", err)
	return err
}

func insertNewBrand(c context.Context, httpc *http.Client, site int, Brand *mister_gadget.Brand) {
	helpers.Debugf(c, "Inserting New Record")

	if Brand.Image.Valid {
		if _, err = igallery.Exec(Brand.Phone, Brand.Id, site, Brand.Image.String); err != nil {
			helpers.Debugf(c, "Could not insert picture href: %v: ", err)
		}
	}

	Brand.Activated = time.Now()

	if _, err := insert.Exec(Brand.Phone, Brand.Id, Brand.Name, Brand.Image, Brand.City, Brand.Area,
		Brand.Country, Brand.Detail_site, Brand.Height,
		Brand.Age, Brand.Battery, Brand.Height, Brand.Screen, Brand.Details, Brand.System,
		Brand.Text, Brand.Earphones, Brand.Rom, Brand.Ram, Brand.Storage, Brand.Cables, Brand.Size,
		Brand.Weight, Brand.Material, Brand.Gsm, Brand.Colors, Brand.P30Min, Brand.P60Min, Brand.P30MinOut,
		Brand.P60MinOut, Brand.BatteryExtra, Brand.EarphonesExtra, Brand.RomExtra, Brand.StorageExtra, Brand.SizeExtra,
		Brand.Created, Brand.Updated, Brand.Reviews_site, Brand.Activated); err != nil {

		log.Errorf(c, "Could not insert Brand: %v: ", err)
	} else {
		log.Debugf(c, "New Brand inserted")
		mister_gadget.InsertPost(c, httpc, Brand, Brand.Image.String, mister_gadget.PostNewBrand)
	}
}

func updateBrand(c context.Context, httpc *http.Client, site int, Brand *mister_gadget.Brand, etu *mister_gadget.Brand) {
	helpers.Debugf(c, "Updating Brand Record")

	if site == SamsungSite && etu.Id != "0" {
		Brand.Id = etu.Id
	}

	// always keep original name when updating an Brand profile
	Brand.Name = etu.Name

	if !Brand.Screen.Valid {
		Brand.Screen = etu.Screen
	}
	if Brand.Country == "" {
		Brand.Country = etu.Country
	}
	if !Brand.Area.Valid {
		Brand.Area = etu.Area
	}
	if !Brand.Details.Valid {
		Brand.Details = etu.Details
	}
	if !Brand.Age.Valid {
		Brand.Age = etu.Age
	}
	if !Brand.System.Valid {
		Brand.System = etu.System
	}
	if !Brand.Height.Valid {
		Brand.Height = etu.Height
	}
	if !Brand.Height.Valid {
		Brand.Height = etu.Height
	}
	if !Brand.Battery.Valid {
		Brand.Battery = etu.Battery
	}
	if !Brand.Detail_site.Valid {
		Brand.Detail_site = etu.Detail_site
	}
	if !Brand.Text.Valid {
		Brand.Text = etu.Text
	}
	if !Brand.Earphones.Valid {
		Brand.Earphones = etu.Earphones
	}
	if !Brand.Rom.Valid {
		Brand.Rom = etu.Rom
	}
	if !Brand.Ram.Valid {
		Brand.Ram = etu.Ram
	}
	if !Brand.Storage.Valid {
		Brand.Storage = etu.Storage
	}
	if !Brand.Cables.Valid {
		Brand.Cables = etu.Cables
	}
	if !Brand.Size.Valid {
		Brand.Size = etu.Size
	}
	if !Brand.Weight.Valid {
		Brand.Weight = etu.Weight
	}
	if !Brand.Material.Valid {
		Brand.Material = etu.Material
	}
	if !Brand.Gsm.Valid {
		Brand.Gsm = etu.Gsm
	}
	if !Brand.Colors.Valid {
		Brand.Colors = etu.Colors
	}
	if !Brand.P30Min.Valid {
		Brand.P30Min = etu.P30Min
	}
	if !Brand.P60Min.Valid {
		Brand.P60Min = etu.P60Min
	}
	if !Brand.P30MinOut.Valid {
		Brand.P30MinOut = etu.P30MinOut
	}
	if !Brand.P60MinOut.Valid {
		Brand.P60MinOut = etu.P60MinOut
	}
	if !Brand.BatteryExtra.Valid {
		Brand.BatteryExtra = etu.BatteryExtra
	}
	if !Brand.EarphonesExtra.Valid {
		Brand.EarphonesExtra = etu.EarphonesExtra
	}
	if !Brand.RomExtra.Valid {
		Brand.RomExtra = etu.RomExtra
	}
	if !Brand.StorageExtra.Valid {
		Brand.StorageExtra = etu.StorageExtra
	}
	if !Brand.SizeExtra.Valid {
		Brand.SizeExtra = etu.SizeExtra
	}
	if !Brand.Reviews_site.Valid {
		Brand.Reviews_site = etu.Reviews_site
	}

	if (!etu.Active && time.Now().After(etu.Updated.Add(72*time.Hour))) ||
		(Brand.City != etu.City && !mister_gadget.ExistPost(Brand.Phone, Brand.Id, Brand.City)) {
		mister_gadget.InsertPost(c, httpc, Brand, Brand.Image.String, mister_gadget.PostBackBrand)
		Brand.Activated = time.Now()
	} else {
		Brand.Activated = etu.Activated
	}

	if Brand.Image.Valid {
		if _, err = igallery.Exec(Brand.Phone, Brand.Id, site, Brand.Image.String); err != nil {
			helpers.Debugf(c, "Could not insert picture href: %v: ", err)
		}
	}

	if Brand.City != etu.City && Brand.Area.Valid {
		Brand.Area.String = ""
	}

	if _, err := update.Exec(Brand.Id, Brand.Name, Brand.Image, Brand.City, Brand.Area, Brand.Country, Brand.Detail_site, Brand.Height,
		Brand.Age, Brand.Battery, Brand.Height, Brand.Screen, Brand.Details, Brand.System,
		Brand.Text, Brand.Earphones, Brand.Rom, Brand.Ram, Brand.Storage, Brand.Cables, Brand.Size,
		Brand.Weight, Brand.Material, Brand.Gsm, Brand.Colors, Brand.P30Min, Brand.P60Min, Brand.P30MinOut,
		Brand.P60MinOut, Brand.BatteryExtra, Brand.EarphonesExtra, Brand.RomExtra, Brand.StorageExtra, Brand.SizeExtra,
		Brand.Updated, Brand.Reviews_site, Brand.Activated, &etu.Phone_active, Brand.Phone, etu.Id); err != nil {

		log.Errorf(c, "Could not update Brand: %v: ", err)
	} else {
		log.Debugf(c, "Brand updated")
	}
}

func insertOrUpdateOrigin(c context.Context, site int, Brand *mister_gadget.Brand, etu *mister_gadget.Brand, urlp string) {
	if site == BrandForumSite && etu.Id == "0" {
		if _, err := u2orig.Exec(Brand.Id, Brand.Phone, "0", SamsungSite); err != nil {
			log.Errorf(c, "Could not update origin site 2: %v: ", err)
		}
		if _, err := u2gallery.Exec(Brand.Id, Brand.Phone, "0", SamsungSite); err != nil {
			log.Errorf(c, "Could not delete picture href: %v: ", err)
		}
	}

	if _, err := iorig.Exec(Brand.Phone, Brand.Id, site, urlp); err != nil {
		helpers.Debugf(c, "Could not insert origin site: %v: ", err)
		if _, err := uorig.Exec(urlp, Brand.Phone, Brand.Id, site); err != nil {
			log.Errorf(c, "Could not update origin site: %v: ", err)
		}
	}
}

func insertGallery(c context.Context, site int, Brand *mister_gadget.Brand, gallery []string) {
	for _, pic := range gallery {
		if _, err = igallery.Exec(Brand.Phone, Brand.Id, site, pic); err != nil {
			helpers.Debugf(c, "Could not insert picture href: %v: ", err)
		}
	}
}

func postTask(c context.Context, site int, Brand *mister_gadget.Brand) {
	t := taskqueue.NewPOSTTask("/filer", url.Values{"phone": {Brand.Phone}, "id": {Brand.Id}, "site": {strconv.Itoa(site)}})
	if _, err := taskqueue.Add(c, t, filer); err != nil {
		log.Errorf(c, "Could not insert task: %v: ", err)
	}
}


