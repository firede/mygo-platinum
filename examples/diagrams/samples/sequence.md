---
title: Sequence
description: Participants side by side, with the messages they send each other in the order they are sent.
beta: false
featured: 2
---

## Signing in with OAuth

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant App as Web app
    participant Auth as Identity provider
    participant API as Profile API
    User->>App: Click "Sign in"
    App->>Auth: Redirect with client ID
    Auth->>User: Show the sign-in page
    User->>Auth: Enter credentials
    Auth-->>App: Redirect with a code
    App->>+Auth: Exchange the code
    Auth-->>-App: Access and ID tokens
    App->>+API: GET /me
    API-->>-App: Profile
    App-->>User: Dashboard
```

## Paying for an order

```mermaid
sequenceDiagram
    participant Shop as Checkout service
    participant Pay as Payment provider
    participant DB as Orders database
    Shop->>+Pay: Authorize $84.20
    alt Card approved
        Pay-->>Shop: Authorization ID
        Shop->>DB: Mark the order as paid
        Shop->>Pay: Capture the payment
    else Card declined
        Pay-->>Shop: Declined, insufficient funds
        Shop->>DB: Mark the order as failed
    else No answer in 10 seconds
        Shop->>Pay: Void the authorization
    end
    deactivate Pay
    opt Customer opted in
        Shop->>DB: Save the card token
    end
```

## Loops and parallel steps in a rollout

```mermaid
sequenceDiagram
    participant CD as Deploy pipeline
    participant LB as Load balancer
    participant Web as Web servers
    participant Mon as Monitoring
    loop For each batch of 10 servers
        CD->>LB: Drain the batch
        CD->>Web: Install release 4.12
        par Health checks
            CD->>Web: GET /healthz
        and Error budget
            CD->>Mon: Query the error rate
        end
        CD->>LB: Put the batch back in rotation
    end
    break Error rate above 1%
        CD->>Web: Roll back to 4.11
    end
```

## Notes and boxes for a ferry booking

```mermaid
sequenceDiagram
    actor Rider
    box Ferry company
        participant Site as Booking site
        participant Sched as Schedule service
    end
    box Partner
        participant Port as Harbor authority
    end
    Rider->>Site: Find Saturday crossings
    Site->>Sched: List sailings
    Note right of Sched: Times shift with the tide
    Sched->>Port: Check berth availability
    Port-->>Sched: Berths 2 and 4 open
    Sched-->>Site: 7:40, 10:15 and 1:30
    Note over Rider,Site: Seats are held for 15 minutes
    Rider->>Site: Book the 10:15 sailing
    Site-->>Rider: Ticket with a QR code
```

## Participant types in an upload pipeline

```mermaid
sequenceDiagram
    actor Editor
    participant API@{ "type": "boundary" } as Upload API
    participant Jobs@{ "type": "queue" } as Job queue
    participant Worker@{ "type": "control" } as Thumbnailer
    participant Store@{ "type": "database" } as S3 bucket
    Editor->>API: POST /photos
    API->>Store: Put the original
    API-)Jobs: Enqueue a resize job
    API-->>Editor: 202 Accepted
    Jobs-)Worker: Deliver the job
    Worker->>Store: Read the original
    Worker->>Store: Write three thumbnails
    Worker--xJobs: Acknowledge
```

## Creating and destroying participants

```mermaid
sequenceDiagram
    participant Sched as Scheduler
    participant Pool as Runner pool
    Sched->>Pool: Request a runner for build 2817
    create participant Runner as Ephemeral runner
    Pool->>Runner: Start a fresh VM
    Runner->>Sched: Ready
    Sched->>Runner: Run the build
    critical Upload artifacts
        Runner->>Sched: Upload the test report
    option Upload fails
        Runner->>Sched: Retry once, then fail the build
    end
    destroy Runner
    Pool-xRunner: Tear down the VM
    Note over Sched,Pool: Each build gets a clean machine
```

## Booking a table, in Japanese

```mermaid
sequenceDiagram
    autonumber
    actor 客 as お客様
    participant アプリ
    participant 店 as お店
    客->>アプリ: 日時と人数を選ぶ
    アプリ->>店: 空席を確認
    店-->>アプリ: 窓側が空いています
    アプリ-->>客: 予約内容を表示
    客->>アプリ: 予約を確定
    アプリ->>店: 予約を登録
    Note over アプリ,店: 前日にリマインド
    アプリ-->>客: 確認メールを送信
```
