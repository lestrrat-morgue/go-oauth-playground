# oauth-consumer

Does three-legged Oauth2 flow, using a configuration file. This tool is used
to test out applications that rely on OAuth2, and is NOT meant to be used for
anything else.

The easiest mode of operation is to:

1. Configure your identity provider to expect a redirect URL of "http://your.app.your.domain:8080/oauth_callback"
2. Edit your /etc/hosts file so that your.app.your.domain points to 127.0.0.1
3. Fill out config.json (note: `redirect_uri` field must match what you configured your identity provider with)
4. execute `go run cmd/oauth-consumer/oauth-consumer.go`
5. Access `http://your.app.your.domain:8080/`

## Configuration file

The configuration file is in JSON format, and supports the following items

| element name | required? | description |
|:-------------|:---------:|:------------|
| auth_url_params | N | Map of additional parameters to be added when redirecting to the identity provider's authentication url. For Google's G Suite, for example, you can specify the "hd" parameter to restrict the domain where the accounts can be chosen from |
| client_id | Y |The OAuth2 Client ID, as given by the identity provider |
| client_secret | Y | The OAuth2 Client Secret, as given by the identity provider |
| listen | N | The address to listen on. By default the server binds to :8080 |
| redirect_url | Y | The URL which the identity provider redirects back the user to. The domain name et al can be tweaked as needed, but the path must be `/oauth_callback` |
| scopes | Y | The authorization scope to request from the identity provider. It's required most of the time, but depends on the identity provider |