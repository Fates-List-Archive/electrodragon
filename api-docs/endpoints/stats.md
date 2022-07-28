
**API URL**: ``https://api.fateslist.xyz``

**Widgets Documentation:** ``https://lynx.fateslist.xyz/widgets`` (docs for widgets available at https://lynx.fateslist.xyz/widgets)

## Authorization

- **Bot:** These endpoints require a bot token. 
You can get this from Bot Settings. Make sure to keep this safe and in 
a .gitignore/.env. A prefix of `Bot` before the bot token such as 
`Bot abcdef` is supported and can be used to avoid ambiguity but is not 
required. The default auth scheme if no prefix is given depends on the
endpoint: Endpoints which have only one auth scheme will use that auth 
scheme while endpoints with multiple will always use `Bot` for 
backward compatibility

- **Server:** These endpoints require a server
token which you can get using ``/get API Token`` in your server. 
Same warnings and information from the other authentication types 
apply here. A prefix of ``Server`` before the server token is 
supported and can be used to avoid ambiguity but is not required.

- **User:** These endpoints require a user token. You can get this 
from your profile under the User Token section. If you are using this 
for voting, make sure to allow users to opt out! A prefix of `User` 
before the user token such as `User abcdef` is supported and can be 
used to avoid ambiguity but is not required outside of endpoints that 
have both a user and a bot authentication option such as Get Votes. 
In such endpoints, the default will always be a bot auth unless 
you prefix the token with `User`. **A access token (for custom clients)
can also be used on *most* endpoints as long as the token is prefixed with 
``Frostpaw``**

- **Special:** These endpoint employ their own authentication system (such as ``slwebset``)

## Base Response

A default API Response will be of the below format:

```json
{
    done: false | true,
    reason: "" | null,
    context: "" | null
}
```

## Get Bot List Stats
### GET `https://api.fateslist.xyz`/stats

Returns the bot list stats. This currently returns the full list of all bots
as a vector/list of IndexBot structs.

As a client, it is your responsibility, to parse this. Pagination may be added
if the list grows and then requires it.



**Response Body**

- **total_bots** => i64 [0]
- **total_servers** => i64 [0]
- **total_users** => i64 [0]
- **bots** => (Array) Struct IndexBot 
	- **guild_count** => i64 [30]
	- **description** => string ["My description"]
	- **banner** => string ["My banner or default banner url"]
	- **votes** => i64 [40]
	- **state** => i32 [3]
	- **user** => Struct User 
		- **id** => string [""]
		- **username** => string [""]
		- **disc** => string [""]
		- **avatar** => string [""]
		- **bot** => bool [false]
		- **status** => string ["Unknown"]



	- **flags** => (Array) 
	- **created_at** => string ["2022-07-26T10:42:34.895837087Z"]



- **servers** => (Array) 
- **uptime** => f64 [0]



**Response Body Example**

```json
{
    "total_bots": 0,
    "total_servers": 0,
    "total_users": 0,
    "bots": [
        {
            "guild_count": 30,
            "description": "My description",
            "banner": "My banner or default banner url",
            "votes": 40,
            "state": 3,
            "user": {
                "id": "",
                "username": "",
                "disc": "",
                "avatar": "",
                "bot": false,
                "status": "Unknown"
            },
            "flags": [],
            "created_at": "2022-07-26T10:42:34.895837087Z"
        }
    ],
    "servers": [],
    "uptime": 0.0
}
```


**Authorization Needed** | None


