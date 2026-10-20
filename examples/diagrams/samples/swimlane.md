---
title: Swimlane
description: A process laid out in lanes, one for each person or team that does its steps.
beta: true
featured: 2
---

## Incident response across teams

```mermaid
swimlane-beta TB
  subgraph Monitoring
    alert([Error rate spikes])
  end

  subgraph oncall [On-call]
    ack[Acknowledge the page]
    assess{Customers affected?}
    mitigate[Roll back the deploy]
  end

  subgraph commander [Commander]
    declare[Declare an incident]
    review[Run the postmortem]
  end

  subgraph Support
    status[Post a status update]
    close[Close the status page]
  end

  alert --> ack --> assess
  assess -->|No| mitigate
  assess -->|Yes| declare
  declare --> status
  declare --> mitigate
  mitigate --> close --> review
```

## From pull request to production

```mermaid
swimlane-beta LR
  subgraph Developer
    open[Open a pull request]
    revise[Address comments]
  end

  subgraph Reviewer
    review{Approve?}
  end

  subgraph CI [CI pipeline]
    test[Run the tests]
    build[Build the image]
  end

  subgraph Release
    stage[Deploy to staging]
    prod[Deploy to production]
  end

  open --> review
  review -->|Changes requested| revise --> review
  review -->|Approved| test --> build --> stage --> prod
```

## An expense report with named lanes

```mermaid
swimlane-beta LR
  subgraph emp [Employee]
    submit[Submit the report]
    fix[Attach missing receipts]
  end

  subgraph mgr [Manager]
    approve{Within policy?}
  end

  subgraph fin [Finance team]
    audit[Check the receipts]
    pay[Reimburse by payroll]
  end

  submit --> approve
  approve -->|Yes| audit
  approve -->|Missing receipts| fix --> submit
  audit --> pay
```

## A kelp harvest, lanes running down

```mermaid
swimlane-beta TB
  subgraph crew [Farm crew]
    check[Check the longlines]
    cut[Cut the kelp]
  end

  subgraph plant [Processing]
    rinse[Rinse and sort]
    grade{Food grade?}
    dry[Dry and pack]
    compost[Set aside as fertilizer]
  end

  subgraph Sales
    ship[Ship to restaurants]
  end

  check --> cut --> rinse --> grade
  grade -->|Yes| dry --> ship
  grade -->|No| compost
```

## Returning a purchase, from start to end

```mermaid
swimlane-beta LR
  subgraph Customer
    start([Request a return])
    label[Print the label]
    refunded((Refunded))
  end

  subgraph Store
    eligible{Within 30 days?}
    decline[Explain the policy]
  end

  subgraph Warehouse
    receive[Receive the parcel]
    inspect[Inspect the item]
  end

  start --> eligible
  eligible -->|Yes| label --> receive --> inspect --> refunded
  eligible -->|No| decline
```
