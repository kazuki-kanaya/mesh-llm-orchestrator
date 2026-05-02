# mesh-llm-orchestrator

A design repository for an orchestration platform that makes Mesh LLM practical for production-oriented, on-premises distributed LLM inference.

---

## Status

* Currently in the planning and design phase before implementation
* At this stage, the repository contains only architectural direction and design notes
* The initial architecture separates a Kubernetes-based control plane from the execution plane

---

## What This Repository Aims to Build

* Add the control layer required for operations without exposing Mesh LLM directly
* Safely process inference requests from multiple users as asynchronous jobs
* Stream long-running inference results to preserve user experience
* Assume a single Mesh LLM cluster at first, while keeping the design extensible for multiple execution targets later

---

## Overview

This system uses Mesh LLM as the inference engine and builds an orchestration layer around it to provide asynchronous, real-time, and scalable distributed LLM inference in on-premises environments.

Mesh LLM is responsible for distributed inference across multiple GPU nodes, but it does not fully provide the operational capabilities needed in production, such as job management, queueing, streaming control, API exposure, and load control. This system complements those gaps and extends Mesh LLM into a form suitable for production use.

The orchestration layer is designed to run on Kubernetes as a scalable control platform. At the current stage, the system assumes a single Mesh LLM cluster as the execution target, while leaving room to expand toward multiple execution targets in the future.

---

## Objectives

### Build a Practical Local LLM Platform

Provide a platform that integrates multiple GPU servers in laboratories or on-premises environments and can handle inference requests from multiple users.

---

### Enable Asynchronous and Real-Time Inference

* Process inference requests as asynchronous jobs
* Stream progress and results even for long-running workloads

---

### Use the Inference Platform Efficiently

* Let Mesh LLM distribute execution internally across multiple nodes
* Smooth load through queueing
* Support per-job model selection

---

### Ensure Scalability

* The orchestration layer scales horizontally on Kubernetes
* The structure should remain extensible for future support of multiple execution targets and router introduction

---

## Scope

### In Scope

* Inference job submission API
* Asynchronous queueing
* Streaming delivery of inference results
* Persistence of job state
* Dispatching execution requests to a single Mesh LLM cluster
* Forwarding model-specified inference requests

---

### Out of Scope

* Training foundation models themselves
* Modifying the internal implementation of Mesh LLM
* Immediate automatic scaling of the GPU cluster
* Multi-tenant billing or advanced organizational management
* Providing a complete end-to-end MLOps pipeline

---

## Technology

### Languages

* Go
  API Gateway, Job Service, and Execution Service

* Python
  Auxiliary processing where needed

---

### Communication

* HTTP
  External communication with clients

* WebSocket
  Real-time delivery to clients

* gRPC
  Internal communication between orchestration services

---

### Inference Backend

* Mesh LLM
  Distributed inference and model execution within a single cluster

---

### Asynchronous Processing

* NATS JetStream
  Inference job queueing and asynchronous processing

---

### Data Management

* PostgreSQL
  Persistent storage for job state and history

---

### Infrastructure

* Kubernetes
  Execution and scaling platform for the orchestration layer

* Current state
  A single Mesh LLM cluster in a local environment

* Future state
  Expansion toward multiple execution targets and Kubernetes-managed infrastructure

---

## Expected Use Cases

* Multiple users within a lab submit inference jobs concurrently
* The system handles not only short one-shot inference, but also long-running generation jobs
* Users create jobs through an HTTP API and receive progress or partial results over WebSocket
* Users specify a `model` per job, and execution runs on a single Mesh LLM cluster
* The orchestration layer is internally split into multiple services that communicate over gRPC

---

## Design

### Overall Architecture

```text
Client
  ↓ (HTTP / WebSocket)
API Gateway (Kubernetes)
  ↓ (gRPC)
Job Service (Kubernetes)
  ↓ (JetStream / PostgreSQL)
Job / Queue Layer
  ↓ (gRPC)
Execution Service (Kubernetes)
  ↓
Mesh LLM cluster
  ↓ (streaming)
Streaming / Result
  ↓
Client
```

---

### Layer Structure

#### 1. Orchestration Layer (running on Kubernetes)

Components:

* API Gateway
* Job Service
* Execution Service
* Streaming control

Responsibilities:

* Accept requests and return responses
* Manage job state and persistence
* Dispatch execution requests to Mesh LLM
* Control the end-user experience

Characteristics:

* Runs as multiple Pods on Kubernetes
* Supports horizontal scaling by increasing or decreasing Pods
* Uses gRPC for internal service-to-service communication
* Uses a single Mesh LLM cluster as the execution target at the current stage

---

#### 2. Execution Layer

Components:

* Mesh LLM

Responsibilities:

* Execute distributed inference
* Handle inter-node communication
* Run models

Characteristics:

* Handles distribution across GPU nodes inside the cluster
* Executes inference based on the specified model
* Can later expand to a multi-cluster configuration

---

#### 3. Infrastructure Layer

Components:

* Mesh LLM cluster
* Kubernetes cluster

Responsibilities:

* Provide the inference execution platform
* Manage Pod placement and scaling

---

### Component Responsibilities

| Component | Responsibility |
| --- | --- |
| API Gateway | External interface, authentication, and request intake |
| Job Service | Job state management, JetStream integration, and persistence |
| Execution Service | Send inference requests to Mesh LLM and receive results |
| Mesh LLM | Execution of distributed inference |

---

### Where Kubernetes Is Used

#### 1. Scaling the Orchestration Layer

* Deploy API Gateway, Job Service, and Execution Service as Pods
* Scale horizontally using mechanisms such as HPA

---

#### 2. Availability

* Automatic Pod restarts
* Rolling updates

---

#### 3. Future Expansion of Execution Targets

* Support multiple Mesh LLM clusters
* Add new execution targets
* Introduce a Router when needed

---

### Design Principles

#### 1. Wrap Mesh LLM

Mesh LLM is not exposed directly to external users; it is wrapped by the orchestration layer.

---

#### 2. Separate GPU Resource Management

GPU node management and distributed execution are the responsibility of Mesh LLM.

* The orchestrator does not directly handle GPU-node-level details
* The orchestrator treats a single Mesh LLM cluster as its execution target

---

#### 3. Forward Model Selection Transparently

The per-job `model` specification is forwarded by the orchestrator to Mesh LLM as-is.

---

#### 4. Async-First Design

All inference requests go through JetStream.

---

#### 5. Separation of Responsibilities

Each internal service has a single responsibility, and services communicate over gRPC.

---

#### 6. Design for Scalability from the Start

* The orchestration layer scales on Kubernetes
* The structure should leave room for future support of multiple execution targets and Router introduction

---

### Current State vs. Future State

| Item | Current | Future |
| --- | --- | --- |
| Orchestration Layer | Kubernetes | Kubernetes |
| Execution target configuration | Single Mesh LLM cluster | Multiple execution targets |
| Model selection | Forward the `model` field as-is | Extend through policy or Router |
| CPU-side scaling | Kubernetes | Kubernetes |
| Execution-side scaling | Depends on Mesh LLM | Expand to multiple execution targets or Kubernetes-managed infrastructure |

---

## Summary

This system uses Mesh LLM as the inference engine and extends it into an operationally viable platform for distributed LLM inference by introducing an orchestration layer running on Kubernetes.

At the current stage, the system assumes a single Mesh LLM cluster and places job management, streaming, cancellation, and persistence on the orchestrator side. The design remains extensible for future support of multiple execution targets and Router introduction.
