
**API URL**: ``https://fates-api.select-list.xyz``

**Widgets Documentation:** ``https://fates-lynx.select-list.xyz/widgets`` 

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

## Get User From ID
### GET `https://fates-api.select-list.xyz`/blazefire/{id}

Gets a User object given a ID. 

Internally is used by client for extra owner rendering etc.

May be used by our partners to get user information.

**Path Parameters**

- **id** => i64 [0]





**Response Body**

- **id** => string [""]
- **username** => string [""]
- **disc** => string [""]
- **avatar** => string [""]
- **bot** => bool [false]
- **status** => string ["Unknown"]



**Response Body Example**

```json
{
    "id": "",
    "username": "",
    "disc": "",
    "avatar": "",
    "bot": false,
    "status": "Unknown"
}
```


**Authorization Needed** | None


## Get User Perms
### GET `https://fates-api.select-list.xyz`/baypaw/perms/{id}

Gets the permissions of a user from Baypaw (our microservices handling user fetching and permissions)

Internally is used by client for extra owner rendering etc.

**Path Parameters**

- **id** => i64 [0]





**Response Body**

- **id** => string [""]
- **staff_id** => string [""]
- **perm** => f32 [0]
- **fname** => string [""]



**Response Body Example**

```json
{
    "id": "",
    "staff_id": "",
    "perm": 0.0,
    "fname": ""
}
```


**Authorization Needed** | None


## Get Profile
### GET `https://fates-api.select-list.xyz`/profiles/{id}
Gets a user profile.

**Path Parameters**

- **id** => i64 [0]





**Response Body**

- **user** => Struct User 
	- **id** => string [""]
	- **username** => string [""]
	- **disc** => string [""]
	- **avatar** => string [""]
	- **bot** => bool [false]
	- **status** => string ["Unknown"]



- **connections** => (Array) Struct FrostpawUserConnection 
	- **client** => Struct FrostpawClient 
		- **id** => string [""]
		- **name** => string [""]
		- **domain** => string [""]
		- **privacy_policy** => string [""]
		- **owner** => Struct User 
			- **id** => string [""]
			- **username** => string [""]
			- **disc** => string [""]
			- **avatar** => string [""]
			- **bot** => bool [false]
			- **status** => string ["Unknown"]



		- **verified** => bool [false]



	- **expires_on** => string ["1970-01-01T00:00:00Z"]
	- **repeats** => i64 [0]



- **bots** => (Array) 
- **description_raw** => string [""]
- **description** => string [""]
- **profile_css** => string [""]
- **user_css** => string [""]
- **vote_reminder_channel** => (Optional) string ["939123825885474898"]
- **packs** => (Array) 
- **state** => i32 [0]
- **site_lang** => string [""]
- **action_logs** => (Array) Struct ActionLog 
	- **user_id** => string [""]
	- **bot_id** => string [""]
	- **action** => i32 [0]
	- **action_time** => string ["1970-01-01T00:00:00Z"]
	- **context** => (Optional) string ["Some context as to why the action happened"]



- **user_experiments** => (Array) 
- **flags** => (Array) 
- **extra_links** => Map (key/value)  






**Response Body Example**

```json
{
    "user": {
        "id": "",
        "username": "",
        "disc": "",
        "avatar": "",
        "bot": false,
        "status": "Unknown"
    },
    "connections": [
        {
            "client": {
                "id": "",
                "name": "",
                "domain": "",
                "privacy_policy": "",
                "owner": {
                    "id": "",
                    "username": "",
                    "disc": "",
                    "avatar": "",
                    "bot": false,
                    "status": "Unknown"
                },
                "verified": false
            },
            "expires_on": "1970-01-01T00:00:00Z",
            "repeats": 0
        }
    ],
    "bots": [],
    "description_raw": "",
    "description": "",
    "profile_css": "",
    "user_css": "",
    "vote_reminder_channel": "939123825885474898",
    "packs": [],
    "state": 0,
    "site_lang": "",
    "action_logs": [
        {
            "user_id": "",
            "bot_id": "",
            "action": 0,
            "action_time": "1970-01-01T00:00:00Z",
            "context": "Some context as to why the action happened"
        }
    ],
    "user_experiments": [],
    "flags": [],
    "extra_links": {}
}
```


**Authorization Needed** | None


## Update Profile
### PATCH `https://fates-api.select-list.xyz`/profiles/{id}

Edits a user profile.

``user`` can be completely empty valued but the keys present in a User must
be present

**Path Parameters**

- **id** => i64 [0]




**Request Body**

- **user** => Struct User 
	- **id** => string [""]
	- **username** => string [""]
	- **disc** => string [""]
	- **avatar** => string [""]
	- **bot** => bool [false]
	- **status** => string ["Unknown"]



- **connections** => (Array) 
- **bots** => (Array) 
- **description_raw** => string [""]
- **description** => string [""]
- **profile_css** => string [""]
- **user_css** => string [""]
- **vote_reminder_channel** => (Optional) string ["939123825885474898"]
- **packs** => (Array) 
- **state** => i32 [0]
- **site_lang** => string [""]
- **action_logs** => (Array) Struct ActionLog 
	- **user_id** => string [""]
	- **bot_id** => string [""]
	- **action** => i32 [0]
	- **action_time** => string ["1970-01-01T00:00:00Z"]
	- **context** => (Optional) string ["Some context as to why the action happened"]



- **user_experiments** => (Array) 
- **flags** => (Array) 
- **extra_links** => Map (key/value)  






**Request Body Example**

```json
{
    "user": {
        "id": "",
        "username": "",
        "disc": "",
        "avatar": "",
        "bot": false,
        "status": "Unknown"
    },
    "connections": [],
    "bots": [],
    "description_raw": "",
    "description": "",
    "profile_css": "",
    "user_css": "",
    "vote_reminder_channel": "939123825885474898",
    "packs": [],
    "state": 0,
    "site_lang": "",
    "action_logs": [
        {
            "user_id": "",
            "bot_id": "",
            "action": 0,
            "action_time": "1970-01-01T00:00:00Z",
            "context": "Some context as to why the action happened"
        }
    ],
    "user_experiments": [],
    "flags": [],
    "extra_links": {}
}
```


**Response Body**

- **done** => bool [true]
- **reason** => None (unknown value type)
- **context** => None (unknown value type)



**Response Body Example**

```json
{
    "done": true,
    "reason": null,
    "context": null
}
```


**Authorization Needed** | [User](#authorization)


## Receive Profile Roles
### PUT `https://fates-api.select-list.xyz`/profiles/{id}/roles
Gives user roles on the Fates List support server

**Path Parameters**

- **id** => i64 [0]





**Response Body**

- **bot_developer** => bool [false]
- **certified_developer** => bool [false]



**Response Body Example**

```json
{
    "bot_developer": false,
    "certified_developer": false
}
```


**Authorization Needed** | [User](#authorization)


