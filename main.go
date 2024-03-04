package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/gin-gonic/gin"
	"html"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Solution struct {
		Response string            `json:"response"`
		Headers  map[string]string `json:"headers"`
	} `json:"solution"`
}

func main() {
	app := gin.Default()
	app.GET("/*url", func(c *gin.Context) {
		url := c.Param("url")[1:]
		urlWithQuery := url + "?" + c.Request.URL.RawQuery
		payload := []byte(fmt.Sprintf(`{"cmd": "request.get", "url": "%s"}`, urlWithQuery))
		buffer := bytes.NewBuffer(payload)
		post, err := http.Post("http://localhost:8191/v1", "application/json", buffer)
		if err != nil {
			print(err)
			return
		}
		defer post.Body.Close()
		all, err := io.ReadAll(post.Body)
		if err != nil {
			print(err)
			return
		}
		var response Response
		err = json.Unmarshal(all, &response)
		if err != nil {
			c.String(200, string(all))
			return
		}
		if response.Status != "ok" {
			c.String(500, response.Message)
			return
		}
		c.Header("Content-Type", response.Solution.Headers["content-type"])
		if strings.Contains(response.Solution.Response, "<body><div id=\"webkit-xml-viewer-source-xml\">") {
			doc, err := xmlquery.Parse(strings.NewReader(response.Solution.Response))
			if err != nil {
				return
			}
			node := xmlquery.FindOne(doc, "//div[@id=\"webkit-xml-viewer-source-xml\"]")
			c.Header("Content-Type", "text/xml;charset=UTF-8")
			c.String(200, node.OutputXML(false))
			return
		}
		if strings.Contains(response.Solution.Response, "?xml version") {
			if strings.Contains(response.Solution.Response, "<body><pre") {
				root, err := htmlquery.Parse(strings.NewReader(response.Solution.Response))
				if err != nil {
					c.String(500, err.Error())
					return
				}
				c.Header("Content-Type", "text/xml;charset=UTF-8")
				c.String(200, html.UnescapeString(htmlquery.OutputHTML(htmlquery.FindOne(root, "//pre"), false)))
				return
			}
		}
		c.String(200, response.Solution.Response)
	})
	app.Run(":8192")
}
