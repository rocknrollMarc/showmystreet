package geoweb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"appengine"
	"appengine/urlfetch"
)

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/showimage", showimage)
}

// handles the Request.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, rootForm)
}

const rootForm = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Accept Address</title>
  </head>
  <body>
    <h1>Accept Address</h1>
    <p>Please enter your address:</p>
    <form action="/showimage" method="post" accept-charset="utf-8">
      <input type="text" name="str" value="Type address..." id="str" />
      <input type="submit" value=".. and see the image!" />
    </form
  </body>
</html>
`

var upperTemplate = template.Must(template.New("showimage").Parse(upperTemplateHTML))

func showimage(w http.ResponseWriter, r *http.Request) {

	addr := r.FormValue("str")

	safeAddr := url.QueryEscape(addr)
	fullUrl := fmt.Sprintf("http://maps.googleapis.com/maps/api/geocode/json?sensor=false&address=%s", safeAddr)

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)

	resp, err := client.Get(fullUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Read the context into a byte array
	body, dataReadErr := ioutil.ReadAll(resp.Body)
	if dataReadErr != nil {
		log.Fatal("ReadAll: ", dataReadErr)
		return
	}

	res := make(map[string][]map[string]map[string]map[string]interface{}, 0)
	json.Unmarshal(body, &res)

	lat, _ := res["results"][0]["geometry"]["location"]["lat"]
	lng, _ := res["results"][0]["geometry"]["location"]["lng"]

	// %.13f is used to convert float64 to a string
	queryUrl := fmt.Sprintf("http://maps.googleapis.com/maps/api/streetview?sensor=false&size=600x300&location=%.13f,%.13f", lat, lng)

	tempErr := upperTemplate.Execute(w, queryUrl)
	if tempErr != nil {
		http.Error(w, tempErr.Error(), http.StatusInternalServerError)
	}
}

const upperTemplateHTML = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Display Image</title>
  </head>
  <body>
    <h3>Image at your Address</h3>
    <img src="{{html .}}" alt="Image" />
  </body>
</html>
`
