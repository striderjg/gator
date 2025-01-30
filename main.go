package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"

	"github.com/striderjg/gator/internal/config"
	"github.com/striderjg/gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdHandlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmdHandlers[name] = f
}
func (c *commands) run(s *state, cmd command) error {
	if _, ok := c.cmdHandlers[cmd.name]; !ok {
		return fmt.Errorf("command: %v does not exist", cmd.name)
	}
	return c.cmdHandlers[cmd.name](s, cmd)
}

// ===================== HANDLERS ===============================================
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("user %v doesn't exist: %w", cmd.args[0], err)
	}

	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User: %v has been set\n", cmd.args[0])
	return nil
}

func handlerAddFeed(s *state, cmd command, usr database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("the addfeed handler expects two arguments: Usage: addfeed NAME URL")
	}
	ctx := context.Background()
	_, err := fetchFeed(ctx, cmd.args[1])
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}

	feedEntry, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    usr.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed entry: %w", err)
	}
	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    usr.ID,
		FeedID:    feedEntry.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed_follows entry: %w", err)
	}

	fmt.Println("Added to feeds:")
	fmt.Println("=================================")
	fmt.Printf("id: %v, created_at: %v, updated_at: %v\n", feedEntry.ID, feedEntry.CreatedAt, feedEntry.UpdatedAt)
	fmt.Println("\tname: ", feedEntry.Name)
	fmt.Println("\turl: ", feedEntry.Url)
	fmt.Println("\tuser_id: ", feedEntry.UserID)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds from db: %w", err)
	}
	for _, feed := range feeds {
		fmt.Println("1)")
		fmt.Println("\tName: ", feed.Name)
		fmt.Println("\tUrl: ", feed.Url)
		fmt.Println("\tOwner: ", feed.Username)
		fmt.Println("=================================================")
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, usr database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("unfollow expects an argument.  Usage: unfollow URL")
	}

	ctx := context.Background()
	feed, err := s.db.GetFeed(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("error retrieving feed: %w", err)
	}

	// TODO:  Change query and return more useful stuff for a action performed message

	_, err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: usr.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error deleting feed_follow entry: %w", err)
	}

	return nil
}

func handlerFollow(s *state, cmd command, usr database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("follow expect an argument.  Usage: follow URL")
	}

	ctx := context.Background()
	feed, err := s.db.GetFeed(ctx, cmd.args[0])
	if err != nil {
		return fmt.Errorf("error retrieving feed: %w", err)
	}

	ff, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    usr.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed_follows entry: %w", err)
	}

	fmt.Printf("User %v is now following %v\n", ff.UserName, ff.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, usr database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), usr.ID)
	if err != nil {
		fmt.Errorf("error retrieving follows for user: %w", err)
	}

	fmt.Printf("User: %v is following:\n", s.cfg.Current_user)
	fmt.Println("===============================")
	for _, feed := range feeds {
		fmt.Println("  * ", feed.FeedName)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error fetching feed: %w", err)
	}
	//feed = feed

	fmt.Printf("<TITLE>%v</TITLE>\n", feed.Channel.Title)
	fmt.Printf("<LINK>%v</LINK>\n", feed.Channel.Link)
	fmt.Printf("<Description>%v</Description>\n", feed.Channel.Description)
	fmt.Println("============================================================")

	for _, item := range feed.Channel.Item {
		fmt.Printf("\t<TITLE>%v</TITLE>\n", item.Title)
		fmt.Printf("\t<LINK>%v</LINK>\n", item.Link)
		fmt.Printf("\t<Description>%v</Description>\n", item.Description)
		fmt.Println("============================================================")
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving users: %w", err)
	}
	for _, usr := range users {
		fmt.Printf("* %v", usr)
		if usr == s.cfg.Current_user {
			fmt.Println(" (current)")
		} else {
			fmt.Println("")
		}
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	if err := s.db.ClearDB(ctx); err != nil {
		return fmt.Errorf("error reseting the db: %w", err)
	}
	fmt.Println("Database reset")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the register command expects a single argument, the username")
	}
	ctx := context.Background()
	usr, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("error creating user %v: %w", cmd.args[0], err)
	}
	if err := s.cfg.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Printf("User %v was created:\n", cmd.args[0])
	fmt.Printf("(id: %v, created_at: %v, updated_at: %v, name: %v\n", usr.ID, usr.CreatedAt, usr.UpdatedAt, usr.Name)
	return nil
}

// ============================== Utility Functions ==================================================

func middlewareLoggedIn(handler func(s *state, cmd command, usr database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		usr, err := s.db.GetUser(context.Background(), s.cfg.Current_user)
		if err != nil {
			return fmt.Errorf("error fetching current user: %w", err)
		}
		return handler(s, cmd, usr)
	}
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting responce: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New("bad Status Code from response")
	}

	/*
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading response body: %w", err)
		}
		fmt.Println("Raw XML Body: ", string(body))
		if err := xml.Unmarshal(body, &feed); err != nil {
			return nil, fmt.Errorf("error during Unmarshaling: %w", err)
		}
	*/

	var feed RSSFeed
	//var feed TestRSSFeed
	//var feed TestRSSFeed
	//var feedT TestChannel

	decoder := xml.NewDecoder(res.Body)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("error decoding RSSFeed: %w", err)
	}
	// ===================== TODO:  Try to figure out what's going on with channel link
	//  == Leaving the comments in for later use.
	/*
			for {
				t, err := decoder.Token()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, fmt.Errorf("error reading XML token: %w", err)
				}
				switch elem := t.(type) {
				case xml.StartElement:
					fmt.Println("In Start Element")
					if elem.Name.Local == "link" {
						var link string
						fmt.Println(elem.Name)
						if err := decoder.DecodeElement(&link, &elem); err != nil {
							return nil, fmt.Errorf("error decoding <link>: %w", err)
						}
						fmt.Println("Raw <link> value: ", link)
					}
				case xml.CharData:
					fmt.Println("CharData")
					fmt.Println(elem)
					fmt.Println("Lenght:", len(elem))
					for data := range elem {
						fmt.Println(string(data))
					}
				default:
					fmt.Println("In Default")
					//var decodedElem string
					//if err := decoder.DecodeElement(&decodedElem, &elem); err != nil {
					//	return nil, fmt.Errorf("error decoding <link>: %w", err)
					//}
					fmt.Println(elem)
				}
			}

		fmt.Println("Decoded Link: ", feed.Channel.Link)
	*/

	//fmt.Println(feed.Channel.Link)

	// TODO:  Remove workaround when figure out what's going on with Channel.Link
	if feed.Channel.Link == "" {
		feed.Channel.Link = feedURL
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, _ := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
}

// . ================================ ENTRY POINT ============================================
func main() {
	// --------- INIT
	mainState := state{}
	var err error
	mainState.cfg, err = config.Read()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmds := commands{
		cmdHandlers: make(map[string]func(*state, command) error),
	}

	db, err := sql.Open("postgres", mainState.cfg.Db_url)
	if err != nil {
		fmt.Printf("Error opening database %v\n", err.Error())
	}
	mainState.db = database.New(db)

	// -- Handler registration
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("users", handlerGetUsers)
	cmds.register("reset", handlerReset)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	// -- Start
	if len(os.Args) < 2 {
		fmt.Println("Usage: gator requires a command in format gator COMMAND [ARGUMENTS]")
		os.Exit(1)
	}
	err = cmds.run(&mainState, command{os.Args[1], os.Args[2:]})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
