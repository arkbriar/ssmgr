package slack

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

type SlackrusHook struct {
	logrus.Hook

	Channel        string
	Token          string
	AcceptedLevels []logrus.Level

	c    *slack.Client
	msgQ chan *logrus.Entry
}

func (h *SlackrusHook) Levels() []logrus.Level {
	return h.AcceptedLevels
}

func color(l logrus.Level) string {
	switch l {
	case logrus.DebugLevel:
		return "#9B30FF"
	case logrus.InfoLevel:
		return "good"
	case logrus.WarnLevel:
		return "warning"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return "danger"
	default:
		return "warning"
	}
}

func fire(c *slack.Client, channel string, e *logrus.Entry) error {
	logText := fmt.Sprintf("%s [%s] %s", strings.ToUpper(fmt.Sprint(e.Level)), e.Time.Format(time.RubyDate), e.Message)

	params := slack.PostMessageParameters{
		AsUser: true,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color: color(e.Level),
				Text:  logText,
			},
		},
		EscapeText: true,
	}

	// fill attachment fields

	attach := &params.Attachments[0]
	if len(e.Data) > 0 {
		for k, v := range e.Data {
			field := slack.AttachmentField{
				Title: k,
				Value: fmt.Sprint(v),
			}
			if len(field.Value) < 20 {
				field.Short = true
			}
			attach.Fields = append(attach.Fields, field)
		}
	}

	_, _, err := c.PostMessage(channel, "", params)
	return err
}

func (h *SlackrusHook) open() {
	h.c = slack.New(h.Token)
	h.msgQ = make(chan *logrus.Entry, 200)

	// start the message posting goroutine
	go func(c *slack.Client, channel string, in chan *logrus.Entry) {
		for msg := range in {
			if err := fire(c, channel, msg); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to send log to slack: %v\n", err)
			}
		}
	}(h.c, h.Channel, h.msgQ)
}

func (h *SlackrusHook) Connect() *SlackrusHook {
	h.open()
	return h
}

func (h *SlackrusHook) Fire(e *logrus.Entry) error {
	if h.c == nil {
		return errors.New("slack is not connected")
	}

	if len(h.msgQ) == 200 {
		return errors.New("message queue is full")
	}

	h.msgQ <- e

	return nil
}
