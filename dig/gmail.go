package dig

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"time"

	"google.golang.org/api/gmail/v1"
)

const (
	user        = "me"
	searchQuery = "in:inbox from:noreply@bandcamp.com"
	maxPageSize = 100

	quotaUnitPerSecondPerUser = 250
	quotaUnitsListMessages    = 5
	quotaUnitsGetMessage      = 5
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

func (g *Gmail) Collect(c *Collection) (*Collection, error) {
	if c == nil {
		return nil, fmt.Errorf("uninitialised result")
	}

	messages, err := g.ListAllMessages()
	if err != nil {
		return nil, fmt.Errorf("list all messages: %w", err)
	}

	for _, msg := range messages {
		getMsgReq := g.service.Users.Messages.Get(user, msg.Id)
		fullMsg, err := getMsgReq.Do()
		if err != nil {
			log.Printf("unable to retrieve message %v: %v", msg.Id, err)
			continue
		}

		log.Printf("retrieved full message %v", msg.Id)

		urls, err := extractURLs(fullMsg)
		if err != nil {
			log.Printf("unable to extract urls from message %v: %v", msg.Id, err)
			continue
		}

		for _, url := range urls {
			c.Add(url)
		}

		time.Sleep((quotaUnitsGetMessage * time.Second) / quotaUnitPerSecondPerUser)
	}

	return c, nil
}

func (g *Gmail) ListAllMessages() ([]*gmail.Message, error) {
	var messages []*gmail.Message

	listMsgsReq := g.service.Users.Messages.List(user).Q(searchQuery).MaxResults(maxPageSize)

	for {
		msgs, err := listMsgsReq.Do()
		if err != nil {
			return nil, fmt.Errorf("retrieve messages: %w", err)
		}

		log.Printf("retrieved %d Gmail message ids", len(msgs.Messages))

		messages = append(messages, msgs.Messages...)

		if msgs.NextPageToken == "" {
			break
		}

		listMsgsReq.PageToken(msgs.NextPageToken)

		time.Sleep((quotaUnitsListMessages * time.Second) / quotaUnitPerSecondPerUser)
	}

	return messages, nil
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
