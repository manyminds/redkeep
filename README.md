# Red Keep 
[![Build Status](https://travis-ci.org/manyminds/redkeep.svg?branch=master)](https://travis-ci.org/manyminds/redkeep)
[![Coverage Status](https://coveralls.io/repos/github/manyminds/redkeep/badge.svg?branch=master)](https://coveralls.io/github/manyminds/redkeep?branch=master)

## A MongoDB redundancy keeper.

This project automatically tracks changes and adds redundant fields on references.

![The Westeros Red Keep](http://awoiaf.westeros.org/images/thumb/2/22/Red_Keep.jpg/800px-Red_Keep.jpg)

Source: http://awoiaf.westeros.org/index.php/File:Red_Keep.jpg

# Functionality

Redkeep runs in the background and denormalizes your references for you. Just define what you want to get denormalized, it's easy.

# Installation and Usage

```
go get -d -u github.com/manyminds/redkeep/redkeepcli
```

Will install the redkeepcli client. Have a look at the example configuration to see how to configure redkeep the way you want.
Let's have a look at the configuration of one watch in detail:
```json
    {
      "trackCollection": "application.user",
      "trackFields": ["name", "username"],
      "targetCollection": "application.answer",
      "targetNormalizedField": "meta",
      "triggerReference": "user",
      "behaviourSettings": {
        "cascadeDelete": false
      }
    }
```

This will watch for changes in the database *application* and the collection *user*. If a new *answer* will be inserted with a reference to 
*application.user* the fields *name* and *username* will automatically be stored in the newly created *answer* as the fields *meta.name* and *meta.username*.
