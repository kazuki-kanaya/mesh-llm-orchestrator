# mesh-llm-orchestrator

Infrastructure for running Mesh LLM asynchronously and reliably as an on-premises LLM inference server.

By preserving transparency, the design is intended to remain broadly compatible with arbitrary HTTP APIs, even though Mesh LLM is the primary target.

---

## Status

* Currently in the early implementation stage
* The repository contains design documents and an initial backend implementation
* At the moment, the system is being designed as a Redis-centered asynchronous execution platform

---

## What This Repository Aims to Build

* Accept arbitrary HTTP requests as jobs and execute them asynchronously
* Handle requests and responses transparently
* Preserve consistency through state management while assuming an imperfect queue
* Build an execution platform that can scale horizontally even in on-premises environments

---

## Overview

This system accepts arbitrary HTTP API requests as jobs rather than executing them immediately, and then executes them asynchronously through downstream workers.

Requests and responses are stored and forwarded as-is, without interpreting their contents. The design does not assume JSON parsing or field-level transformation, and HTTP headers are also treated transparently.

The queue is treated as an imperfect delivery mechanism, while Redis serves as the center of state management and consistency control.

---

## Objectives

### Build an Asynchronous HTTP Execution Platform

Enable long-running or high-load HTTP workloads to be executed safely as jobs, instead of requiring synchronous handling.

---

### Preserve Transparency

* Treat request bodies as raw bytes
* Store and return response bodies as raw bytes as well
* Avoid dependence on application-specific JSON schemas or business logic

---

### Maintain Consistency Under Distributed Failure Assumptions

The design assumes:

* worker crashes
* duplicate queue delivery
* network interruptions

Rather than trying to prevent all of these completely, the system preserves eventual consistency through state management.

---

### Ensure Scalability

Design the system so that the following can be scaled horizontally on Kubernetes in the future:

* API intake
* asynchronous execution
* worker count

---

## Scope

### In Scope

* Job creation API
* Job status lookup API
* Job result retrieval
* Enqueueing into an asynchronous queue
* HTTP execution by workers
* Redis-based state management
* Future support for incremental delivery via streaming or Pub/Sub

---

### Out of Scope

* Semantic interpretation of HTTP request contents
* Domain-specific validation
* Per-API response transformation
* Building the training stack or the model execution platform itself

---

## Technology

### Language

* Go
  Implementation of the API Gateway, Orchestrator, and Executor

---

### Communication

* HTTP
  External client communication and transparent forwarding to target APIs

* SSE
  External streaming mechanism for future real-time delivery

---

### Data and State Management

* Redis
  Management of job state, requests, responses, locks, queue, and Pub/Sub

---

### Infrastructure

* Kubernetes
  Scaling platform for the API Gateway, Orchestrator, and Executor

* On-premises environment
  Intended for local validation and self-managed deployments

---

## Expected Use Cases

* A client sends an HTTP request that may take a long time to complete
* The server immediately returns a `job_id`, and the client retrieves the result later
* Multiple workers process jobs concurrently
* Even if duplicate queue delivery occurs, state management keeps the final result consistent

---

## Design

### Overall Architecture

```text
Client
  ↓
API Gateway
  ↓
Orchestrator
  ↓
Queue (Redis List)
  ↓
Executor (Worker x N)
  ↓
Redis (state / request / response / lock)
```

---

### Layer Structure

#### 1. API Gateway

Responsibilities:

* Accept HTTP requests from clients
* Provide the job creation API
* Provide the job status lookup API
* Provide the result retrieval API

---

#### 2. Orchestrator

Responsibilities:

* Generate `job_id`
* Persist requests
* Initialize the job state as `queued`
* Enqueue the job
* Return the `job_id` to the client

Characteristics:

* Does not rely on the queue for correctness
* Treats Redis state as the source of truth

---

#### 3. Queue

Responsibilities:

* Deliver `job_id`

Characteristics:

* Uses a Redis List
* Tolerates duplicate delivery and ordering issues
* Serves only as a delivery mechanism, not as the primary reliability layer

---

#### 4. Executor

Responsibilities:

* Pull `job_id` from the queue
* Acquire a Redis lock
* Update state to `running`
* Load the stored HTTP request
* Send the request transparently to the external API
* Persist the response
* Update state to `completed` or `failed`

---

#### 5. Redis

Responsibilities:

* State management
* Request storage
* Response storage
* Lock management
* Queue
* Pub/Sub

Characteristics:

* Treated as the single source of truth in this system

---

### Component Responsibilities

| Component | Responsibility |
| --- | --- |
| API Gateway | External API surface and job status lookup |
| Orchestrator | Job creation, initial state setup, and queue insertion |
| Queue | Delivery of `job_id` |
| Executor | HTTP execution, result persistence, and state updates |
| Redis | Unified management of state, data, and locks |

---

### Execution Flow

#### 1. Job Creation

1. A client sends an HTTP request
2. The Orchestrator generates a `job_id`
3. The request is stored as-is
4. `status = queued` is set
5. The `job_id` is pushed into the queue
6. The `job_id` is returned to the client

---

#### 2. Job Execution

1. An Executor pulls a `job_id` from the queue
2. It acquires a lock using `SETNX`
3. If lock acquisition fails, it skips execution
4. `status = running` is set
5. The stored request is loaded
6. The request is forwarded transparently to the external API

---

#### 3. Completion Handling

1. The response is received
2. The response is stored
3. `status = completed` is set

On failure, `status = failed` is set instead.

---

#### 4. Result Retrieval

1. The client queries by `job_id`
2. The API Gateway or Orchestrator reads the state from Redis
3. If the job is `completed`, the response is returned

---

### State Transitions

```text
queued -> running -> completed
                 \-> failed
```

---

### Concurrency Control

```text
SETNX(lock:job:{id})
```

* Only the Executor that acquires the lock first is allowed to run the job
* Locks are expected to have TTLs

---

### Consistency Model

#### at-least-once

The design tolerates the possibility that the same job may be executed more than once.

---

#### first-writer-wins

The first `completed` result is treated as authoritative, and later results are ignored.

---

### Re-execution

If a job remains in `running` for too long, it is treated as stale and may be returned to `queued` when necessary.

---

### TTL Policy

* `completed` jobs are deleted after a fixed retention period
* `failed` jobs are also deleted after a fixed retention period

---

### Streaming Extension

The design leaves room for future support of:

* incremental response publishing from Executors through Pub/Sub
* per-`job_id` subscription by the Orchestrator
* real-time forwarding to clients

---

## Design Principles

### 1. Preserve Transparency

Requests and responses are treated as raw bytes without interpretation.

---

### 2. Do Not Trust the Queue

The queue is used as a delivery mechanism, while correctness is determined from Redis state.

---

### 3. Treat Redis as the Single Source of Truth

Job state, execution state, responses, and locks are all managed through Redis.

---

### 4. Control Execution Through Locks

Execution exclusivity is enforced through `SETNX`.

---

### 5. Do Not Return Responses Through the Queue

Responses are stored in Redis and retrieved through lookup APIs.

---

## Summary

This system is an asynchronous HTTP execution platform that uses an imperfect queue for delivery while relying on Redis-based state management to maintain final consistency.

It prioritizes transparency, asynchrony, distributed execution, and recoverability, with the goal of remaining scalable and operable in on-premises environments.
