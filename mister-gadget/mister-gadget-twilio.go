package mister_gadget_bot

import (
	"database/sql"
	"errors"
	"github.com/giona69/http-helpers"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/runtime"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"net/url"
	"strings"
	"time"
)


func Twilio(w http.ResponseWriter, r *http.Request) {

	helpers.SetHeaderJSON(w)

	ctx := appengine.NewContext(r)
	err := runtime.RunInBackground(ctx, func(c context.Context) {
		log.Debugf(c, "send twilio START")

		sendTwilio(c)

		log.Debugf(c, "send twilio END")
	})
	if err != nil {
		helpers.Errorf(w, r, http.StatusInternalServerError, "Could not launch send twilio process: %v", err)
		return
	}

	helpers.ReturnOkJSON(w, r, "sending twilio ...")
}

func sendTwilio(c context.Context) {
	log.Debugf(c, "send TWILIO START")
	twilioQ, err  = db.Prepare("SELECT phone, Screen FROM Brand where active = TRUE AND phone != '+39123321123321' and Brand.phone like '%+39%' and phone not in (select phone from Brand_sms) LIMIT 25")
	if err != nil {
		log.Errorf(c,"Error in db.Prepare: %v", err)
		return
	}

	iMessage := "Metti anche tu il tuo annuncio su mister_gadget"
	eMessage := "Put your ad on mister_gadget"
	pMessage := "Coloque seu anúncio no mister_gadget"
	sMessage := "¡Pon tu anuncio en mister_gadget"
	rMessage := "Поместите свое объявление в mister_gadget"

	rows, err := twilioQ.Query()
	if err != nil {
		log.Errorf(c,"Error in db.Query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var nation sql.NullString
		var phone string
		if err := rows.Scan(&phone, &nation); err != nil {
			log.Errorf(c, "Error in rows.Scan: %v", err)
			continue
		}

		switch nation.String {
		case "Brazilian":
			twilioApiMessage(c, phone, pMessage)
		case "Portuguese":
			twilioApiMessage(c, phone, pMessage)
		case "Argentinian":
			twilioApiMessage(c, phone, sMessage)
		case "Dominican":
			twilioApiMessage(c, phone, sMessage)
		case "Colombiana":
			twilioApiMessage(c, phone, sMessage)
		case "Spanish":
			twilioApiMessage(c, phone, sMessage)
		case "Portoricana":
			twilioApiMessage(c, phone, sMessage)
		case "Venezuelan":
			twilioApiMessage(c, phone, sMessage)
		case "Costa Rican":
			twilioApiMessage(c, phone, sMessage)
		case "Cuban":
			twilioApiMessage(c, phone, sMessage)
		case "Italian":
			twilioApiMessage(c, phone, iMessage)
		case "Russian":
			twilioApiMessage(c, phone, rMessage)
			twilioApiMessage(c, phone, eMessage)
		case "Australian":
			twilioApiMessage(c, phone, eMessage)
		case "American":
			twilioApiMessage(c, phone, eMessage)
		case "British":
			twilioApiMessage(c, phone, eMessage)
		default:
			twilioApiMessage(c, phone, iMessage)
		}
		if _, err = twilioI.Exec(phone); err != nil {
			helpers.Debugf(c, "Could not insert Brand_sms record: %v: ", err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Errorf(c, "Error in Row error: %v", err)
	}

	log.Debugf(c, "send TWILIO END")
}

func twilioApiMessage(ctx context.Context, phone string, body string) (err error) {
	ctxWithDeadline, _ := context.WithTimeout(ctx, 60*time.Second)
	httpc := urlfetch.Client(ctxWithDeadline)

	form := url.Values{}
	form.Add("From", "mister_gadget")
	form.Add("To", phone)
	form.Add("Body", body)
	req, err := http.NewRequest("POST", "https://api.twilio.com/2010-04-01/Accounts/AC0fff8f6c33ab83d72a80c019647b8a09/Messages", strings.NewReader(form.Encode()))
	req.PostForm = form
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("AC0fff8f6c33ab83d72a80c019647b8a09", "8dab951119bf01aff83d59c9b930350f")
	res, err := httpc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errors.New("error twilio response")
	}

	return nil
}