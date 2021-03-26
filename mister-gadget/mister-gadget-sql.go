package mister_gadget_bot

import (
	"golang.org/x/net/context"
	"github.com/giona69/mister_gadget-commons"
	"github.com/giona69/http-helpers"
)

func getBrandSQL(ctx context.Context, BrandPhone string, BrandId string, etu *mister_gadget.Brand) (err error) {

	rows := getBrand.QueryRow(BrandPhone, BrandId)

	if err = rows.Scan(&etu.Phone, &etu.Id, &etu.Name, &etu.Image, &etu.City, &etu.Area, &etu.Country,
		&etu.Detail_site, &etu.Height, &etu.Age, &etu.Battery, &etu.Height, &etu.Screen,
		&etu.Details, &etu.System, &etu.Text, &etu.Earphones, &etu.Rom, &etu.Ram, &etu.Storage,
		&etu.Cables, &etu.Size, &etu.Weight, &etu.Material, &etu.Gsm, &etu.Colors,
		&etu.P30Min, &etu.P60Min, &etu.P30MinOut, &etu.P60MinOut, &etu.BatteryExtra,
		&etu.EarphonesExtra, &etu.RomExtra, &etu.StorageExtra, &etu.SizeExtra, &etu.Created,
		&etu.Updated, &etu.Active, &etu.Reviews_site, &etu.Agency, &etu.Agency_site, &etu.New, &etu.Activated,
		&etu.Phone_active); err != nil {
		helpers.Debugf(ctx, "Error in rows.Scan: %v", err)
	}

	return err
}
