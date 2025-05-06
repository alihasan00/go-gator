package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/alihasan00/gator/internal/config"
	"github.com/alihasan00/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	config *config.Config
	db     *database.Queries
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	commandHandler, exists := c.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("command does not exist")
	}
	err := commandHandler(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

// middlewareLoggedIn is a middleware function that checks if a user is logged in
// and passes the user to the handler function
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.config.CurrentUserLoggedIn == "" {
			return fmt.Errorf("you must be logged in to use this command")
		}

		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserLoggedIn)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		return handler(s, cmd, user)
	}
}

func HandleLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login requires only one argument")
	}

	user, err := s.db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User name has been set to %v\n", s.config.CurrentUserLoggedIn)
	return nil
}

func HandleRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login requires only one argument")
	}

	user := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	newUser, err := s.db.CreateUser(context.Background(), user)
	if err != nil {
		return err
	}

	err = s.config.SetUser(newUser.Name)
	if err != nil {
		return err
	}

	fmt.Printf("New user is created %v and logged to new user %v\n", newUser.Name, s.config.CurrentUserLoggedIn)
	return nil
}

func HandleReset(s *state, cmd command) error {
	err := s.db.ResetUser(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func HandleUsers(s *state, cmd command) error {
	users, err := s.db.GeAllUser(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if s.config.CurrentUserLoggedIn == user.Name {
			fmt.Println(user.Name, "(current)")
		} else {
			fmt.Println(user.Name)
		}
	}

	return nil
}

func HandleAgg(s *state, cmd command) error {
	fullUrl := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), fullUrl)
	if err != nil {
		return err
	}
	fmt.Println(feed)
	return nil
}

// Refactored to use the logged-in middleware
func HandlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("addfeed requires two arguments: name and url")
	}

	feed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      sql.NullString{String: cmd.Args[0], Valid: true},
		Url:       sql.NullString{String: cmd.Args[1], Valid: true},
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	}

	createdFeed, err := s.db.CreateFeed(context.Background(), feed)
	if err != nil {
		return err
	}

	fellowFeed := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	}

	_, err = s.db.CreateFeedFollow(context.Background(), fellowFeed)
	if err != nil {
		return err
	}

	fmt.Println("Feed created successfully:")
	fmt.Printf("ID: %s\n", createdFeed.ID)
	fmt.Printf("Name: %s\n", createdFeed.Name.String)
	fmt.Printf("URL: %s\n", createdFeed.Url.String)
	fmt.Printf("User ID: %s\n", createdFeed.UserID.UUID)
	fmt.Printf("Created At: %s\n", createdFeed.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated At: %s\n", createdFeed.UpdatedAt.Format(time.RFC3339))
	return nil
}

func HandleFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println("Feed created successfully:")
		fmt.Printf("ID: %s\n", feed.ID)
		fmt.Printf("Name: %s\n", feed.Name.String)
		fmt.Printf("URL: %s\n", feed.Url.String)
		user, err := s.db.GetUserByID(context.Background(), feed.UserID.UUID)
		if err != nil {
			return err
		}
		fmt.Printf("User Name: %s\n", user.Name)
	}

	return nil
}

// Refactored to use the logged-in middleware
func HandlerFollowFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("follow requires one argument: feed URL")
	}

	feed, err := s.db.GetFeedsByUrl(context.Background(), sql.NullString{String: cmd.Args[0], Valid: true})
	if err != nil {
		return err
	}

	fellowFeed := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    uuid.NullUUID{UUID: feed.ID, Valid: true},
	}

	createdFeed, err := s.db.CreateFeedFollow(context.Background(), fellowFeed)
	if err != nil {
		return err
	}
	fmt.Printf("feed fellow entry feed name %v\n", createdFeed.FeedName)
	fmt.Printf("feed fellow entry user name %v\n", createdFeed.UserName)

	return nil
}

// Refactored to use the logged-in middleware
func HandlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("User %v is following the feed %v\n", feed.UserName, feed.FeedName)
	}

	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {

	feed, err := s.db.GetFeedsByUrl(context.Background(), sql.NullString{String: cmd.Args[0], Valid: true})

	if err != nil {
		return err
	}

	followFeed := database.DeleteFeedFollowsParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: uuid.NullUUID{UUID: feed.ID, Valid: true},
	}

	err = s.db.DeleteFeedFollows(context.Background(), followFeed)
	if err != nil {
		return err
	}
	return nil
}

func CommandHandler(s *state) {
	cmds := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	cmds.register("login", HandleLogin)
	cmds.register("register", HandleRegister)
	cmds.register("reset", HandleReset)
	cmds.register("users", HandleUsers)
	cmds.register("agg", HandleAgg)
	cmds.register("feeds", HandleFeeds)

	// Use middlewareLoggedIn for commands that require authentication
	cmds.register("addfeed", middlewareLoggedIn(HandlerAddFeed))
	cmds.register("follow", middlewareLoggedIn(HandlerFollowFeed))
	cmds.register("following", middlewareLoggedIn(HandlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handleUnfollow))

	args := os.Args
	fmt.Println(args)

	// Process command if arguments are provided
	if len(args) == 1 && args[0] == "users" || args[0] == "reset" || args[0] == "agg" || args[0] == "feeds" || args[0] == "following" {
		cmd := command{
			Name: args[0],
		}
		err := cmds.run(s, cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
	if len(args) > 1 {
		cmd := command{
			Name: args[1],
			Args: args[2:],
		}

		err := cmds.run(s, cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		os.Exit(1)
	}
}
