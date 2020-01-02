# fairbinden_lunch_golang

<a href="https://ibb.co/JCM4516"><img src="https://i.ibb.co/nR5vzGd/Screen-Shot-2020-01-02-at-17-30-50.png" alt="Screen-Shot-2020-01-02-at-17-30-50" border="0"></a>

This program is to scrape daily lunch menu at Fairbinden blog and extract the main information to send to your Slack channel during weekdays.

Tech Stack
- Golang
- Colly
- CloudFunction


1. Set up channel urls for productin and staging as export variables before cloud function is deployed and triggered  
Examples
```
export Channel_PRD = "https://hooks.slack.com/services/xxxxxx/xxxxx/xxxxxxxxxxxxxxxxxxxxx"
export Channel_STG = "https://hooks.slack.com/services/qqqqqq/sssss/sssssssssssssssssssss"
```

2. Deploy the code to Cloud Function with Golang 1.1 runtime

3. Set up a cloud scheduler to trigger the function at 10 a.m. every Weekday 