# gator
Name: gator
Desc: Aggregates blogs for multiple users with a CLI.  Not for real use.  Exersize only.
TODO:  FIX DOCS:)

Requirements:
Postgres
Go

Installation
run go install github.com/striderjg/gator@latest
Create a file '.gatorconfig.json' in your home directory containing:
{
  "db_url": "postgres://{postgresUsername}:{PASSWORD}@localhost:5432/gator"

}
Create a db in postgres called gator
Run an .sql file on that database with the following commands:

CREATE TABLE users(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE feeds(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE feed_follows(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    feed_id UUID NOT NULL,
    CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_feeds FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(user_id, feed_id)
);

ALTER TABLE feeds ADD last_fetched_at TIMESTAMP;

CREATE TABLE posts(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL,
    CONSTRAINT fk_feeds FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

curse me for not making an install script
register a user with: gator register USERNAME

basic usage is: gator COMMAND [Args]
Available commands are:
login USERNAME - sets the active user
register USERNAME - registers a user
users - lists the users
reset - resets all the databases (WARNING: DELETES EVERYTHING)
agg TIMEDURATION - Checks for need items on followed feeds every TIMEDURATION.  TIMEDURATION must be > 1s.  Format is #m#s ect.
addfeed DESCRIPTION URL = Adds a feed ot the database with DESCRIPTION at URL. Automatically follows the feed for the current user.
feeds - List the feeds in the DB
follow URL - Follows a feed in the DB with URL
following - Lists the feeds the active user is following
unfollow URL - unfollows the feed with URL
browse  [NUMBER_POSTS(DEFUALT=2)] - Lists the NUMBER_POSTS to list from followed feeds.  Most recent first.  Defaults to 2 posts.

