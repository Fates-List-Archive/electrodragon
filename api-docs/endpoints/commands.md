
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

## Add Command
### POST `https://fates-api.select-list.xyz`/bots/{id}/commands

Creates a command.

The ``id`` here must be the bot id you wish to add the command for

**This performs a *upsert* meaning it will either create or update 
the command depending on its ``name``.**

**Only post up to 10-20 commands at a time, otherwise requests may be truncated
or otherwise fail with odd errors.  If you have more than this, then perform 
multiple requests**

``target_type`` is a [TargetType](./enums#targettype)
**Query Parameters**

- **target_type** => i32 [0]




**Path Parameters**

- **id** => i64 [0]




**Request Body**

- **commands** => (Array) Struct BotCommand 
	- **cmd_type** => i32 [0]
	- **groups** => (Array) 
	- **name** => string [""]
	- **vote_locked** => bool [false]
	- **description** => string [""]
	- **args** => (Array) 
	- **examples** => (Array) 
	- **premium_only** => bool [false]
	- **notes** => (Array) 
	- **doc_link** => None (unknown value type)
	- **id** => None (unknown value type)
	- **nsfw** => bool [false]






**Request Body Example**

```json
{
    "commands": [
        {
            "cmd_type": 0,
            "groups": [],
            "name": "",
            "vote_locked": false,
            "description": "",
            "args": [],
            "examples": [],
            "premium_only": false,
            "notes": [],
            "doc_link": null,
            "id": null,
            "nsfw": false
        }
    ]
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


**Authorization Needed** | [Bot](#authorization)


## Delete Commands
### DELETE `https://fates-api.select-list.xyz`/bots/{id}/commands

DELETE a command.

The ``id`` here must be the bot id you wish to add the command for

``names`` and ``ids`` must be a ``|`` seperated list of ``names`` or valid
UUIDs in the case of ids. Bad names/ids will be ignored
**Query Parameters**

- **nuke** => (Optional) bool [false]
- **names** => (Optional) string ["command name|command name 2"]
- **ids** => (Optional) string ["id 1|id 2"]




**Path Parameters**

- **id** => i64 [0]





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


**Authorization Needed** | [Bot](#authorization)


