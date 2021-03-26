package mister_gadget_bot_module

import (
	"github.com/giona69/http-helpers"
	mister_gadget_bot "github.com/giona69/mister-gadget-bot/mister-gadget"
	"github.com/giona69/mister_gadget-commons"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// init is run before the application starts serving.
func init() {
	r := mux.NewRouter()

	r.Path("/mister_gadget-cron/v1/twilio").Methods("GET").HandlerFunc(mister_gadget_bot.Twilio)
	r.Path("/mister_gadget-cron/v1/twilio").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/sitemap").Methods("GET").HandlerFunc(mister_gadget_bot.CreateSitemap)
	r.Path("/mister_gadget-cron/v1/sitemap").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/once").Methods("GET").HandlerFunc(mister_gadget_bot.OncePerDay)
	r.Path("/mister_gadget-cron/v1/once").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/bonus").Methods("GET").HandlerFunc(mister_gadget_bot.BonusPost)
	r.Path("/mister_gadget-cron/v1/bonus").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseAll").Methods("GET").HandlerFunc(mister_gadget_bot.ParseAllBrands)
	r.Path("/mister_gadget-cron/v1/parseAll").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseReinforce").Methods("GET").HandlerFunc(mister_gadget_bot.ParseReinforce)
	r.Path("/mister_gadget-cron/v1/parseReinforce").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseEFList").Methods("GET").HandlerFunc(mister_gadget_bot.ParseBrandForum)
	r.Path("/mister_gadget-cron/v1/parseEFList").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseRRList").Methods("GET").HandlerFunc(mister_gadget_bot.ParseSamsung)
	r.Path("/mister_gadget-cron/v1/parseRRList").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/checkReadyBrands").Methods("GET").HandlerFunc(mister_gadget_bot.CheckReadyBrands)
	r.Path("/mister_gadget-cron/v1/checkReadyBrands").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/resizeAllPictures").Methods("GET").HandlerFunc(mister_gadget_bot.ResizeAllPictures)
	r.Path("/mister_gadget-cron/v1/resizeAllPictures").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/resizeMainPictures").Methods("GET").HandlerFunc(mister_gadget_bot.ResizeMainPictures)
	r.Path("/mister_gadget-cron/v1/resizeMainPictures").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/resizePostPictures").Methods("GET").HandlerFunc(mister_gadget_bot.ResizePostPictures)
	r.Path("/mister_gadget-cron/v1/resizePostPictures").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/resizeOnePicture").Methods("GET").HandlerFunc(mister_gadget_bot.ResizeOnePicture)
	r.Path("/mister_gadget-cron/v1/resizeOnePicture").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseEFAjax").Methods("GET").HandlerFunc(mister_gadget_bot.ParseBrandForumAjaxToScreen)
	r.Path("/mister_gadget-cron/v1/parseEFAjax").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseEFSingle").Methods("GET").HandlerFunc(mister_gadget_bot.ParseBrandForumSingle)
	r.Path("/mister_gadget-cron/v1/parseEFSingle").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseEFSingleToScreen").Methods("GET").HandlerFunc(mister_gadget_bot.ParseBrandForumSingleToScreen)
	r.Path("/mister_gadget-cron/v1/parseEFSingleToScreen").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/mister_gadget-cron/v1/parseRRSingle").Methods("GET").HandlerFunc(mister_gadget_bot.ParseSamsungSingle)
	r.Path("/mister_gadget-cron/v1/parseRRSingle").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/filer").HandlerFunc(mister_gadget_bot.Filer)
	r.Path("/filer").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/render").HandlerFunc(mister_gadget_bot.SaveRenderedPage)
	r.Path("/render").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	r.Path("/resizer").HandlerFunc(mister_gadget_bot.Resizer)
	r.Path("/resizer").Methods("OPTIONS").HandlerFunc(helpers.GetOptions)

	http.Handle("/", r)
	log.Print("init done")
	mister_gadget_bot.InitAPI()
	mister_gadget.InitAPI()
}

