package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	if err := _main(); err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	}
}

type config struct {
	AuthURLParams map[string]string `json:"auth_url_params"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	Listen        string            `json:"listen"`
	Scopes        []string          `json:"scopes"`
	RedirectURL   string            `json:"redirect_url"`
}

func _main() error {
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "")
	flag.Parse()

	fh, err := os.Open(configFile)
	if err != nil {
		return errors.Wrap(err, `failed to open config file`)
	}
	defer fh.Close()

	var c config
	if err := json.NewDecoder(fh).Decode(&c); err != nil {
		return errors.Wrap(err, `failed to parse config file`)
	}
	if len(c.Listen) == 0 {
		c.Listen = ":8080"
	}

	var C = oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       c.Scopes,
		RedirectURL:  c.RedirectURL,
	}
	var options []oauth2.AuthCodeOption
	for k, v := range c.AuthURLParams {
		options = append(options, oauth2.SetAuthURLParam(k, v))
	}

	states := make(map[string]time.Time)
	var mu sync.RWMutex

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h := sha512.New()
		io.CopyN(h, rand.Reader, 512)
		s := fmt.Sprintf("%x", h.Sum(nil))

		mu.Lock()
		states[s] = time.Now().Add(15 * time.Minute)
		mu.Unlock()

		w.Header().Set("Location", C.AuthCodeURL(s, options...))
		w.WriteHeader(http.StatusFound)
	})

	http.HandleFunc("/oauth_callback", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		expires, ok := states[r.FormValue("state")]
		mu.RUnlock()
		if !ok {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		// regardless of the any subsequent errors, we will
		// delete this state
		mu.Lock()
		delete(states, r.FormValue("state"))
		mu.Unlock()

		if time.Now().After(expires) {
			http.Error(w, "expired state", http.StatusBadRequest)
			return
		}

		token, err := C.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			http.Error(w, "failed to exchange code", http.StatusInternalServerError)
			return
		}

		buf, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			http.Error(w, "failed to marshal token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	})

	log.Printf("Starting to serve...")
	return http.ListenAndServe(c.Listen, nil)
}
