# Spotify `nowplaying`

Just a simple script I got up to show my current Spotify "now playing" onto a max7219 display I have through a ntfy server.

> Note: Due to the display's character constraints, I'm normalizing the strings to remove diacritics. Might remove that at some point in the future

# Setting up the app and getting a token
- [Register an app on Spotify](https://developers.spotify.com)
- > Note: you should probably use a local address with an unused port like `http://localhost:3000` in the registration process
- Clone this repo and go into it
- Run `go run ./cmd/auth/main.go`
  - Input the app's registered data
  - Go to the URL and allow your app to access your account
  - Copy the authorization code from the URL you've been redirected to and input it
  - You should get a refresh token, which you can then give to the `nowplaying` app
- Setup the environment variables as described below
- Build and run `./cmd/nowplaying/main.go`

# Environment variables
When running the `nowplaying` executable, you'll need to setup the following environment variables:
- `CLIENT_ID`: the client ID you got during the app registration
- `CLIENT_SECRET`: the client secret you got during the app registration
- `REFRESH_TOKEN`: the refresh token you got while doing the setup step
- `NTFY_URL`: the topic name to send now playing messages to (e.g. `https://ntfy.sh/topicname`)

# Crediting
- https://github.com/Regis24GmbH/go-diacritics (string normalization, might remove later)