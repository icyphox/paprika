package plugins

import (
	"context"
	"net/url"
	"strings"
	"time"

	"git.icyphox.sh/paprika/config"
	"github.com/dustin/go-humanize"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func YoutubeDescription(url *url.URL) (string, error) {
	apiKey := ""
	ok := false
	if apiKey, ok = config.C.ApiKeys["youtube"]; !ok {
		return "[Youtube] No API Key!", nil
	}

	var vid string
	if url.Host == "youtu.be" {
		vid = url.Path[1:]
	} else {
		vid = url.Query().Get("v")
	}

	if vid == "" {
		return "[Youtube] Could not find video id.", nil
	}

	service, err := youtube.NewService(
		context.TODO(),
		option.WithAPIKey(apiKey),
		option.WithUserAgent("github.com/icyphox/paprika"),
		option.WithTelemetryDisabled(), // wtf?
	)
	if err != nil {
		return "", err
	}

	vidservice := youtube.NewVideosService(service)

	vcall := vidservice.List([]string{"snippet","statistics","contentDetails"});
	vcall = vcall.Id(vid)
	vres, err := vcall.Do()
	if err != nil {
		return "", err
	}

	if len(vres.Items) == 0 {
		return "[Youtube] No video found", nil
	}

	snippet := vres.Items[0]
	var out strings.Builder

	
	// Title
	out.WriteByte(2)
	out.WriteString(snippet.Snippet.Title)
	out.WriteByte(2)

	// Duration
	out.WriteByte(' ')
	duration := strings.ToLower(snippet.ContentDetails.Duration[2:])
	out.WriteString(duration)
	out.WriteString(" - ")

	// Likes
	out.WriteByte(3)
	out.WriteString("03")
	out.WriteRune('↑')
	out.WriteByte(' ')
	likes := humanize.Comma(int64(snippet.Statistics.LikeCount))
	out.WriteString(likes)
	out.WriteByte(3)
	out.WriteByte(' ')
	// Dislikes
	// Deprecated on 2021, Dec 13 ????
	out.WriteByte(3)
	out.WriteString("04")
	out.WriteRune('↓')
	out.WriteByte(' ')
	out.WriteString("???")
	out.WriteByte(3)
	out.WriteString(" - ")

	// Views
	out.WriteByte(2)
	views := humanize.Comma(int64(snippet.Statistics.ViewCount))
	out.WriteString(views)
	out.WriteByte(2)
	out.WriteString(" Views - ")

	// Channel / Author
	out.WriteByte(2)
	out.WriteString(snippet.Snippet.ChannelTitle)
	out.WriteByte(2)
	out.WriteByte(' ')
	// date
	publishedParsed, err := time.Parse(time.RFC3339, snippet.Snippet.PublishedAt)
	var published string
	if err == nil {
		published = humanize.Time(publishedParsed)
	} else {
		published = snippet.Snippet.PublishedAt[:10]
	}
	out.WriteString(published)


	return out.String(), nil
}
