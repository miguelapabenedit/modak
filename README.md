# Modak Take Home Exercise
#### Candidate

- Miguel Benedit

## Overview
This repository contains my solution to the Modak Take Home Exercise. As part of this solution, I've developed a notification package (`modak/notification/`) in accordance with the challenge requirements.

## Time and Effort
I initiated work on this challenge on Monday 21 and finish on Wednesday 23, and invested approximately consolidated 9 hours to complete it.

## Project Structure
It's important to acknowledge that the structure of a project or library can vary significantly based on the client/project's context, data involved, and the established standards of the team. For this example, I've chosen an organizational structure that strikes a balance between adhering to go's best practices and adopting commonly seen layouts, informed by my experience.

## Decisions Made

1. I made the decision to change some types and params names like userID from string to int

2. I have decided to mock the dependencies implementation like db and cache for the sake of simplicity.

3. The existing webhook api is not required but I used in the development of the app.

I appreciate your time and effort in reviewing my solution to this challenge. If you have any feedback or questions, please feel free to reach out. Thank you.

### Desing 

Base design I draw to start the challange

![Notification Flow](notification_flow.png?raw=true "Title")


### Run the Webhook App

Required Enviroment Variables
```
 - PORT
 - CACHE_TTL
```

1- Run the DockerCompose

```
docker compose up
```

# Examples

## 1-  Valid request
The assigment example data request

```
curl --location 'http://localhost:8080/webhook/notification' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 3,
    "type": "Status",
    "message": "1000000"
}'
```
Response
```
{
   OK STATUS
}
```

## 1-  RateLimit Reached request
Once the limit is reached we recieved a 429 error code

```
curl --location 'http://localhost:8080/webhook/notification' \
--header 'Content-Type: application/json' \
--data '{
    "user_id": 3,
    "type": "Status",
    "message": "1000000"
}'
```
Response
```
 notification rate limit reached
```