package main

import (
    "time"
    "github.com/githubtofu/gator/internal/database"
    "github.com/google/uuid"
    //"html"
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

func scrapeFeeds(st *state) error {
    ctx := context.Background()
    f, err := st.db.GetNextFeedToFetch(ctx)
    if err != nil {
        return fmt.Errorf("Cannot get next feed. %w", err)
    }
    fmt.Println("[scrape]Got next feed")
    st.db.MarkFeedFetched(ctx, f.ID)
    rf, err := fetchFeed(ctx, f.Url)
    if err != nil {
        return fmt.Errorf("Cannot fetch feeds. %w", err)
    }
    fmt.Println("[scrape]Got next feed fetched .. stopping at get feed by url?")
    fmt.Println("LINK:", f.Url)
    _, err = st.db.GetFeedByUrl(ctx, f.Url)
    if err != nil {
        return fmt.Errorf("Cannot get feed by url. %w", err)
    }
    fmt.Println("[scrape]Got feed, Length:", len(rf.Channel.Item))
    for i, item := range(rf.Channel.Item) {
        fmt.Println("TITLE#", i, ":", item.Title, "PUB:", item.PubDate)
        /*
        cpParams := database.CreatePostParams{
            ID: uuid.New(),
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            Title:  item.Title,
            Url:    item.Link,
            Description: item.Description,
            PublishedAt: time.Now(),
            FeedID: uuid.New(),
        }
        _, err := st.db.CreatePost(ctx, cpParams)
        if err != nil {
            return fmt.Errorf("Cannot crete post. %w", err)
        }
        */
    }
    fmt.Println("[scrape]Finished...  next feed fetched")
    return nil
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
    /*
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
    */
    return &feed, nil
}

func handlerAgg(st *state, cmd command) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("Wrong number of arguments. needs time duration between fetches")
    }
    dur, err := time.ParseDuration(cmd.args[0])
    if err != nil {
        return fmt.Errorf("Cannot parse duration. %w", err)
    }
    fmt.Println("Collecting feeds every", cmd.args[0])
    ticker := time.NewTicker(dur)
    for ;; <-ticker.C {
        fmt.Println("[handlerAgg] Ticker...")
        scrapeFeeds(st)
    }
    /*
    f, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
    if err != nil {
        return fmt.Errorf("Cannot create request. %w", err)
    }
    fmt.Println("=========================")
    fmt.Printf("%+v\n", f)

    fmt.Println("==========Item ===============")
    fmt.Printf("%v\n", f.Channel.Item[len(f.Channel.Item)-1].Description)
    fmt.Println("=========================")
    */
    return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        cUser, err := s.db.GetUser(context.Background(), s.c.CurrentUserName)
        if err != nil {
            return err
        }
        return handler(s, cmd, cUser)
    }
}

func handlerAddFeed(st *state, cmd command, user database.User) error {
    fmt.Println("[AddFeed] In..")
    if len(cmd.args) != 2 {
        return fmt.Errorf("Wrong number of arguments provided. Needs name and url.")
    }
    feedParams := database.CreateFeedParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:   cmd.args[0],
        Url:    cmd.args[1],
        UserID: user.ID,
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
        UserID: user.ID,
        FeedID: f.ID,
    } 
    if _, err := st.db.CreateFeedFollow(context.Background(), ffParams); err != nil {
        return fmt.Errorf("Cannot create feed follow. %w", err)
    }
    return nil
}

func handlerFeeds(st *state, cmd command) error {
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

func handlerFollow(st *state, cmd command, user database.User) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("[handlerFollow] Wrong number of arguments provided")
    }
    cFeed, err := st.db.GetFeedByUrl(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("Cannot get feed by url. %w", err)
    }
    ffParams := database.CreateFeedFollowParams{
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID: user.ID,
        FeedID: cFeed.ID,
    }
    ffRow, err := st.db.CreateFeedFollow(context.Background(), ffParams)
    if err != nil {
        return fmt.Errorf("Cannot create feed follow. %w", err)
    }
    fmt.Println(ffRow.UserName, "follows", ffRow.FeedName)
    return nil
}

func handlerUnfollow(st *state, cmd command, user database.User) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("[handlerUnfollow] Wrong number of arguments provided")
    }
    dffParams := database.DeleteFeedFollowParams{
        UserID: user.ID,
        Url   : cmd.args[0], 
    }
    err := st.db.DeleteFeedFollow(context.Background(), dffParams)
    if err != nil {
        return err
    }
    return nil
}

func handlerFollowing(st *state, cmd command, user database.User) error {
    if len(cmd.args) > 0 {
        return fmt.Errorf("Wrong number of arguments provided")
    }
    cFeeds, err := st.db.GetFeedFollowsForUser(context.Background(), user.ID)
    if err != nil {
        return fmt.Errorf("Cannot get feed follows for user. %w", err)
    }
    fmt.Println(user.Name, "follows")
    for _, f := range(cFeeds) {
        fmt.Println(f.FeedName)
    }
    return nil
}
