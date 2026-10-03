# planka-notify
Sends an email when a planka task is due today. It's a hacky, thrown together thing that panics on every error and doesn't check if the required env vars are set. But it works!

Behaviour:
- If a task is due today, it will send an email with the task's name as the subject and a link to it in the body
- When it checks for same date, it compare's the task's due date and the host machine's, with both timezones set to `$TIMEZONE`
- It only sends for the same day, it will not resend old ones

Requires the following environment variables:
```
BASE_URL           # no trailing slash
PLANKA_TOKEN
TIMEZONE           # leave blank for UTC
GMAIL_TO_ADDRESS
GMAIL_FROM_ADDRESS
GMAIL_APP_PASSWORD
```