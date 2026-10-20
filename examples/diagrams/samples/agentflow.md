---
title: Agentflow
description: The tasks, tools and decisions of AI agents, and how they hand work to each other.
beta: true
featured: 2
---

## A code review agent

```mermaid
agentflow-beta TB
  connector github["GitHub API"]
  github@{ protocol: "http", endpoint: "https://api.github.com" }

  pr["Pull request opened"]@{ shape: input }

  flow reviewer["Review agent"]
    read["Read the diff"]@{ shape: task }
    tests["run_tests"]@{ shape: tool, returns: "TestReport" }
    lint["run_linter"]@{ shape: tool, returns: "Findings" }
    guide["Style guide"]@{ shape: refdoc }
    weigh["Weigh the findings"]@{ shape: task }
    ok{"Ready to merge?"}@{ shape: decision }
    read --> tests --> weigh
    read --> lint --> weigh
    read -.- guide
    weigh --> ok
  end
  reviewer@{ instruction: "Review the change for bugs, missing tests and style." }

  approve["Approve the pull request"]@{ shape: action }
  comment["Request changes"]@{ shape: action }
  post["post_review"]@{ shape: tool, connectorRef: "github.create_review" }

  pr --> read
  ok -- yes --> approve
  ok -- no --> comment
  approve --> post
  comment --> post
```

## Support tickets handed between agents

```mermaid
agentflow-beta TB
  ticket["New support ticket"]@{ shape: input }

  flow triage["Triage agent"]
    classify["Classify the ticket"]@{ shape: task }
    kind{"Billing or technical?"}@{ shape: decision }
    classify --> kind
  end

  flow billing["Billing agent"]
    invoice["lookup_invoice"]@{ shape: tool }
    refund["issue_refund"]@{ shape: tool }
    policy["Refund policy"]@{ shape: refdoc }
    invoice --> refund
    refund -.- policy
  end

  flow tech["Technical agent"]
    logs["search_logs"]@{ shape: tool }
    kb["Known issues"]@{ shape: refdoc }
    diagnose["Diagnose the fault"]@{ shape: task }
    logs --> diagnose
    diagnose -.- kb
  end

  reply["Reply to the customer"]@{ shape: action }

  ticket --> classify
  kind -- billing --> invoice
  kind -- technical --> logs
  billing --> reply
  tech --> reply
```

## A retrieval step drawn collapsed

```mermaid
agentflow-beta TB
  question["Question about the docs"]@{ shape: input }

  flow retrieve["Retrieve passages"]
    embed["Embed the question"]@{ shape: task }
    search["vector_search"]@{ shape: tool }
    rerank["Rerank the passages"]@{ shape: task }
    embed --> search --> rerank
  end
  retrieve@{ view: "collapsed" }

  flow answer["Answer agent"]
    draft["Draft an answer"]@{ shape: task }
    cite["Cite the sources"]@{ shape: task }
    draft --> cite
  end

  send["Send the answer"]@{ shape: action }

  question --> retrieve
  retrieve --> answer
  answer --> send
```

## Agents that share a protocol

```mermaid
agentflow-beta TB
  global
    protocol["Sample handling protocol"]@{ shape: refdoc }
  end

  sample["Seawater sample arrives"]@{ shape: input, value: "Bottle 0412" }

  flow intake["Intake agent"]
    label["Read the bottle label"]@{ shape: task }
    log["record_sample"]@{ shape: tool, params: "id :: String, depth :: Number" }
    label --> log
    protocol -.- label
  end

  flow qc["Quality check agent"]
    check["Check the chain of custody"]@{ shape: task }
    usable{"Usable?"}@{ shape: decision }
    check --> usable
    protocol -.- check
  end

  store["Send to cold storage"]@{ shape: action }
  discard["Flag for disposal"]@{ shape: action }

  sample --> label
  log --> check
  usable -- yes --> store
  usable -- no --> discard
```

## A fix loop with a way out

```mermaid
agentflow-beta TB
  failure["CI run failed"]@{ shape: input }
  read["Read the build log"]@{ shape: task }
  patch["Write a patch"]@{ shape: task }
  run["run_ci"]@{ shape: tool }
  pass{"Checks pass?"}@{ shape: decision }
  open["Open a pull request"]@{ shape: action }
  handoff["Hand off to an engineer"]@{ shape: action }
  log["Build log"]@{ shape: refdoc }

  failure --> read --> patch --> run --> pass
  read -.- log
  pass -- yes --> open
  pass -- no, third try --x handoff
  pass -- no --> patch
```
