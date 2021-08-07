package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/auth"
)

type QiniuCDNClientDef struct {
	QINIU_AK     string
	QINIU_SK     string
	QINIU_DOMAIN string
	AuthClient   *auth.Credentials
}

var QiniuCDNClient QiniuCDNClientDef

type CertManagerDef struct {
	ACME_KEY_PATH string
	DOMAIN        string
	Name          string //  date
	// openssl rsa -in '*.muxixyz.com.key' -out '*.muxixyz.com.key.pem' -outform PEM
	// openssl x509 -in fullchain.cer -out fullchain.pem -outform PEM => nouse
	Pri string // doman.key -> domain.key.pem
	Ca  string // fullchan.cer -> fullchan.pem => pem will lose chan, so back to cer
}

var CertManager CertManagerDef

func init() {
	CertManager = CertManagerDef{
		ACME_KEY_PATH: os.Getenv("ACME_KEY_PATH"),
		Name:          time.Now().Format("06-01-02") + " " + os.Getenv("QINIU_DOMAIN"),
		DOMAIN:        os.Getenv("QINIU_DOMAIN"),
	}

	QiniuCDNClient = QiniuCDNClientDef{
		QINIU_AK:     os.Getenv("QINIU_AK"),
		QINIU_SK:     os.Getenv("QINIU_SK"),
		QINIU_DOMAIN: os.Getenv("QINIU_DOMAIN"),
	}
	QiniuCDNClient.AuthClient = auth.New(QiniuCDNClient.QINIU_AK, QiniuCDNClient.QINIU_SK)
}

type GetDomainListResponse struct {
	Domains []struct {
		Name               string      `json:"name"`
		PareDomain         string      `json:"pareDomain"`
		Type               string      `json:"type"`
		Cname              string      `json:"cname"`
		TestURLPath        string      `json:"testURLPath"`
		Protocol           string      `json:"protocol"`
		Platform           string      `json:"platform"`
		GeoCover           string      `json:"geoCover"`
		QiniuPrivate       bool        `json:"qiniuPrivate"`
		OperationType      string      `json:"operationType"`
		OperatingState     string      `json:"operatingState"`
		OperatingStateDesc string      `json:"operatingStateDesc"`
		FreezeType         string      `json:"freezeType"`
		CreateAt           time.Time   `json:"createAt"`
		ModifyAt           time.Time   `json:"modifyAt"`
		CouldOperateBySelf bool        `json:"couldOperateBySelf"`
		UIDIsFreezed       bool        `json:"uidIsFreezed"`
		OemMail            string      `json:"oemMail"`
		TagList            interface{} `json:"tagList"`
		OperTaskID         string      `json:"operTaskId"`
		OperTaskType       string      `json:"operTaskType"`
		OperTaskErrCode    int         `json:"operTaskErrCode"`
	} `json:"domains"`
	Marker string `json:"marker"`
}

func (c QiniuCDNClientDef) getDomainList() GetDomainListResponse {
	url := "http://api.qiniu.com/domain?limit=500"
	method := "GET"

	client := &http.Client{
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Println(err)
	}

	err = c.AuthClient.AddToken(auth.TokenQBox, req)
	if err != nil {
		log.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return GetDomainListResponse{}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return GetDomainListResponse{}
	}

	//log.Println(string(body))
	var respBody GetDomainListResponse
	err = json.Unmarshal(body, &respBody)

	return respBody
}

type updateCertRequest struct {
	CertId      string `json:"certId"`
	ForceHttps  bool   `json:"forceHttps"`
	HTTP2Enable bool   `json:"http2Enable"`
}

func (c QiniuCDNClientDef) UpdateDomainsSSL(certID string, domainName string) {
	url := fmt.Sprintf("http://api.qiniu.com/domain/%s/httpsconf", domainName)
	method := "PUT"

	reqBody := updateCertRequest{
		CertId:      certID,
		ForceHttps:  false,
		HTTP2Enable: true,
	}

	jsons, err := json.Marshal(reqBody)

	payload := strings.NewReader(string(jsons))

	client := &http.Client{
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	err = c.AuthClient.AddToken(auth.TokenQBox, req)
	if err != nil {
		log.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(body))
	//var respBody GetDomainListResponse
	//err = json.Unmarshal(body, &respBody)

	return
}

type certRequest struct {
	Name       string `json:"name"`
	CommonName string `json:"common_name"`
	CA         string `json:"ca"`
	Pri        string `json:"pri"`
}

type UploadPemCertResponse struct {
	Code   int    `json:"code"`
	Error  string `json:"error"`
	CertID string `json:"certID"`
}

func (c QiniuCDNClientDef) UploadPemCert(certManager CertManagerDef) string {
	url := "http://api.qiniu.com/sslcert"
	method := "POST"

	reqBody := certRequest{
		Name:       certManager.Name,
		CommonName: certManager.DOMAIN,
		CA:         certManager.Ca,
		Pri:        certManager.Pri,
	}

	jsons, err := json.Marshal(reqBody)

	payload := strings.NewReader(string(jsons))

	client := &http.Client{
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	err = c.AuthClient.AddToken(auth.TokenQBox, req)
	if err != nil {
		log.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	//log.Println(string(body))

	var respBody UploadPemCertResponse
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Println(err)
	}

	return respBody.CertID
}

func (c *CertManagerDef) GeneratePemFile() {
	// TODO
	// using system openssl tools
}

func (c *CertManagerDef) ReadPemCert() {
	// //Users/mac/.acme.sh + *.muxixyz.com + /
	path := c.ACME_KEY_PATH + c.DOMAIN + "/"

	content, err := ioutil.ReadFile(path + c.DOMAIN + ".key.pem")
	if err != nil {
		log.Println(err)
	}

	c.Pri = string(content)

	content, err = ioutil.ReadFile(path + "fullchain.cer")
	if err != nil {
		log.Println(err)
	}

	c.Ca = string(content)
}

func main() {
	//log.Println(QINIU_AK, QINIU_SK, QINIU_DOMAIN, ACME_KEY_PATH)
	CertManager.ReadPemCert()
	//log.Println(CertManager)
	newCertId := QiniuCDNClient.UploadPemCert(CertManager)
	domains := QiniuCDNClient.getDomainList().Domains
	for _, v := range domains {
		if v.Protocol == "https" {
			QiniuCDNClient.UpdateDomainsSSL(newCertId, v.Name)
			log.Println("Update dns " + v.Name + " successfully.")
		}
	}
}
