# gator
Desc: Aggregates blogs for multiple users with a CLI.  Not for real use.  Exersize only.
TODO:  FIX DOCS:)

Requirements:
Postgres
Go

Installation
run go install github.com/striderjg/gator@latest
Create a file '.gatorconfig.json' in your home directory containing:
{
  "db_url": "postgres://example"
}
register a user with: gator register USERNAME

Dont' know if you have to muck around with migrating the db up. I'll try installing on my laptop later and update.  

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

