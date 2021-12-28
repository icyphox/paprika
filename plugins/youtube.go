package plugins

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.icyphox.sh/paprika/config"
	"github.com/dustin/go-humanize"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/irc.v3"
)

type Youtube struct{}

func (Youtube) Triggers() []string {
	return []string{".yt"}
}

func (Youtube) Execute(m *irc.Message) (string, error) {
	parsed := strings.SplitN(m.Trailing(), " ", 2)
	if len(parsed) == 1 && parsed[0] == ".yt" {
		return "Usage: .yt <query>", nil
	} else if parsed[0] != ".yt" {
		return "", NoReply // ???
	}

	return YoutubeSearch(parsed[1])
}

func init() {
	Register(Youtube{})
}

var NoYtApiKey = errors.New("No Youtube Api Key")

func constructApiService() (*youtube.Service, error) {
	if apiKey, ok := config.C.ApiKeys["youtube"]; ok {
		service, err := youtube.NewService(
			context.TODO(),
			option.WithAPIKey(apiKey),
			option.WithUserAgent("github.com/icyphox/paprika"),
			option.WithTelemetryDisabled(), // wtf?
		)
		return service, err
	} else {
		return nil, NoYtApiKey
	}
}

func YoutubeSearch(query string) (string, error) {
	service, err := constructApiService()
	if err != nil {
		return "", err
	}

	search := youtube.NewSearchService(service).List([]string{})
	search = search.Q(query)
	res, err := search.Do()
	if err != nil {
		return "", err
	}

	if len(res.Items) == 0 ||
		res.Items[0].Id == nil ||
		res.Items[0].Id.VideoId == "" {
		return "[Youtube] No videos found.", nil
	}
	item := res.Items[0]

	vid := item.Id.VideoId
	println(vid)
	description, err := YoutubeDescription(vid)
	if err != nil {
		return description, err
	}

	return fmt.Sprintf("%s - https://youtu.be/%s", description, vid), nil
}

func YoutubeDescriptionFromUrl(url *url.URL) (string, error) {
	var vid string
	if url.Host == "youtu.be" {
		vid = url.Path[1:]
	} else {
		vid = url.Query().Get("v")
	}

	return YoutubeDescription(vid)
}

func YoutubeDescription(vid string) (string, error) {
	if vid == "" {
		return "[Youtube] Could not find video id.", nil
	}

	service, err := constructApiService()
	if err != nil {
		return "", err
	}

	vidservice := youtube.NewVideosService(service)
	vcall := vidservice.List([]string{"snippet", "statistics", "contentDetails"})
	vcall = vcall.Id(vid)
	vres, err := vcall.Do()
	if err != nil {
		return "", err
	}

	if len(vres.Items) == 0 {
		return "[Youtube] No video found", nil
	}

	snippet := vres.Items[0]
	if snippet.Snippet == nil || snippet.ContentDetails == nil || snippet.Statistics == nil {
		return "[Youtube] API Error. Required fields are nil.", nil
	}

	title := snippet.Snippet.Title
	duration := strings.ToLower(snippet.ContentDetails.Duration[2:])
	likes := humanize.Comma(int64(snippet.Statistics.LikeCount))
	// disabled
	// dislikes := humanize.Comma(int64(snippet.Statistics.DislikeCount))
	views := humanize.Comma(int64(snippet.Statistics.ViewCount))
	channelName := snippet.Snippet.ChannelTitle
	publishedParsed, err := time.Parse(time.RFC3339, snippet.Snippet.PublishedAt)
	var published string
	if err == nil {
		published = humanize.Time(publishedParsed)
	} else {
		published = snippet.Snippet.PublishedAt[:10]
	}

	return fmt.Sprintf(
		"\x02%s\x02 %s - \x0303↑ %s\x03 \x0304↓ ?\x03 - \x02%s\x02 views - \x02%s\x02 %s",
		title,
		duration,
		likes,
		views,
		channelName,
		published,
	), nil
}
