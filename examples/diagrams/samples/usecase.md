---
title: Use case
description: The actors outside a system and the use cases they take part in.
beta: true
featured: 2
---

## A public library's lending system

```mermaid
usecase-beta
direction LR
actor Member
actor Librarian
systemBoundary library["Lending system"]
  Search("Search the catalog")
  Borrow("Borrow a book")
  Renew("Renew a loan")
  Hold("Place a hold")
  Return("Check in a return")
  Catalog("Add a new title")
  Card("Verify the card")
end
Member --> Search
Member --> Borrow
Member --> Renew
Member --> Hold
Librarian --> Return
Librarian --> Catalog
Borrow ..> : include Card
Renew ..> : include Card
```

## A ride-hailing app

```mermaid
usecase-beta
direction LR
actor Rider
actor Driver
actor Payments("Payment provider")
systemBoundary app["Ride app"]
  Request("Request a ride")
  Accept("Accept a trip")
  Track("Track the car")
  Pay("Pay for the trip")
  Rate("Rate the trip")
end
Rider --> Request
Rider --> Track
Rider --> Rate
Driver --> Accept
Driver --> Rate
Pay --> Payments
Request ..> : include Pay
```

## Help desk roles that build on each other

```mermaid
usecase-beta
direction TB
actor Customer
actor Agent("Support agent")
actor Lead("Team lead")
Open("Open a ticket")
Answer("Answer a ticket")
Escalate("Escalate a bug")
Reassign("Reassign a ticket")
Customer --> Open
Agent --> Answer
Lead --> Reassign
Lead --|> Agent
Escalate ..> : extend Answer
```

## An insurance claim as business use cases

```mermaid
usecase-beta
direction LR
actor Policyholder@{ business: true }
actor Adjuster("Claims adjuster")@{ type: hollow, business: true } <<Employee>>
systemBoundary claims["Claims office"]@{ type: package }
  File("File a claim")@{ business: true }
  Assess("Assess the damage")@{ business: true } <<Core>>
  Settle("Settle the claim")@{ business: true }
end
note for Assess "Photos are required for claims over $2,000"
Policyholder --> File
Adjuster --> Assess
Adjuster --> Settle
```

## A tide gauge reporting its readings

```mermaid
usecase-beta
direction TB
actor Gauge("Tide gauge")@{ type: hollow }
actor Forecaster("Harbor forecaster")
Upload("Upload a reading")
Check("Check the water level")
json Reading@{
  "station": "Pier 7",
  "levelMeters": 2.84,
  "recordedAt": "06:15 UTC",
  "battery": "ok"
}
Gauge --> Upload
Forecaster --> Check
Upload --> Reading
Check --> Reading
```
