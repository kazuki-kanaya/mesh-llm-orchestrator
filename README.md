# mesh-llm-orchestrator

A design repository for an orchestration platform that enables production-oriented, on-premises distributed LLM inference with Mesh LLM.

---

## Status

* Currently in the planning and design phase, prior to implementation
* At this stage, the repository contains only the architectural direction and design notes
* The initial architecture separates the Kubernetes-based control plane from the execution plane running on fixed GPU nodes

---

## What This Repository Aims to Build

* Add the control layer required for real-world operations without exposing Mesh LLM directly
* Safely handle inference requests from multiple users as asynchronous jobs
* Stream progress and results for long-running inference workloads to preserve user experience
* Keep the current fixed GPU node setup easy to evolve toward Kubernetes-managed infrastructure in the future

---

## Overview

This system uses Mesh LLM as the inference engine and adds an orchestration layer around it to provide asynchronous, real-time, and scalable distributed LLM inference in on-premises environments.

Mesh LLM is responsible for distributed inference across multiple GPU nodes, but it does not fully provide the operational capabilities needed in production, such as job management, queueing, streaming control, API exposure, and load control. This system complements those gaps and extends Mesh LLM into a form that is suitable for production use.

The orchestration layer is designed to run on Kubernetes as a scalable control platform. GPU resources, meanwhile, are currently treated as fixed nodes, with a structure that can later transition to Kubernetes-based dynamic scaling.

---

## Objectives

### Build a Practical Local LLM Platform

Provide a platform that integrates multiple GPU servers in laboratories or on-premises environments and can process inference requests from multiple users.

---

### Enable Asynchronous and Real-Time Inference

* Process inference requests as asynchronous jobs
* Stream progress and results back to clients, even for long-running workloads

---

### Use GPU Resources Efficiently

* Distribute execution across multiple nodes
* Smooth load through queueing
* Select the most suitable node through the Router

---

### Ensure Scalability

* The orchestration layer scales horizontally on Kubernetes
* GPU workers can later support node scaling through Kubernetes

---

## Scope

### In Scope

* Inference job submission API
* Asynchronous queueing and retry control
* Selection of target inference nodes
* Streaming delivery of inference results
* Persistence of job state
* Abstraction of the fixed pool of GPU nodes

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
  API Gateway, Router, and job management

* Python
  LLM inference processing

---

### Communication

* HTTP
  External communication with clients

* WebSocket
  Real-time delivery to clients

* gRPC
  Service-to-service communication with low latency and streaming support

---

### Inference Backend

* Mesh LLM
  Distributed inference execution across multiple GPU nodes

---

### Asynchronous Processing

* NATS JetStream
  Queueing, persistence, and retry control for inference jobs

---

### Data Management

* PostgreSQL
  Persistent storage for job state and history

* Redis
  Cache, shared state, and lightweight metrics management

---

### Infrastructure

* Kubernetes
  Execution and scaling platform for the orchestration layer

* Current state
  Fixed GPU server pool in a local environment

* Future state
  Kubernetes-based GPU node management and scaling

---

## Expected Use Cases

* Multiple users within a lab submit inference jobs concurrently
* The system handles not only short single-shot inference, but also long-running generation jobs
* Users create jobs through an HTTP API and receive progress or partial results over WebSocket
* Operators manage the state of fixed GPU nodes through the Router and NodeRegistry

---

## Design

### Overall Architecture

```text
Client
  ↓ (HTTP / WebSocket)
API Gateway (Kubernetes)
  ↓
Job / Queue Layer (JetStream / Kubernetes)
  ↓
Router (Kubernetes)
  ↓
Mesh LLM (GPU nodes)
  ↓ (gRPC / streaming)
Streaming / Result
  ↓
Client
```

---

### Layer Structure

#### 1. Orchestration Layer (running on Kubernetes)

Components:

* API Gateway
* Queue / Job Management (JetStream)
* Router
* Streaming control

Responsibilities:

* Accept requests and return responses
* Manage asynchronous processing
* Select target inference nodes
* Control the end-user experience

Characteristics:

* Runs as Pods on Kubernetes
* Supports horizontal scaling by increasing or decreasing Pods
* Assumes a stateless design

---

#### 2. Execution Layer

Components:

* Mesh LLM

Responsibilities:

* Execute distributed inference
* Handle inter-node communication
* Run models

Characteristics:

* Currently runs on fixed GPU nodes
* Can later move to Kubernetes-managed worker Pods

---

#### 3. Infrastructure Layer

Components:

* GPU nodes
* Kubernetes cluster

Responsibilities:

* Provide compute resources
* Manage Pod placement and scaling

---

### Component Responsibilities

| Component | Responsibility |
| --- | --- |
| API Gateway | External interface, authentication, and request intake |
| Queue (JetStream) | Asynchronous processing, persistence, and retry control |
| Router | Selection of inference nodes and models |
| Mesh LLM | Execution of distributed inference |
| Worker | GPU-side inference processing |

---

### Where Kubernetes Is Used

#### 1. Scaling the Orchestration Layer

* Deploy API Gateway, Router, Queue, and related services as Pods
* Scale horizontally using mechanisms such as HPA

---

#### 2. Availability

* Automatic Pod restarts
* Rolling updates

---

#### 3. Future GPU Scaling

* Adding GPU nodes
* Scheduling worker Pods
* Scaling nodes through Cluster Autoscaler

---

### Design Principles

#### 1. Wrap Mesh LLM

Mesh LLM is not exposed directly to external users; it is wrapped by the orchestration layer.

---

#### 2. Separate GPU Resource Management

GPU nodes are abstracted as compute providers.

* Current: fixed nodes
* Future: Kubernetes

---

#### 3. Abstract Node Discovery Through NodeRegistry

Node information is obtained through an abstract interface.

---

#### 4. Async-First Design

All inference requests go through JetStream.

---

#### 5. Separation of Responsibilities

Each component has a single responsibility.

---

#### 6. Design for Scalability from the Start

* The orchestration layer scales on Kubernetes
* GPU workers can later scale on Kubernetes

---

### Current State vs. Future State

| Item | Current | Future |
| --- | --- | --- |
| Orchestration Layer | Kubernetes | Kubernetes |
| GPU node management | Fixed | Kubernetes |
| Node discovery | Static | Dynamic (API / Service Discovery) |
| CPU-side scaling | Kubernetes | Kubernetes |
| GPU-side scaling | None | Kubernetes |

---

## Summary

This system uses Mesh LLM as the inference engine and extends it into an operationally viable platform for distributed LLM inference by introducing an orchestration layer running on Kubernetes.

From the outset, the orchestration layer is intended to run in a scalable way on Kubernetes, while GPU resources start as fixed nodes and remain structured so they can later transition to Kubernetes-based dynamic scaling.
