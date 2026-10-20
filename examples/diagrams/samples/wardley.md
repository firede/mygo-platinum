---
title: Wardley map
description: The components of a value chain, placed by how visible they are and how far they have evolved.
beta: true
featured: 2
---

## A video streaming service

```mermaid
wardley-beta
title Video streaming service

anchor Viewer [0.95, 0.70]
component Watch a show [0.86, 0.62]
component Recommender [0.72, 0.38] (build)
component Video player [0.68, 0.72] (buy)
component Catalog [0.55, 0.55]
component Viewing history [0.45, 0.30]
component Encoding [0.38, 0.65]
component CDN [0.25, 0.85] (outsource)
component Compute [0.10, 0.92] (market)

Viewer -> Watch a show
Watch a show -> Recommender
Watch a show -> Video player
Recommender -> Catalog
Recommender -> Viewing history
Video player -> CDN
Catalog -> Encoding
Encoding -> Compute
Viewing history -> Compute
CDN -> Compute

evolve Recommender 0.55
```

## A seaweed farm's value chain

```mermaid
wardley-beta
title Seaweed farm

anchor Restaurants [0.95, 0.55]
anchor Fertilizer buyers [0.92, 0.84]
component Fresh kelp [0.80, 0.45]
component Dried kelp [0.76, 0.76]
component Harvest [0.60, 0.50]
component Drying [0.55, 0.84]
component Grow lines [0.42, 0.35]
component Seed stock [0.30, 0.15] (inertia)
component Boats [0.30, 0.86]
component Cold chain [0.14, 0.72]

Restaurants -> Fresh kelp
Fertilizer buyers -> Dried kelp
Fresh kelp -> Harvest
Fresh kelp -> Cold chain
Dried kelp -> Drying
Drying -> Harvest
Harvest -> Grow lines
Harvest -> Boats
Grow lines -> Seed stock

note "Hatchery skills are rare" [0.10, 0.04]
```

## Hosting as a pipeline

```mermaid
wardley-beta
title Where the app runs

anchor Customer [0.95, 0.75]
component Web app [0.80, 0.55]
component Database [0.55, 0.82] (buy)
component Hosting [0.35, 0.50]

pipeline Hosting {
  component "Own servers" [0.10]
  component "VMs" [0.38]
  component "Containers" [0.65]
  component "Functions" [0.92]
}

Customer -> Web app
Web app -> Database
Web app -> Hosting
```

## Forces on a data platform

```mermaid
wardley-beta
title Data platform

anchor Analysts [0.95, 0.55]
component Dashboards [0.82, 0.72]
component Data models [0.65, 0.35]
component Warehouse [0.40, 0.55] (buy)
component Legacy database [0.22, 0.48] (inertia)
component Cloud storage [0.12, 0.90]

Analysts -> Dashboards
Dashboards -> Data models
Data models -> Warehouse
Data models -> Legacy database
Warehouse -> Cloud storage

evolve Warehouse 0.80
accelerator "Managed warehouses" [0.64, 0.64]
deaccelerator "Data residency rules" [0.30, 0.10]
```

## Machine learning on custom stages

```mermaid
wardley-beta
title Fraud detection
evolution Novel -> Emerging -> Good practice -> Best practice

anchor Payments team [0.95, 0.60]
component Fraud score [0.80, 0.35]
component Model [0.62, 0.25]
component Feature store [0.48, 0.55]
component Labeled cases [0.40, 0.15]
component GPU training [0.20, 0.80]

Payments team -> Fraud score
Fraud score -> Model
Model -> Feature store
Model -> Labeled cases
Model -> GPU training

annotations [0.22, 0.03]
annotation 1,[0.69, 0.21] "Retrained weekly"
annotation 2,[0.33, 0.11] "Labeled by hand"
```
