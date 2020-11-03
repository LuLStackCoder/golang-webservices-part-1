package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestServer struct {
	server *httptest.Server
	Search SearchClient
}

type XmlData struct {
	XmlName string   `xml:"root"`
	XmlRows []XmlRow `xml:"row"`
}

type XmlRow struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

const AccessToken = "sasdf123"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != AccessToken {
		sendError(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		return
	}
	bytes, err := ioutil.ReadAll(file)
	data := XmlData{}
	err = xml.Unmarshal(bytes, &data)
	if err != nil {
		return
	}
	q := r.URL.Query()
	resp := make([]User, 0)
	toFind := q.Get("query")
	if toFind != "" {
		toFind := q.Get("query")
		for i := range data.XmlRows {
			if strings.Contains(data.XmlRows[i].FirstName, toFind) ||
				strings.Contains(data.XmlRows[i].LastName, toFind) ||
				strings.Contains(data.XmlRows[i].About, toFind) {
				resp = append(resp, User{
					Id:     data.XmlRows[i].ID,
					Name:   data.XmlRows[i].FirstName + " " + data.XmlRows[i].LastName,
					Age:    data.XmlRows[i].Age,
					About:  data.XmlRows[i].About,
					Gender: data.XmlRows[i].Gender,
				})
			}
		}
	} else {
		for i := range data.XmlRows {
			resp = append(resp, User{
				Id:     data.XmlRows[i].ID,
				Name:   data.XmlRows[i].FirstName + " " + data.XmlRows[i].LastName,
				Age:    data.XmlRows[i].Age,
				About:  data.XmlRows[i].About,
				Gender: data.XmlRows[i].Gender,
			})
		}
	}

	sortBy := r.URL.Query().Get("order_field")
	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		return
	}
	if orderBy != OrderByAsIs {
		var isLess func(u1, u2 User) bool

		switch sortBy {
		case "Id":
			isLess = func(u1, u2 User) bool {
				return u1.Id < u2.Id
			}
		case "Age":
			isLess = func(u1, u2 User) bool {
				return u1.Age < u2.Age
			}
		case "Name":
			fallthrough
		case "":
			isLess = func(u1, u2 User) bool {
				return u1.Name < u2.Name
			}
		default:
			sendError(w, "ErrorBadOrderField", http.StatusBadRequest)
			return
		}

		sort.Slice(resp, func(i, j int) bool {
			return isLess(resp[i], resp[j]) && (orderBy == orderDesc)
		})
	}

	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil {
		return
	}

	offset, err := strconv.Atoi(q.Get("offset"))
	if err != nil {
		return
	}

	if limit > 0 {
		from := offset
		if from > len(resp)-1 {
			resp = []User{}
		} else {
			to := offset + limit
			if to > len(resp) {
				to = len(resp)
			}

			resp = resp[from:to]
		}
	}

	js, err := json.Marshal(resp)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	return
}

func (ts *TestServer) Close() {
	ts.server.Close()
}

func sendError(w http.ResponseWriter, error string, code int) {
	js, err := json.Marshal(SearchErrorResponse{error})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(js))
}

func newTestServer(token string) TestServer {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{token, server.URL}

	return TestServer{server, client}
}

func TestLimitLow(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	_, err := ts.Search.FindUsers(SearchRequest{
		Limit: -1,
	})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "limit must be > 0" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestLimitHigh(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	response, _ := ts.Search.FindUsers(SearchRequest{
		Limit: 100,
	})

	if len(response.Users) != 25 {
		t.Errorf("Invalid number of users: %d", len(response.Users))
	}
}

func TestInvalidToken(t *testing.T) {
	ts := newTestServer(AccessToken + "invalid")
	defer ts.Close()

	_, err := ts.Search.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "Bad AccessToken" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestInvalidOrderField(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	_, err := ts.Search.FindUsers(SearchRequest{
		OrderBy:    OrderByAsc,
		OrderField: "Foo",
	})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "OrderFeld Foo invalid" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestOffsetLow(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	_, err := ts.Search.FindUsers(SearchRequest{
		Offset: -1,
	})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "offset must be > 0" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestFindUserByName(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	response, _ := ts.Search.FindUsers(SearchRequest{
		Query: "Annie",
		Limit: 1,
	})

	if len(response.Users) != 1 {
		t.Errorf("Invalid number of users: %d", len(response.Users))
		return
	}

	if response.Users[0].Name != "Annie Osborn" {
		t.Errorf("Invalid user found: %v", response.Users[0])
		return
	}
}

func TestLimitOffset(t *testing.T) {
	ts := newTestServer(AccessToken)
	defer ts.Close()

	response, _ := ts.Search.FindUsers(SearchRequest{
		Limit:  3,
		Offset: 0,
	})

	if len(response.Users) != 3 {
		t.Errorf("Invalid number of users: %d", len(response.Users))
		return
	}

	if response.Users[2].Name != "Brooks Aguilar" {
		t.Errorf("Invalid user at position 3: %v", response.Users[2])
		return
	}

	response, _ = ts.Search.FindUsers(SearchRequest{
		Limit:  5,
		Offset: 2,
	})

	if len(response.Users) != 5 {
		t.Errorf("Invalid number of users: %d", len(response.Users))
		return
	}

	if response.Users[0].Name != "Brooks Aguilar" {
		t.Errorf("Invalid user at position 3: %v", response.Users[0])
		return
	}
}

func TestFatalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Fatal Error", http.StatusInternalServerError)
	}))
	client := SearchClient{AccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if err.Error() != "SearchServer fatal error" {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestCantUnpackError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Some Error", http.StatusBadRequest)
	}))
	client := SearchClient{AccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestUnknownBadRequestError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sendError(w, "Unknown Error", http.StatusBadRequest)
	}))
	client := SearchClient{AccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if !strings.Contains(err.Error(), "unknown bad request error") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestCantUnpackResultError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "None")
	}))
	client := SearchClient{AccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	client := SearchClient{AccessToken, server.URL}
	defer server.Close()

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if !strings.Contains(err.Error(), "timeout for") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}

func TestUnknownError(t *testing.T) {
	client := SearchClient{AccessToken, "http://invalid-server/"}

	_, err := client.FindUsers(SearchRequest{})

	if err == nil {
		t.Errorf("Empty error")
	} else if !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("Invalid error: %v", err.Error())
	}
}
