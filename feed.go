package main

import (
    "time"
    "github.com/githubtofu/gator/internal/database"
    "github.com/google/uuid"
    "html"
    "context"
    "net/http"
    "fmt"
    "io"
    "encoding/xml"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}


func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
    if err != nil {
        return nil, fmt.Errorf("Cannot create request. %w", err)
    }
    req.Header.Set("User-Agent", "gator")
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Error requesting. %w", err)
    }
    data, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, fmt.Errorf("Cannot read from response. %w", err)
    }
    feed := RSSFeed{}
    if err := xml.Unmarshal(data, &feed); err != nil {
        return nil, fmt.Errorf("Cannot unmarshal from response. %w", err)
    }
    fmt.Println("[FEED]Channel++++++++++++")
    fmt.Println("TITLE:", feed.Channel.Title)
    fmt.Println("DESC:", feed.Channel.Description)
    feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
    feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
    fmt.Println("[FEED]Channel++++++++++++After unescape")
    fmt.Println("TITLE2:", feed.Channel.Title)
    fmt.Println("DESC2:", feed.Channel.Description)
    for i, item := range(feed.Channel.Item) {
        fmt.Println("[FEED]Item++++++++++++Before unescape")
        fmt.Println("TITLE#", i, ":", item.Title)
        fmt.Println("DESC#", i, ":", item.Description)
        feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
        feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
        fmt.Println("[FEED]Item++++++++++++After unescape")
        fmt.Println("TITLE2#", i, ":", item.Title)
        fmt.Println("DESC2#", i, ":", item.Description)
    }
    return &feed, nil
}

func handlerAgg(st state, cmd command) error {
    f, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
    if err != nil {
        return fmt.Errorf("Cannot create request. %w", err)
    }
    fmt.Println("=========================")
    fmt.Printf("%+v\n", f)

    fmt.Println("==========Item ===============")
    fmt.Printf("%v\n", f.Channel.Item[len(f.Channel.Item)-1].Description)
    fmt.Println("=========================")
    return nil
}

func handlerAddFeed(st state, cmd command) error {
    if len(cmd.args) != 2 {
        return fmt.Errorf("Wrong number of arguments provided")
    }
    cUser, err := st.db.GetUser(context.Background(), st.c.CurrentUserName)
    if err != nil {
        return fmt.Errorf("Cannot get current user. %w", err)
    }
    feedParams := database.CreateFeedParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:   cmd.args[0],
        Url:    cmd.args[1],
        UserID: cUser.ID,
    }
    f, err := st.db.CreateFeed(context.Background(), feedParams)
    if err != nil {
        return fmt.Errorf("Cannot create feed. %w", err)
    }
    fmt.Printf("NEW FEED:%+v\n", f)
    ffParams := database.CreateFeedFollowParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID: cUser.ID,
        FeedID: f.ID,
    } 
    if _, err := st.db.CreateFeedFollow(context.Background(), ffParams); err != nil {
        return fmt.Errorf("Cannot create feed follow. %w", err)
    }
    return nil
}

func handlerFeeds(st state, cmd command) error {
    if len(cmd.args) > 0 {
        return fmt.Errorf("Wrong number of arguments provided")
    }
    feeds, err := st.db.GetFeeds(context.Background())
    if err != nil {
        return fmt.Errorf("Cannot get feeds. %w", err)
    }
    for _, a_feed := range(feeds) {
        u, err := st.db.GetUserById(context.Background(), a_feed.UserID)
        if err != nil {
            return fmt.Errorf("Cannot get user by id. %w", err)
        }
        fmt.Println("Name:", a_feed.Name, "URL:", a_feed.Url, "USER:", u.Name)
    }
    return nil
}

func handlerFollow(st state, cmd command) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("Wrong number of arguments provided")
    }
    cUser, err := st.db.GetUser(context.Background(), st.c.CurrentUserName)
    if err != nil {
        return fmt.Errorf("Cannot get current user. %w", err)
    }
    cFeed, err := st.db.GetFeedByUrl(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("Cannot get feed by url. %w", err)
    }
    ffParams := database.CreateFeedFollowParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID: cUser.ID,
        FeedID: cFeed.ID,
    }
    ffRow, err := st.db.CreateFeedFollow(context.Background(), ffParams)
    if err != nil {
        return fmt.Errorf("Cannot create feed follow. %w", err)
    }
    fmt.Println(ffRow.UserName, "follows", ffRow.FeedName)
    return nil
}

func handlerFollowing(st state, cmd command) error {
    if len(cmd.args) > 0 {
        return fmt.Errorf("Wrong number of arguments provided")
    }
    cUser, err := st.db.GetUser(context.Background(), st.c.CurrentUserName)
    if err != nil {
        return fmt.Errorf("Cannot get current user. %w", err)
    }
    cFeeds, err := st.db.GetFeedFollowsForUser(context.Background(), cUser.ID)
    if err != nil {
        return fmt.Errorf("Cannot get feed follows for user. %w", err)
    }
    fmt.Println(cUser.Name, "follows")
    for _, f := range(cFeeds) {
        fmt.Println(f.FeedName)
    }
    return nil
}
