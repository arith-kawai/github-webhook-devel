package aris

//https://developers.google.com/gmail/api/quickstart/go

import (
	"encoding/json"
	"io/ioutil"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"github.com/labstack/echo"

	"golang.org/x/net/context"
	"aris/cofig"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(c *echo.Context, ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(c, ctx, config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(c *echo.Context, ctx context.Context, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the " +
	"authorization code: \n%v\n", authURL)


	var code string = c.Query("code")
	if code == "" {
		log.Fatalf("Unable to read authorization code %v")
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func DeleteAllGmailThreads(c *echo.Context) error {
	ctx := context.TODO()

	b, err := ioutil.ReadFile(arisconf.GetSharedConfig().Googleapi.ClientSecretFilePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/gmail-go-quickstart.json
	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope, gmail.MailGoogleComScope, gmail.GmailComposeScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(c, ctx, config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"

	//	r, err := srv.Users.Labels.List(user).Do()
	nextPageToken := ""
	for x := 0; x < 10000; x++ {
		var r *gmail.ListThreadsResponse
		var err error
		if nextPageToken == "" {
			r, err = srv.Users.Threads.List(user).Do()
		} else {
			r, err = srv.Users.Threads.List(user).PageToken(nextPageToken).Do()
		}

		nextPageToken = r.NextPageToken


		if err != nil {
			log.Fatalf("Unable to retrieve labels. %v", err)
		}
		if (len(r.Threads) > 0) {

			cap := 5
			needle := 0
			for needle + cap < len(r.Threads) {
				ch := make(chan string, cap)
				log.Debugf("needle = %d, cap = %d", needle, cap)
				for _, thread := range r.Threads[needle:needle + cap] {
					go func(c chan string, id string, user string, s *gmail.Service) {
						var err error
						err = s.Users.Threads.Delete(user, id).Do()
						var txt string
						if err != nil {
							txt = err.Error()
						} else {
							txt = id
						}
						ch <- txt
					}(ch, thread.Id, user, srv)
					needle++
				}
				for i := 0; i < cap; i++ {
					result, ok := <-ch
					if ok {
						log.Print(result)
					}
				}
			}

		} else {
			log.Print("No labels found.")
		}
	}

	return c.JSON(http.StatusOK, "abc")
}
