#github-usercheck
Simple tool to check https://github.com/ for a given set of usernames. It prints out the usernames that may not be taken (urls that do not return status code 200) or using github.com/signup_check if you pass an auth_token.

You can pipe in usernames or supply a file. The names need to be separated by a newline.

##Usage
`./github-usercheck <<< ausernamethatnoonehas`

`./github-usercheck -path names.txt -workers 4 >> results.txt`

## Options
```
-auth string
     authenticity_token for post request to github
 -path string
     The filepath of the names
 -sleep int
     Sleep duration between each workers task. (Millisecond) (default 100)
 -workers int
     How many workers to run in parallel. (More scrapers are faster, but more prone to rate limiting or bandwith issues) (default 2)
```
