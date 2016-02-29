package arisgithub

import (
	"github.com/labstack/echo"
	"net/http"
	"github.com/labstack/gommon/log"
	"crypto/hmac"
	"crypto/sha1"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"aris/cofig"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-sql-driver/mysql"
)

func Webhook(c *echo.Context) error {
	//log.Info(c.Request().Header)

	payload, error := extractPayload(c)
	if error != nil {
		log.Error(error)
	}

	hubSignature, error := extractHubSignature(c)
	if error != nil {
		log.Error(error)
	}

	error = validateSignature(arisconf.GetSharedConfig().Github.Secret, payload, hubSignature)
	if error != nil {
		log.Error(error)
	}

	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &dat); err != nil {
		log.Error("...")
	}

	log.Debug(payload)

	githubDelivery := extractGithubDelivery(c)
	if githubDelivery == "" {
		log.Error("githubDelivery is empty.")
	}

	arisconf := arisconf.GetSharedConfig()
	dbconf := &mysql.Config{
		User: arisconf.Database.User,
		Passwd: arisconf.Database.Password,
		Net: arisconf.Database.Protocol,
		Addr: arisconf.Database.Host,
		DBName: arisconf.Database.DbName,
		Params: map[string]string{
			//"ssl-ca":arisconf.Database.CaCertFilePath,
			"tls":"skip-verify",
		},

	}

	db, error := sql.Open(arisconf.Database.Type, dbconf.FormatDSN())
	if error != nil {
		log.Error(error)
	}
	defer db.Close()

	stmt, error := db.Prepare("insert into github_webhook (id, payload) values (?, ?)")
	if error != nil {
		log.Error(error)
	}
	defer stmt.Close()

	result, error := stmt.Exec(githubDelivery, payload)
	if error != nil {
		log.Error(error)
	}
	log.Debug(result)

	return c.JSON(http.StatusOK, "abc")
}

func extractPayload(c *echo.Context) (payload string, error error) {

	if contentType := c.Request().Header.Get("Content-Type"); contentType == "application/json" {
		payBytes := new(bytes.Buffer)
		_, error := payBytes.ReadFrom(c.Request().Body)
		if error != nil {
			//エラー処理
			return "", error
		}
		payload = payBytes.String()
	} else {
		payload = c.Form("payload")
	}
	log.Debug(payload)

	return payload, error
}

func extractHubSignature(c *echo.Context) (hubSignature string, error error) {
	//githubEvent := c.Request().Header.Get("X-Github-Event")
	//githubDelivery := c.Request().Header.Get("X-Github-Delivery")
	hubSignature = c.Request().Header.Get("X-Hub-Signature")
	//log.Debugf("X-Github-Event:%s X-Github-Delivery:%s X-Hub-Signature: %s", githubEvent, githubDelivery, hubSignature)

	if hubSignature == "" {
		error = errors.New("A signature does not exist.")
	}
	return hubSignature, error
}

func extractGithubDelivery(c *echo.Context) (githubDelivery string){
	githubDelivery = c.Request().Header.Get("X-Github-Delivery")
	return githubDelivery
}

func validateSignature(secret string, payload string, hubSignature string) error {
	var error error
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(payload))
	mySignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(hubSignature), []byte(mySignature)) {
		error = errors.New("HMAC verification failed")
	}
	return error
}