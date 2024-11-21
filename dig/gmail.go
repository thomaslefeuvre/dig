package dig

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"

	"google.golang.org/api/gmail/v1"
)

const (
	user        = "me"
	searchQuery = "in:inbox from:noreply@bandcamp.com"
	maxResults  = 500
)

var (
	urlRegex = regexp.MustCompile(`https?://[^\s"]+`)
)

type Gmail struct {
	service *gmail.Service
}

func NewGmail(service *gmail.Service) *Gmail {
	return &Gmail{
		service: service,
	}
}

func (g *Gmail) Collect(r *Result) (*Result, error) {
	if r == nil {
		return nil, fmt.Errorf("uninitialised result")
	}

	if r.Full() {
		return r, nil
	}

	user := "me"
	listMsgsReq := g.service.Users.Messages.List(user).Q(searchQuery).MaxResults(maxResults)

	messages, err := listMsgsReq.Do()
	if err != nil {
		return nil, fmt.Errorf("retrieve messages: %w", err)
	}

	for _, msg := range messages.Messages {
		if r.Full() {
			break
		}
		getMsgReq := g.service.Users.Messages.Get(user, msg.Id)
		fullMsg, err := getMsgReq.Do()
		if err != nil {
			log.Printf("Unable to retrieve message %v: %v", msg.Id, err)
			continue
		}

		urls, err := extractURLs(fullMsg)
		if err != nil {
			log.Printf("Unable to extract urls from message %v: %v", msg.Id, err)
			continue
		}

		for _, url := range urls {
			r.Add(url)
		}
	}

	return r, nil
}

func extractURLs(msg *gmail.Message) ([]string, error) {
	urls := []string{}

	for _, part := range msg.Payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			body, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				return nil, fmt.Errorf("decode message body: %w", err)
			}

			partUrls := urlRegex.FindAllString(string(body), -1)
			urls = append(partUrls, urls...)
		}
	}

	return urls, nil
}
