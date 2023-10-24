# SendX-Backend-Assignment IIT2020190

Webpage Crawler. Server provide the service to crawl the webpage and obtain all the links available in the page. In this server. This server is designed to support retry feature, concurrent crawling, prioritizing paying customers, admin controls.

## Getting Started
These instructions will help you set up the project on your local machine for testing purposes.

### Prerequisities
- [Go]()

### Installation
1. Clone the repository:
   ```bash
        git clone https://github.com/Malyala-Meghamsh/sendx-backend-IIT2020190
    ```
2. Change the directory
    ```bash
        cd SENDX_PROJECT
    ```
3. Installing the dependencies
    ``` bash
        go get
    ```
4. Final Step
    ```bash
        go run main.go
    ```
Server will be starting in the port 3000.
Follow this link : 
http://localhost:3000/

## Features

### General features

1. **In the Landing page** (localhost:3000) with a search bar we enter the URL to be crawled and it crawls the links available in that page.

2. **Cache Mechanism** : Server checks if the url was crawled in last 60 minutes if found then the crawled page stored on the disk is read and returned back to the user.

3. **Retry Feature** : As page is not availabe all the times, we will be retring those pages.

4. **Subscription Feature** : We have two types of users, paying and non paying customers. Priority is always given to paying customers. (More crawlers, More Retries, More Crawling Speed)

5. **Concurrency Crawling** : The application is capable of crawling multiple pages concurrently. We have multiple workers for maximum throughput. URL maybe picked by any worker.

### Admin Controls
1. **Rate Limiting** : Admin can set some limit for number of pages crawled.
2. **Control Number of Workers** 

## API Endpoints

### For Fetching, Web Crawlng
- Endpoint for crawling the URLs 
    `http://localhost:3000/query?url={url}`
    For paid customers you can append `&isPaidCustomer=1` and use this endpoint 
    `http://localhost:3000/query?url={url}&isPaidCustomer=1`

### Admin Control Panel

- To set the number of crawler worker for Paying and Non Paying Workers
    `http://localhost:3000/api/setCrawlers?paidWorkers={setNumberOfWorkersforPaidUsers}`
    
    You can also Append this `&unpaidWorkers={workersCountforUnpayingUsers}` or you can even just set the Workers for unpaid users.

- To Adjust the crawling speed per hour per worker.
    `http://localhost:3000/api/setSpeed?speed={setSpeed}`
    You can set the max pages that a worker can crawl using this api.

 ## Demo Of Project
 Link[]
