package main

import (
	"io"
	"net/http"

	"html/template"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/rs/cors"
	"github.com/thoas/stats"
//	"github.com/labstack/gommon/log"
	_ "os"
	_ "bufio"
	"bytes"
	"github.com/labstack/gommon/log"
//	"time"
//	"golang.org/x/oauth2/google"
//	"golang.org/x/net/context"
	"google.golang.org/api/gmail/v1"
	"golang.org/x/oauth2"
//	"encoding/json"
//	"fmt"
//	"io/ioutil"
//	"encoding/json"
//	"io/ioutil"
	"aris"
	"aris/github"
	"aris/cofig"
	"strconv"
)

type (
// Template provides HTML template rendering
	Template struct {
		templates *template.Template
	}

	user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

var (
	users map[string]user
)

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//----------
// Handlers
//----------

func welcome(c *echo.Context) error {
	return c.Render(http.StatusOK, "welcome", "Joe")
}

func gConnect(c *echo.Context) error {
	// 認証トークンを取得する。（取得後、キャッシュへ）
	//	conf := getOAuthConf()
	//	token, err := conf.Exchange(c, c.Get("code"))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	text, err := json.Marshal(token)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Printf("token is saved to ./token/token.json")
	//	err = ioutil.WriteFile("./token/token.json", text, 0777)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	return c.JSON(http.StatusOK, "")
}

func getOAuthConf() (cfg *oauth2.Config) {
	return &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "",
		Scopes:       []string{gmail.MailGoogleComScope},
		Endpoint: oauth2.Endpoint{},
	}
}

func gAuth(c *echo.Context) error {
	//https://developers.google.com/identity/protocols/application-default-credentials
	//https://github.com/golang/oauth2/blob/master/google/example_test.go
	conf := getOAuthConf()
	authUrl := conf.AuthCodeURL("")
	return c.Redirect(302, authUrl)


	// 認証トークンを取得する。（取得後、キャッシュへ）
	/*token, err := conf.Exchange(c, code)
	if err != nil {
		log.Fatal(err)
	}
	text, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("token is saved to ./token/token.json")
	err = ioutil.WriteFile("./token/token.json", text, 0777)
	if err != nil {
		log.Fatal(err)
	}*/
	//	ctx := context.TODO()
	//	client, error := google.DefaultClient(ctx, gmail.MailGoogleComScope)
	//	if error != nil {
	//		return error
	//	}
	//	//	client.CheckRedirect()
	//	service, error := gmail.New(client)
	//	if error != nil {
	//		return error
	//	}
	//	//	service.Users.Threads
	//	log.Info(service)
	//	return c.Render(http.StatusOK, "authUrl", authUrl)
}

func createUser(c *echo.Context) error {
	u := new(user)
	body := c.Request().Body
	//	log.Info(body)
	//	writer := os.Stdout
	//	writer := bytes.Buffer{}
	//	io.Copy(writer, body)
	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(body)

	log.SetLevel(log.DEBUG)
	log.Debug(n, err)
	log.Debug(buf.String())

	if err := c.Bind(u); err != nil {
		return err
	}
	users[u.ID] = *u
	return c.JSON(http.StatusCreated, u)
}

func getUsers(c *echo.Context) error {
	return c.JSON(http.StatusOK, users)
}

func getUser(c *echo.Context) error {
	return c.JSON(http.StatusOK, users[c.P(0)])
}

func main() {
	//conf := getApplicationConfig("./application.yaml")
	conf := arisconf.GetSharedConfig()

	e := echo.New()

	// Middleware
	e.Use(mw.Logger())
	e.Use(mw.Recover())
	e.Use(mw.Gzip())


	log.SetLevel(log.DEBUG)
	log.Debugf("application config: %d, %s, %s", conf.Webserver.Port, conf.Github.Secret, conf.Googleapi.ClientSecretFilePath)

	//------------------------
	// Third-party middleware
	//------------------------

	// https://github.com/rs/cors
	e.Use(cors.Default().Handler)

	// https://github.com/thoas/stats
	s := stats.New()
	e.Use(s.Handler)
	// Route
	e.Get("/stats", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, s.Data())
	})

	// Serve index file
	e.Index("public/index.html")

	// Serve static files
	e.Static("/scripts", "public/scripts")

	//--------
	// Routes
	//--------

	e.Post("/users", createUser)
	e.Get("/users", getUsers)
	e.Get("/users/:id", getUser)

	//-----------
	// Templates
	//-----------


	e.SetRenderer(&Template{
		// Cached templates
		templates: template.Must(template.ParseFiles(
			"public/views/welcome.html",
			"public/views/gconnect.html",
		)),
	})
	e.Get("/welcome", welcome)
	e.Get("/gconnect", gConnect)
	e.Get("/gauth", gAuth)
	e.Get("/delete-all-gmail-threads", aris.DeleteAllGmailThreads)


	//-------
	// Group
	//-------

	// Group with parent middleware
	a := e.Group("/admin")
	a.Use(func(c *echo.Context) error {
		// Security middleware
		return nil
	})
	a.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Welcome admin!")
	})

	// Group with no parent middleware
	g := e.Group("/files", func(c *echo.Context) error {
		// Security middlewareb
		return nil
	})
	g.Get("", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Your files!")
	})

	githubGroup := e.Group("/github")
	githubGroup.Post("/webhook", arisgithub.Webhook)

	server := e.Server(":" + strconv.Itoa(conf.Webserver.Port))
	server.TLSConfig = nil
	server.ListenAndServe()
}

func init() {
	users = map[string]user{
		"1": user{
			ID:   "1",
			Name: "Wreck-It Ralph",
		},
	}
}