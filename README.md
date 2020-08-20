A simple CLI-script that automatically create Milestone.

It is also possible to send notifications to Rocket.Chat.

Available arguments:
```
gitlaburl - Gitlab URL
token - Gitlab Token 
group - id group in Gitlab 
mllength - New milestone length in days 
mlname - New milestone name. %W & %Y - reqiured. Example: "Week \%W/\%Y"
rocketurl - Rocket.Chat URL
user - rocketchat username
pass - password of rocketuser
channel - rocketchat channel to notify
```

Example:

```
go build -o autocreate main.go
./autocreate -gitlaburl http://gitlab.dev -token aaaaBBBBcccc1111 -group 10 -mllength 7 -mlname "Week \%W/\%Y"
```

Example with Rocket.Chat

```
go build -o autocreate main.go
./autocreate -gitlaburl http://gitlab.dev -token aaaaBBBBcccc1111 -group 10 -mllength 7 -mlname "Week \%W/\%Y" -rocketurl https://rocket.company.io -user bot -pass botpass -channel "#rocketchannel"
```

Thats all.

PS If you need notify in another messenger, open issue.