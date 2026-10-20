---
title: Cynefin
description: Situations sorted into the domains of the Cynefin framework, from clear to chaotic.
beta: true
featured: 1
---

## Sorting a platform team's backlog

```mermaid
cynefin-beta
  title Platform team backlog

  clear
    "Rotate TLS certificates"
    "Bump the Go version"

  complicated
    "Tune Postgres for the new load"
    "Plan the region failover"

  complex
    "Cut cloud spend by a third"
    "Adopt a service mesh"

  chaotic
    "Payments are down in production"

  confusion
    "Flaky tests nobody owns"
```

## How a new product's practices drift

```mermaid
cynefin-beta
  title A product's first year

  complex
    "Find what customers pay for"

  complicated
    "Scale the onboarding flow"

  clear
    "Weekly release checklist"

  chaotic
    "Launch-day outage"

  complex --> complicated : "A pattern emerges"
  complicated --> clear : "Experts write it down"
  clear --> chaotic : "Complacency"
  chaotic --> complex : "Service restored"
```

## Running a kelp hatchery

```mermaid
cynefin-beta
  title Kelp hatchery season

  clear
    "Log tank temperatures"

  complicated
    "Choose a seed strain"

  complex
    "Spores settle unevenly"

  chaotic
    "A tank heater fails"

  chaotic --> clear : "Spare heater on hand"
  complex --> complicated
```
