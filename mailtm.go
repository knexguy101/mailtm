package mailtm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
type Domain struct {
	Path      string    `json:"@id"`
	Type      string    `json:"@type"`
	Id        string    `json:"id"`
	Name      string    `json:"domain"`
	IsActive  bool      `json:"isActive"`
	IsPrivate bool      `json:"isPrivate"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type DomainResponse struct {
	Context     string   `json:"@context"`
	Path        string   `json:"@id"`
	Type        string   `json:"@type"`
	Domains     []Domain `json:"hydra:member"`
	DomainCount int      `json:"hydra:totalItems"`
}
type TokenResponse struct {
	Token string `json:"token"`
}
type Message struct {
	Path           string    `json:"@id"`
	Type           string    `json:"@type"`
	Id             string    `json:"id"`
	AccountId      string    `json:"accountId"`
	MessageId      string    `json:"msgid"`
	From           User      `json:"from"`
	To             []User    `json:"to"`
	Subject        string    `json:"subject"`
	Intro          string    `json:"intro"`
	Seen           bool      `json:"seen"`
	IsDeleted      bool      `json:"isDeleted"`
	HasAttachments bool      `json:"hasAttachments"`
	Size           int       `json:"size"`
	DownloadUrl    string    `json:"downloadUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
type MessageResponse struct {
	Context      string    `json:"@context"`
	Path         string    `json:"@id"`
	Type         string    `json:"@type"`
	Messages     []Message `json:"hydra:member"`
	MessageCount int       `json:"hydra:totalItems"`
}
type AccountInfo struct {
	Context    string    `json:"@context"`
	Path       string    `json:"@id"`
	Type       string    `json:"@type"`
	Id         string    `json:"id"`
	Email      string    `json:"address"`
	Quota      int       `json:"quota"`
	UsedCount  int       `json:"used"`
	IsDisabled bool      `json:"isDisabled"`
	IsDeleted  bool      `json:"isDeleted"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
type MailClient struct {
	T 			*http.Transport
	URL         string
	Email       string
	Password    string
	BearerToken string
	Information AccountInfo
}

func NewMailClient(t *http.Transport) *MailClient {
	return &MailClient{
		T: t,
		URL:      "https://api.mail.tm",
		Email:    "",
		Password: "",
		Information: AccountInfo{
			Context:    "",
			Path:       "",
			Type:       "",
			Id:         "",
			Email:      "",
			Quota:      0,
			UsedCount:  0,
			IsDisabled: false,
			IsDeleted:  false,
			CreatedAt:  time.Time{},
			UpdatedAt:  time.Time{},
		},
	}
}

func (c *MailClient) GetAvailableDomains() ([]Domain, error) {
	req, _ := http.NewRequest("GET", c.URL + "/domains", nil)
	res, err := c.T.RoundTrip(req)
	if err != nil {
		return []Domain{}, errors.New("fetching domains from the mail.tm server failed")
	}

	defer res.Body.Close()

	var domainResponse DomainResponse

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []Domain{}, errors.New("mail.tm server response is not readable")
	}

	err = json.Unmarshal(body, &domainResponse)
	if err != nil {
		return []Domain{}, errors.New("mail.tm server response is not parseable")
	}

	return domainResponse.Domains, nil
}

func (c *MailClient) Register(username string, domain string, password string) error {
	c.Email = username + "@" + domain
	c.Password = password

	req, _ := http.NewRequest("POST", c.URL+"/accounts", bytes.NewBuffer([]byte("{\"address\":\""+username+"@"+domain+"\",\"password\":\""+password+"\"}")))
	req.Header.Add("Content-Type", "\"application/json\"")
	res, err := c.T.RoundTrip(req)
	if err != nil {
		return errors.New("registering a new mail.tm account failed")
	}

	defer res.Body.Close()

	var accountInfo AccountInfo

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("mail.tm server response is not readable")
	}

	err = json.Unmarshal(body, &accountInfo)
	if err != nil {
		return errors.New("mail.tm server response is not parseable")
	}

	c.Information = accountInfo

	return nil
}

func (c *MailClient) Login() error {
	req, _ := http.NewRequest("POST", c.URL+"/token", bytes.NewBuffer([]byte("{\"address\":\""+c.Email+"\",\"password\":\""+c.Password+"\"}")))
	req.Header.Add("Content-Type", "\"application/json\"")
	res, err := c.T.RoundTrip(req)
	if err != nil {
		return errors.New("logging in with account failed")
	}

	defer res.Body.Close()

	var tokenResponse TokenResponse

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("mail.tm server response is not readable")
	}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return errors.New("mail.tm server response is not parseable")
	}

	c.BearerToken = tokenResponse.Token

	return nil
}

func (c *MailClient) Delete() error {
	req, err := http.NewRequest("DELETE", c.URL+"/accounts/"+c.Information.Id, nil)
	if err != nil {
		return errors.New("deleting an account from the mail.tm server failed")
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := c.T.RoundTrip(req)
	if err != nil {
		return errors.New("deleting a mail.tm account failed")
	}

	defer res.Body.Close()

	return nil
}

func (c *MailClient) GetMessages(page int) ([]Message, error) {
	req, err := http.NewRequest("GET", c.URL+"/messages?page="+strconv.Itoa(page), nil)
	if err != nil {
		return []Message{}, errors.New("creating a request to get an email from the mail.tm server failed")
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := c.T.RoundTrip(req)
	if err != nil {
		return []Message{}, errors.New("getting an email from the mail.tm server failed")
	}

	defer res.Body.Close()

	var messageResponse MessageResponse

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []Message{}, errors.New("mail.tm server response is not readable")
	}

	err = json.Unmarshal(body, &messageResponse)
	if err != nil {
		return []Message{}, errors.New("mail.tm server response is not parseable")
	}

	return messageResponse.Messages, nil
}

func (c *MailClient) GetMessage(id string) (Message, error) {

	req, err := http.NewRequest("GET", c.URL+"/messages/"+id, nil)
	if err != nil {
		return Message{}, errors.New("creating a request to get a message from the mail.tm server failed")
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := c.T.RoundTrip(req)
	if err != nil {
		return Message{}, errors.New("getting a message from the mail.tm server failed")
	}

	defer res.Body.Close()

	var message Message

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Message{}, errors.New("mail.tm server response is not readable")
	}

	err = json.Unmarshal(body, &message)
	if err != nil {
		return Message{}, errors.New("mail.tm server response is not parseable")
	}

	return message, nil
}

func (c *MailClient) DeleteMessage(id string) error {

	req, err := http.NewRequest("DELETE", c.URL+"/messages/"+id, nil)
	if err != nil {
		return errors.New("creating a request to delete a message from the mail.tm server failed")
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := c.T.RoundTrip(req)
	if err != nil {
		return errors.New("deleting a message from the mail.tm server failed")
	}

	defer res.Body.Close()

	return nil
}

func (c *MailClient) MarkMessageAsSeen(id string) error {
	var client = &http.Client{}

	req, err := http.NewRequest("PATCH", c.URL+"/messages/"+id, bytes.NewBufferString("true"))
	if err != nil {
		return errors.New("creating a request to mark a message as seen failed")
	}

	req.Header.Add("Content-Type", "merge-patch+json")
	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := client.Do(req)
	if err != nil {
		return errors.New("marking a message as seen failed")
	}

	defer res.Body.Close()

	return nil
}

func (c *MailClient) GetMessageSource(id string) (string, error) {

	req, err := http.NewRequest("GET", c.URL+"/sources/"+id, nil)
	if err != nil {
		return "", errors.New("creating a request to get the source of a message from the mail.tm server failed")
	}

	req.Header.Add("Authorization", "Bearer "+c.BearerToken)

	res, err := c.T.RoundTrip(req)
	if err != nil {
		return "", errors.New("getting the source of a message from the mail.tm server failed")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("mail.tm server response is not readable")
	}

	return string(body), nil
}
